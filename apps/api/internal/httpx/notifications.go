package httpx

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const linePushURL = "https://api.line.me/v2/bot/message/push"

func (s *Server) notifyEvent(ctx context.Context, eventType, message string, payload map[string]any) {
	_, _ = s.sendLineNotification(ctx, eventType, message, payload, false)
}

func (s *Server) notifyActiveStockAlerts(ctx context.Context, productID, locationID uint64) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, a.type, a.message, p.name, l.name
		FROM alerts a
		JOIN products p ON p.id=a.product_id
		JOIN locations l ON l.id=a.location_id
		WHERE a.product_id=? AND a.location_id=? AND a.resolved_at IS NULL AND a.read_at IS NULL
		ORDER BY a.id DESC`, productID, locationID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var alertID uint64
		var alertType, message, productName, locationName string
		if err := rows.Scan(&alertID, &alertType, &message, &productName, &locationName); err != nil {
			return
		}
		s.notifyEvent(ctx, alertType, message, map[string]any{
			"alert_id":      alertID,
			"product_id":    productID,
			"product_name":  productName,
			"location_id":   locationID,
			"location_name": locationName,
		})
	}
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

	body, _ := json.Marshal(map[string]any{
		"to": settings.LineTargetID,
		"messages": []map[string]string{
			{"type": "text", "text": message},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linePushURL, bytes.NewReader(body))
	if err != nil {
		log, logErr := s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "FAILED", err.Error())
		if force && logErr == nil {
			return log, err
		}
		return log, logErr
	}
	req.Header.Set("Authorization", "Bearer "+settings.LineToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 6 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log, logErr := s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "FAILED", trimError(err.Error()))
		if force && logErr == nil {
			return log, err
		}
		return log, logErr
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		errMessage := trimError(fmt.Sprintf("LINE API %d: %s", res.StatusCode, strings.TrimSpace(string(bodyBytes))))
		log, logErr := s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "FAILED", errMessage)
		if force && logErr == nil {
			return log, errors.New(errMessage)
		}
		return log, logErr
	}
	return s.logNotification(ctx, "LINE", settings.LineTargetID, eventType, payloadWithMessage(payload, message), "SENT", "")
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
