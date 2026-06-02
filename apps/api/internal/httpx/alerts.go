package httpx

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type Alert struct {
	ID           uint64     `json:"id"`
	ProductID    uint64     `json:"product_id"`
	ProductName  string     `json:"product_name"`
	SKU          string     `json:"sku"`
	LocationID   uint64     `json:"location_id"`
	LocationName string     `json:"location_name"`
	Type         string     `json:"type"`
	Message      string     `json:"message"`
	ReadBy       *uint64    `json:"read_by"`
	ReadAt       *time.Time `json:"read_at"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	CreatedAt    time.Time  `json:"created_at"`
	Links        AlertLinks `json:"links"`
}

type AlertLinks struct {
	Product       string `json:"product"`
	Restock       string `json:"restock"`
	PurchaseOrder string `json:"purchase_order"`
}

func (s *Server) alerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := s.listAlerts(r.Context(), r)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid ") {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load alerts.")
		return
	}
	response.JSON(w, http.StatusOK, alerts)
}

func (s *Server) readAlert(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid alert id.")
		return
	}
	user, _ := currentUser(r.Context())
	if _, err := s.db.ExecContext(r.Context(), `UPDATE alerts SET read_by=?, read_at=COALESCE(read_at, NOW()) WHERE id=?`, user.ID, id); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "UPDATE_FAILED", "Could not mark alert as read.")
		return
	}
	alert, err := s.alertByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Alert not found.")
		return
	}
	response.JSON(w, http.StatusOK, alert)
}

func (s *Server) readAllAlerts(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r.Context())
	if _, err := s.db.ExecContext(r.Context(), `UPDATE alerts SET read_by=?, read_at=NOW() WHERE read_at IS NULL AND resolved_at IS NULL`, user.ID); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "UPDATE_FAILED", "Could not mark alerts as read.")
		return
	}
	alerts, err := s.listAlerts(r.Context(), r)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load alerts.")
		return
	}
	response.JSON(w, http.StatusOK, alerts)
}

func (s *Server) listAlerts(ctx context.Context, r *http.Request) ([]Alert, error) {
	where := []string{"a.resolved_at IS NULL"}
	args := []any{}
	query := r.URL.Query()
	if query.Get("unread") == "true" || query.Get("unread") == "1" {
		where = append(where, "a.read_at IS NULL")
	}
	if value := strings.TrimSpace(query.Get("type")); value != "" {
		where = append(where, "a.type=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("location_id")); value != "" {
		id, err := parseUintParam(value, "location_id")
		if err != nil {
			return nil, err
		}
		where = append(where, "a.location_id=?")
		args = append(args, id)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, a.product_id, p.name, p.sku, a.location_id, l.name, a.type, a.message,
		       a.read_by, a.read_at, a.resolved_at, a.created_at
		FROM alerts a
		JOIN products p ON p.id=a.product_id
		JOIN locations l ON l.id=a.location_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY a.read_at IS NOT NULL, a.created_at DESC
		LIMIT 300`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	alerts := []Alert{}
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

func (s *Server) alertByID(ctx context.Context, id uint64) (Alert, error) {
	return scanAlert(s.db.QueryRowContext(ctx, `
		SELECT a.id, a.product_id, p.name, p.sku, a.location_id, l.name, a.type, a.message,
		       a.read_by, a.read_at, a.resolved_at, a.created_at
		FROM alerts a
		JOIN products p ON p.id=a.product_id
		JOIN locations l ON l.id=a.location_id
		WHERE a.id=?`, id))
}

type alertScanner interface {
	Scan(dest ...any) error
}

func scanAlert(scanner alertScanner) (Alert, error) {
	var alert Alert
	err := scanner.Scan(&alert.ID, &alert.ProductID, &alert.ProductName, &alert.SKU, &alert.LocationID, &alert.LocationName, &alert.Type, &alert.Message, &alert.ReadBy, &alert.ReadAt, &alert.ResolvedAt, &alert.CreatedAt)
	alert.Links = AlertLinks{
		Product:       "/products",
		Restock:       "/restock",
		PurchaseOrder: fmt.Sprintf("/purchase-orders?product_id=%d&location_id=%d", alert.ProductID, alert.LocationID),
	}
	return alert, err
}

func parseUintParam(value, name string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func recalculateAlerts(ctx context.Context, tx *sql.Tx, productID, locationID uint64) error {
	var productName, locationName string
	var quantity, threshold, reorderPoint int
	if err := tx.QueryRowContext(ctx, `
		SELECT p.name, l.name, ps.quantity, p.threshold, p.reorder_point
		FROM products p
		JOIN locations l ON l.id=?
		JOIN product_stocks ps ON ps.product_id=p.id AND ps.location_id=l.id
		WHERE p.id=?`, locationID, productID).Scan(&productName, &locationName, &quantity, &threshold, &reorderPoint); err != nil {
		return err
	}

	active := map[string]string{}
	if quantity == 0 {
		active["OUT_OF_STOCK"] = productName + " out of stock at " + locationName
	}
	if threshold > 0 && quantity <= threshold {
		active["LOW_STOCK"] = productName + " low stock at " + locationName + ": " + strconv.Itoa(quantity)
	}
	if reorderPoint > 0 && quantity <= reorderPoint {
		active["REORDER_POINT"] = productName + " reached reorder point at " + locationName + ": " + strconv.Itoa(quantity)
	}

	for _, alertType := range []string{"LOW_STOCK", "OUT_OF_STOCK", "REORDER_POINT"} {
		if _, ok := active[alertType]; ok {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE alerts SET resolved_at=NOW() WHERE product_id=? AND location_id=? AND type=? AND resolved_at IS NULL`, productID, locationID, alertType); err != nil {
			return err
		}
	}

	for alertType, message := range active {
		var exists int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM alerts
			WHERE product_id=? AND location_id=? AND type=? AND resolved_at IS NULL AND read_at IS NULL`, productID, locationID, alertType).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO alerts(product_id, location_id, type, message) VALUES (?, ?, ?, ?)`, productID, locationID, alertType, message); err != nil {
			return err
		}
	}
	return nil
}
