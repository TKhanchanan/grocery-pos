package httpx

import (
	"context"
	"net/http"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type AppSettings struct {
	ShopName          string `json:"shop_name"`
	ShopPhone         string `json:"shop_phone"`
	ShopAddress       string `json:"shop_address"`
	DefaultLocationID uint64 `json:"default_location_id"`
	ReceiptFooter     string `json:"receipt_footer"`
	LineEnabled       bool   `json:"line_enabled"`
	LineTokenMasked   string `json:"line_token_masked"`
	LineConfigured    bool   `json:"line_configured"`
	LineTargetID      string `json:"line_target_id"`
}

type ReceiptSettings struct {
	ShopName          string `json:"shop_name"`
	ShopPhone         string `json:"shop_phone"`
	ShopAddress       string `json:"shop_address"`
	DefaultLocationID uint64 `json:"default_location_id"`
	ReceiptFooter     string `json:"receipt_footer"`
}

type LineSettings struct {
	LineEnabled     bool   `json:"line_enabled"`
	LineToken       string `json:"line_token,omitempty"`
	LineTokenMasked string `json:"line_token_masked"`
	LineConfigured  bool   `json:"line_configured"`
	LineTargetID    string `json:"line_target_id"`
}

type NotificationLog struct {
	ID           uint64     `json:"id"`
	Channel      string     `json:"channel"`
	Recipient    string     `json:"recipient"`
	EventType    string     `json:"event_type"`
	Payload      string     `json:"payload"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message"`
	SentAt       *time.Time `json:"sent_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.loadAppSettings(r.Context(), false)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "SETTINGS_LOAD_FAILED", "Could not load settings.")
			return
		}
		response.JSON(w, http.StatusOK, settings)
	case http.MethodPatch:
		var body AppSettings
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if err := s.saveSettings(r.Context(), map[string]string{
			"shop_name":           strings.TrimSpace(body.ShopName),
			"shop_phone":          strings.TrimSpace(body.ShopPhone),
			"shop_address":        strings.TrimSpace(body.ShopAddress),
			"default_location_id": uintToString(body.DefaultLocationID),
			"receipt_footer":      strings.TrimSpace(body.ReceiptFooter),
		}, "shop"); err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "Could not save settings.")
			return
		}
		settings, _ := s.loadAppSettings(r.Context(), false)
		response.JSON(w, http.StatusOK, settings)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) receiptSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.loadReceiptSettings(r.Context())
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "SETTINGS_LOAD_FAILED", "Could not load receipt settings.")
		return
	}
	response.JSON(w, http.StatusOK, settings)
}

func (s *Server) lineSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.loadLineSettings(r.Context(), false)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "LINE_SETTINGS_LOAD_FAILED", "Could not load LINE settings.")
			return
		}
		response.JSON(w, http.StatusOK, settings)
	case http.MethodPatch:
		var body LineSettings
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		values := map[string]string{
			"line_enabled":   boolToString(body.LineEnabled),
			"line_target_id": strings.TrimSpace(body.LineTargetID),
		}
		token := strings.TrimSpace(body.LineToken)
		if token != "" && !strings.Contains(token, "*") {
			values["line_token"] = token
		}
		if err := s.saveSettings(r.Context(), values, "line"); err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "LINE_SETTINGS_SAVE_FAILED", "Could not save LINE settings.")
			return
		}
		settings, _ := s.loadLineSettings(r.Context(), false)
		response.JSON(w, http.StatusOK, settings)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) testLineNotification(w http.ResponseWriter, r *http.Request) {
	log, err := s.sendLineNotification(r.Context(), "LINE_TEST", "Grocery POS LINE notification test", map[string]any{"event": "LINE_TEST"}, true)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "LINE_TEST_FAILED", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, log)
}

func (s *Server) notificationLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, channel, recipient, event_type, COALESCE(CAST(payload AS CHAR), ''), status, error_message, sent_at, created_at
		FROM notification_logs
		ORDER BY id DESC
		LIMIT 300`)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load notification logs.")
		return
	}
	defer rows.Close()
	logs := []NotificationLog{}
	for rows.Next() {
		var log NotificationLog
		if err := rows.Scan(&log.ID, &log.Channel, &log.Recipient, &log.EventType, &log.Payload, &log.Status, &log.ErrorMessage, &log.SentAt, &log.CreatedAt); err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "SCAN_FAILED", "Could not read notification logs.")
			return
		}
		logs = append(logs, log)
	}
	response.JSON(w, http.StatusOK, logs)
}

func (s *Server) loadAppSettings(ctx context.Context, includeToken bool) (AppSettings, error) {
	values, err := s.settingsMap(ctx)
	if err != nil {
		return AppSettings{}, err
	}
	token := values["line_token"]
	return AppSettings{
		ShopName:          valueOr(values, "shop_name", "Grocery POS"),
		ShopPhone:         values["shop_phone"],
		ShopAddress:       values["shop_address"],
		DefaultLocationID: parseSettingUint(values["default_location_id"]),
		ReceiptFooter:     valueOr(values, "receipt_footer", "ขอบคุณที่อุดหนุน"),
		LineEnabled:       parseSettingBool(valueOr(values, "line_enabled", values["line_notifications_enabled"])),
		LineTokenMasked:   maskToken(token),
		LineConfigured:    token != "" && values["line_target_id"] != "",
		LineTargetID:      values["line_target_id"],
	}, nil
}

func (s *Server) loadReceiptSettings(ctx context.Context) (ReceiptSettings, error) {
	values, err := s.settingsMap(ctx)
	if err != nil {
		return ReceiptSettings{}, err
	}
	return ReceiptSettings{
		ShopName:          valueOr(values, "shop_name", "Grocery POS"),
		ShopPhone:         values["shop_phone"],
		ShopAddress:       values["shop_address"],
		DefaultLocationID: parseSettingUint(values["default_location_id"]),
		ReceiptFooter:     valueOr(values, "receipt_footer", "ขอบคุณที่อุดหนุน"),
	}, nil
}

func (s *Server) loadLineSettings(ctx context.Context, includeToken bool) (LineSettings, error) {
	values, err := s.settingsMap(ctx)
	if err != nil {
		return LineSettings{}, err
	}
	token := values["line_token"]
	settings := LineSettings{
		LineEnabled:     parseSettingBool(valueOr(values, "line_enabled", values["line_notifications_enabled"])),
		LineTokenMasked: maskToken(token),
		LineConfigured:  token != "" && values["line_target_id"] != "",
		LineTargetID:    values["line_target_id"],
	}
	if includeToken {
		settings.LineToken = token
	}
	return settings, nil
}

func (s *Server) settingsMap(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT setting_key, setting_value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, rows.Err()
}

func (s *Server) saveSettings(ctx context.Context, values map[string]string, group string) error {
	for key, value := range values {
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO settings(setting_key, setting_value, setting_group)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value), setting_group=VALUES(setting_group)`, key, value, group); err != nil {
			return err
		}
	}
	return nil
}
