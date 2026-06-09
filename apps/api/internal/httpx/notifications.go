package httpx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/line"
)

func (s *Server) notifyEvent(ctx context.Context, eventType, message string, payload map[string]any) {
	_, _ = s.sendLineNotification(ctx, eventType, message, payload, false)
}

func (s *Server) notifySaleCompleted(ctx context.Context, saleID, locationID uint64, receiptNo string, total float64, saleTime time.Time) {
	s.notifyEvent(ctx, "SALE_COMPLETED", "Sale completed "+receiptNo+" total "+strconv.FormatFloat(total, 'f', 2, 64), map[string]any{
		"sale_id":     saleID,
		"location_id": locationID,
		"receipt_no":  receiptNo,
		"total":       total,
		"sale_time":   saleTime,
	})
}

func (s *Server) notifyActiveStockAlerts(ctx context.Context, productID, locationID uint64) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, a.type, a.message, p.name, l.name, ps.quantity, p.reorder_point
		FROM alerts a
		JOIN products p ON p.id=a.product_id
		JOIN locations l ON l.id=a.location_id
		JOIN product_stocks ps ON ps.product_id=p.id AND ps.location_id=l.id
		WHERE a.product_id=? AND a.location_id=? AND a.resolved_at IS NULL AND a.read_at IS NULL
		ORDER BY a.id DESC`, productID, locationID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var alertID uint64
		var alertType, message, productName, locationName string
		var quantity, reorderPoint int
		if err := rows.Scan(&alertID, &alertType, &message, &productName, &locationName, &quantity, &reorderPoint); err != nil {
			return
		}
		s.notifyStockAlert(ctx, alertType, message, line.StockAlertInput{
			ProductName:  productName,
			LocationName: locationName,
			Quantity:     quantity,
			ReorderPoint: reorderPoint,
		}, map[string]any{
			"alert_id":    alertID,
			"product_id":  productID,
			"location_id": locationID,
		})
	}
}

func (s *Server) notifyStockAlert(ctx context.Context, eventType, fallbackText string, input line.StockAlertInput, payload map[string]any) {
	payload["product_name"] = input.ProductName
	payload["location_name"] = input.LocationName
	payload["quantity"] = input.Quantity
	payload["reorder_point"] = input.ReorderPoint
	s.notifyEvent(ctx, eventType, fallbackText, payload)
}

func (s *Server) sendTestLineNotification(ctx context.Context) (*NotificationLog, error) {
	return s.sendLineNotification(ctx, "LINE_TEST", "GroceryPOS LINE notification test", map[string]any{
		"event":     "LINE_TEST",
		"timestamp": time.Now(),
	}, true)
}

func (s *Server) sendLineNotification(ctx context.Context, eventType, message string, payload map[string]any, force bool) (*NotificationLog, error) {
	settings, err := s.loadLineSettings(ctx, true)
	if err != nil {
		if force {
			return nil, err
		}
		return nil, nil
	}
	if !settings.LineEnabled && !force {
		return nil, nil
	}
	if !settings.LineEnabled && force {
		log, logErr := s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "FAILED", "LINE notifications are disabled")
		if logErr != nil {
			return log, logErr
		}
		return log, errors.New("LINE notifications are disabled")
	}
	if strings.TrimSpace(settings.LineToken) == "" || strings.TrimSpace(settings.LineTargetID) == "" {
		log, logErr := s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "FAILED", "LINE token and target id are required")
		if force {
			if logErr != nil {
				return log, logErr
			}
			return log, errors.New("LINE token and target id are required")
		}
		return log, nil
	}

	lineMessage := buildLineMessage(eventType, payload)
	if lineMessage == nil {
		lineMessage = line.NewTextMessage(message)
	}
	err = pushLineMessage(ctx, s.lineClient, settings.LineToken, settings.LineTargetID, eventType, message, lineMessage)
	if err != nil {
		notificationLog, logErr := s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "FAILED", trimError(err.Error()))
		if force && logErr == nil {
			return notificationLog, err
		}
		return notificationLog, logErr
	}
	return s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "SENT", "")
}

func pushLineMessage(ctx context.Context, client linePusher, token, targetID, eventType, fallbackText string, message line.Message) error {
	err := client.Push(ctx, token, targetID, message)
	if err == nil {
		return nil
	}
	if _, isFlex := message.(line.FlexMessage); !isFlex {
		return err
	}
	log.Printf("LINE Flex notification %s failed, retrying as text: %v", eventType, err)
	return client.Push(ctx, token, targetID, line.NewTextMessage(fallbackText))
}

func buildLineMessage(eventType string, payload map[string]any) line.Message {
	switch eventType {
	case "SALE_COMPLETED":
		return line.BuildSaleCompleted(line.SaleCompletedInput{
			ReceiptNo: stringPayload(payload, "receipt_no"),
			Total:     floatPayload(payload, "total"),
			SaleTime:  timePayload(payload, "sale_time"),
		})
	case "LOW_STOCK":
		return line.BuildLowStock(stockAlertPayload(payload))
	case "OUT_OF_STOCK":
		return line.BuildOutOfStock(stockAlertPayload(payload))
	case "REORDER_POINT":
		return line.BuildReorderPoint(stockAlertPayload(payload))
	case "LINE_TEST":
		timestamp := timePayload(payload, "timestamp")
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		return line.BuildTestNotification(line.TestNotificationInput{Timestamp: timestamp})
	default:
		return nil
	}
}

func stockAlertPayload(payload map[string]any) line.StockAlertInput {
	return line.StockAlertInput{
		ProductName:  stringPayload(payload, "product_name"),
		LocationName: stringPayload(payload, "location_name"),
		Quantity:     intPayload(payload, "quantity"),
		ReorderPoint: intPayload(payload, "reorder_point"),
	}
}

func stringPayload(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func intPayload(payload map[string]any, key string) int {
	switch value := payload[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case uint64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func floatPayload(payload map[string]any, key string) float64 {
	switch value := payload[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	default:
		return 0
	}
}

func timePayload(payload map[string]any, key string) time.Time {
	switch value := payload[key].(type) {
	case time.Time:
		return value
	case string:
		parsed, _ := time.Parse(time.RFC3339, value)
		return parsed
	default:
		return time.Time{}
	}
}

func (s *Server) logNotification(ctx context.Context, channel, recipient, eventType string, payload map[string]any, status, errorMessage string) (*NotificationLog, error) {
	payloadBytes, _ := json.Marshal(payload)
	var sentAt any
	if status == "SENT" {
		sentAt = time.Now()
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_logs(channel, recipient, event_type, payload, status, error_message, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, channel, recipient, eventType, string(payloadBytes), status, trimError(errorMessage), sentAt)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	log, err := s.notificationLogByID(ctx, uint64(id))
	return &log, err
}

func (s *Server) notificationLogByID(ctx context.Context, id uint64) (NotificationLog, error) {
	var log NotificationLog
	err := s.db.QueryRowContext(ctx, `
		SELECT id, channel, recipient, event_type, COALESCE(CAST(payload AS CHAR), ''), status, error_message, sent_at, created_at
		FROM notification_logs
		WHERE id=?`, id).Scan(&log.ID, &log.Channel, &log.Recipient, &log.EventType, &log.Payload, &log.Status, &log.ErrorMessage, &log.SentAt, &log.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return NotificationLog{}, err
	}
	return log, err
}

func payloadWithMessage(payload map[string]any, message string) map[string]any {
	out := map[string]any{"message": message}
	for key, value := range payload {
		out[key] = value
	}
	return out
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

func valueOr(values map[string]string, key, fallback string) string {
	if value := strings.TrimSpace(values[key]); value != "" {
		return value
	}
	return fallback
}

func parseSettingBool(value string) bool {
	return strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
}

func parseSettingUint(value string) uint64 {
	id, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return id
}

func boolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func uintToString(value uint64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatUint(value, 10)
}

func trimError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 255 {
		return value[:255]
	}
	return value
}
