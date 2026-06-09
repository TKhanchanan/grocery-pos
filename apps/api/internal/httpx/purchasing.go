package httpx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type Supplier struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	Active    bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type SupplierInput struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type PurchaseOrderInput struct {
	SupplierID uint64              `json:"supplier_id"`
	LocationID uint64              `json:"location_id"`
	Note       string              `json:"note"`
	Items      []PurchaseOrderItem `json:"items"`
}

type PurchaseOrder struct {
	ID           uint64              `json:"id"`
	PONumber     string              `json:"po_number"`
	SupplierID   uint64              `json:"supplier_id"`
	SupplierName string              `json:"supplier_name"`
	LocationID   uint64              `json:"location_id"`
	LocationName string              `json:"location_name"`
	Status       string              `json:"status"`
	TotalCost    float64             `json:"total_cost"`
	Note         string              `json:"note"`
	CreatedBy    *uint64             `json:"created_by"`
	ReceivedBy   *uint64             `json:"received_by"`
	CancelledBy  *uint64             `json:"cancelled_by"`
	ReceivedAt   *time.Time          `json:"received_at"`
	CancelledAt  *time.Time          `json:"cancelled_at"`
	CreatedAt    time.Time           `json:"created_at"`
	Items        []PurchaseOrderItem `json:"items"`
}

type PurchaseOrderItem struct {
	ID               uint64     `json:"id,omitempty"`
	POID             uint64     `json:"po_id,omitempty"`
	ProductID        uint64     `json:"product_id"`
	ProductName      string     `json:"product_name,omitempty"`
	SKU              string     `json:"sku,omitempty"`
	ImageURL         *string    `json:"image_url,omitempty"`
	ImageUpdated     *time.Time `json:"image_updated_at,omitempty"`
	Quantity         int        `json:"quantity"`
	ReceivedQuantity int        `json:"received_quantity"`
	UnitCost         float64    `json:"unit_cost"`
	LineCost         float64    `json:"line_cost"`
}

func (s *Server) suppliers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.listSuppliers(r.Context(), r)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load suppliers.")
			return
		}
		response.JSON(w, http.StatusOK, items)
	case http.MethodPost:
		var body SupplierInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		item, err := s.createSupplier(r.Context(), body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "SUPPLIER_VALIDATION_FAILED", err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, item)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) supplierDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid supplier id.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := s.supplierByID(r.Context(), id)
		if err != nil {
			response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Supplier not found.")
			return
		}
		response.JSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var body SupplierInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		item, err := s.updateSupplier(r.Context(), id, body)
		if err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "SUPPLIER_VALIDATION_FAILED", err.Error())
			return
		}
		response.JSON(w, http.StatusOK, item)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) supplierStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid supplier id.")
		return
	}
	var body struct {
		Active bool `json:"is_active"`
	}
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `UPDATE suppliers SET active=? WHERE id=?`, body.Active, id); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "UPDATE_FAILED", "Could not update supplier.")
		return
	}
	item, err := s.supplierByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Supplier not found.")
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (s *Server) purchaseOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.listPurchaseOrders(r.Context(), r)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load purchase orders.")
			return
		}
		response.JSON(w, http.StatusOK, items)
	case http.MethodPost:
		var body PurchaseOrderInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		user, _ := currentUser(r.Context())
		po, err := s.createPurchaseOrder(r.Context(), user, body)
		if err != nil {
			response.ErrorJSON(w, purchaseErrorStatus(err), purchaseErrorCode(err), err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, po)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) purchaseOrderDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid purchase order id.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		po, err := s.purchaseOrderByID(r.Context(), id)
		if err != nil {
			response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Purchase order not found.")
			return
		}
		response.JSON(w, http.StatusOK, po)
	case http.MethodPatch:
		var body PurchaseOrderInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		po, err := s.updatePurchaseOrder(r.Context(), id, body)
		if err != nil {
			response.ErrorJSON(w, purchaseErrorStatus(err), purchaseErrorCode(err), err.Error())
			return
		}
		response.JSON(w, http.StatusOK, po)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) sendPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	s.transitionPurchaseOrder(w, r, "SENT")
}

func (s *Server) cancelPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	s.transitionPurchaseOrder(w, r, "CANCELLED")
}

func (s *Server) receivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid purchase order id.")
		return
	}
	user, _ := currentUser(r.Context())
	po, err := s.receivePO(r.Context(), user, id)
	if err != nil {
		response.ErrorJSON(w, purchaseErrorStatus(err), purchaseErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusOK, po)
}

func (s *Server) transitionPurchaseOrder(w http.ResponseWriter, r *http.Request, target string) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid purchase order id.")
		return
	}
	user, _ := currentUser(r.Context())
	po, err := s.transitionPO(r.Context(), user, id, target)
	if err != nil {
		response.ErrorJSON(w, purchaseErrorStatus(err), purchaseErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusOK, po)
}

func (s *Server) listSuppliers(ctx context.Context, r *http.Request) ([]Supplier, error) {
	where := []string{"1=1"}
	args := []any{}
	query := r.URL.Query()
	if value := strings.TrimSpace(query.Get("search")); value != "" {
		like := "%" + value + "%"
		where = append(where, "(name LIKE ? OR phone LIKE ? OR email LIKE ? OR address LIKE ?)")
		args = append(args, like, like, like, like)
	}
	if value := strings.TrimSpace(query.Get("status")); value == "active" {
		where = append(where, "active=TRUE")
	} else if value == "inactive" {
		where = append(where, "active=FALSE")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, phone, email, address, active, created_at
		FROM suppliers
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Supplier{}
	for rows.Next() {
		var item Supplier
		if err := rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Email, &item.Address, &item.Active, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Server) createSupplier(ctx context.Context, body SupplierInput) (Supplier, error) {
	if strings.TrimSpace(body.Name) == "" {
		return Supplier{}, errors.New("supplier name is required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO suppliers(name, phone, email, address, active) VALUES (?, ?, ?, ?, TRUE)`, body.Name, body.Phone, body.Email, body.Address)
	if err != nil {
		return Supplier{}, err
	}
	id, _ := result.LastInsertId()
	return s.supplierByID(ctx, uint64(id))
}

func (s *Server) updateSupplier(ctx context.Context, id uint64, body SupplierInput) (Supplier, error) {
	if strings.TrimSpace(body.Name) == "" {
		return Supplier{}, errors.New("supplier name is required")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE suppliers SET name=?, phone=?, email=?, address=? WHERE id=?`, body.Name, body.Phone, body.Email, body.Address, id); err != nil {
		return Supplier{}, err
	}
	return s.supplierByID(ctx, id)
}

func (s *Server) supplierByID(ctx context.Context, id uint64) (Supplier, error) {
	var item Supplier
	err := s.db.QueryRowContext(ctx, `SELECT id, name, phone, email, address, active, created_at FROM suppliers WHERE id=?`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Email, &item.Address, &item.Active, &item.CreatedAt)
	return item, err
}

func validatePOInput(body PurchaseOrderInput) error {
	if body.SupplierID == 0 || body.LocationID == 0 {
		return errors.New("supplier and location are required")
	}
	if len(body.Items) == 0 {
		return errors.New("purchase order requires at least one item")
	}
	for _, item := range body.Items {
		if item.ProductID == 0 || item.Quantity <= 0 || item.UnitCost < 0 {
			return errors.New("purchase order items require product, positive quantity, and non-negative unit cost")
		}
	}
	return nil
}

func (s *Server) createPurchaseOrder(ctx context.Context, user User, body PurchaseOrderInput) (PurchaseOrder, error) {
	if err := validatePOInput(body); err != nil {
		return PurchaseOrder{}, err
	}
	var poID uint64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		total := poTotal(body.Items)
		now := time.Now()
		poNumber := fmt.Sprintf("PO%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
		result, err := tx.ExecContext(ctx, `
			INSERT INTO purchase_orders(po_number, supplier_id, location_id, status, total_cost, note, created_by)
			VALUES (?, ?, ?, 'DRAFT', ?, ?, ?)`, poNumber, body.SupplierID, body.LocationID, total, body.Note, user.ID)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		poID = uint64(id)
		return replacePOItems(ctx, tx, poID, body.Items)
	})
	if err != nil {
		return PurchaseOrder{}, err
	}
	po, err := s.purchaseOrderByID(ctx, poID)
	if err != nil {
		return PurchaseOrder{}, err
	}
	s.notifyEvent(ctx, "PURCHASE_ORDER_CREATED", "Purchase order created "+po.PONumber+" for "+po.SupplierName, map[string]any{
		"purchase_order_id": po.ID,
		"po_number":         po.PONumber,
		"supplier_id":       po.SupplierID,
		"location_id":       po.LocationID,
		"total_cost":        po.TotalCost,
	})
	return po, nil
}

func (s *Server) updatePurchaseOrder(ctx context.Context, poID uint64, body PurchaseOrderInput) (PurchaseOrder, error) {
	if err := validatePOInput(body); err != nil {
		return PurchaseOrder{}, err
	}
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM purchase_orders WHERE id=? FOR UPDATE`, poID).Scan(&status); err != nil {
			return err
		}
		if status != "DRAFT" {
			return errors.New("only draft purchase orders can be edited")
		}
		total := poTotal(body.Items)
		if _, err := tx.ExecContext(ctx, `UPDATE purchase_orders SET supplier_id=?, location_id=?, total_cost=?, note=? WHERE id=?`, body.SupplierID, body.LocationID, total, body.Note, poID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM purchase_order_items WHERE po_id=?`, poID); err != nil {
			return err
		}
		return replacePOItems(ctx, tx, poID, body.Items)
	})
	if err != nil {
		return PurchaseOrder{}, err
	}
	return s.purchaseOrderByID(ctx, poID)
}

func replacePOItems(ctx context.Context, tx *sql.Tx, poID uint64, items []PurchaseOrderItem) error {
	for _, item := range items {
		line := money(float64(item.Quantity) * item.UnitCost)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO purchase_order_items(po_id, product_id, quantity, received_quantity, unit_cost, line_cost)
			VALUES (?, ?, ?, 0, ?, ?)`, poID, item.ProductID, item.Quantity, item.UnitCost, line); err != nil {
			return err
		}
	}
	return nil
}

func poTotal(items []PurchaseOrderItem) float64 {
	var total float64
	for _, item := range items {
		total = money(total + float64(item.Quantity)*item.UnitCost)
	}
	return total
}

func (s *Server) transitionPO(ctx context.Context, user User, poID uint64, target string) (PurchaseOrder, error) {
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM purchase_orders WHERE id=? FOR UPDATE`, poID).Scan(&status); err != nil {
			return err
		}
		switch target {
		case "SENT":
			if status != "DRAFT" {
				return errors.New("only draft purchase orders can be sent")
			}
			_, err := tx.ExecContext(ctx, `UPDATE purchase_orders SET status='SENT' WHERE id=?`, poID)
			return err
		case "CANCELLED":
			if status == "RECEIVED" {
				return errors.New("received purchase orders cannot be cancelled")
			}
			if status == "CANCELLED" {
				return errors.New("purchase order is already cancelled")
			}
			_, err := tx.ExecContext(ctx, `UPDATE purchase_orders SET status='CANCELLED', cancelled_by=?, cancelled_at=NOW() WHERE id=?`, user.ID, poID)
			return err
		}
		return errors.New("invalid purchase order transition")
	})
	if err != nil {
		return PurchaseOrder{}, err
	}
	return s.purchaseOrderByID(ctx, poID)
}

func (s *Server) receivePO(ctx context.Context, user User, poID uint64) (PurchaseOrder, error) {
	affectedProducts := []uint64{}
	var receivedLocationID uint64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var status string
		var locationID uint64
		if err := tx.QueryRowContext(ctx, `SELECT status, location_id FROM purchase_orders WHERE id=? FOR UPDATE`, poID).Scan(&status, &locationID); err != nil {
			return err
		}
		receivedLocationID = locationID
		if status != "SENT" && status != "DRAFT" {
			return errors.New("only draft or sent purchase orders can be received")
		}
		rows, err := tx.QueryContext(ctx, `SELECT id, product_id, quantity, unit_cost FROM purchase_order_items WHERE po_id=? ORDER BY id`, poID)
		if err != nil {
			return err
		}
		defer rows.Close()
		type line struct {
			id        uint64
			productID uint64
			quantity  int
			unitCost  float64
		}
		lines := []line{}
		for rows.Next() {
			var item line
			if err := rows.Scan(&item.id, &item.productID, &item.quantity, &item.unitCost); err != nil {
				return err
			}
			lines = append(lines, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(lines) == 0 {
			return errors.New("purchase order has no items")
		}
		for _, item := range lines {
			before, err := lockedStock(ctx, tx, item.productID, locationID)
			if err != nil {
				return err
			}
			after := before + item.quantity
			if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, after, item.productID, locationID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE purchase_order_items SET received_quantity=? WHERE id=?`, item.quantity, item.id); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE products SET cost=? WHERE id=?`, item.unitCost, item.productID); err != nil {
				return err
			}
			note := fmt.Sprintf("purchase order %d received", poID)
			if _, err := insertStockMovement(ctx, tx, item.productID, locationID, "PO_RECEIVE", item.quantity, before, after, &item.unitCost, note, user.ID); err != nil {
				return err
			}
			if err := recalculateAlerts(ctx, tx, item.productID, locationID); err != nil {
				return err
			}
			affectedProducts = append(affectedProducts, item.productID)
		}
		_, err = tx.ExecContext(ctx, `UPDATE purchase_orders SET status='RECEIVED', received_by=?, received_at=NOW() WHERE id=?`, user.ID, poID)
		return err
	})
	if err != nil {
		return PurchaseOrder{}, err
	}
	po, err := s.purchaseOrderByID(ctx, poID)
	if err != nil {
		return PurchaseOrder{}, err
	}
	s.notifyEvent(ctx, "PURCHASE_ORDER_RECEIVED", "Purchase order received "+po.PONumber+" at "+po.LocationName, map[string]any{
		"purchase_order_id": po.ID,
		"po_number":         po.PONumber,
		"supplier_id":       po.SupplierID,
		"location_id":       po.LocationID,
		"total_cost":        po.TotalCost,
	})
	for _, productID := range affectedProducts {
		s.notifyActiveStockAlerts(ctx, productID, receivedLocationID)
	}
	return po, nil
}

func (s *Server) listPurchaseOrders(ctx context.Context, r *http.Request) ([]PurchaseOrder, error) {
	where := []string{"1=1"}
	args := []any{}
	query := r.URL.Query()
	if value := strings.TrimSpace(query.Get("search")); value != "" {
		like := "%" + value + "%"
		where = append(where, `(po.po_number LIKE ? OR po.note LIKE ? OR EXISTS (
			SELECT 1
			FROM purchase_order_items search_poi
			JOIN products search_p ON search_p.id=search_poi.product_id
			WHERE search_poi.po_id=po.id AND (search_p.name LIKE ? OR search_p.sku LIKE ?)
		))`)
		args = append(args, like, like, like, like)
	}
	if value := strings.TrimSpace(query.Get("supplier_id")); value != "" {
		where = append(where, "po.supplier_id=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("location_id")); value != "" {
		where = append(where, "po.location_id=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("status")); value != "" {
		where = append(where, "po.status=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("date_from")); value != "" {
		where = append(where, "po.created_at >= ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("date_to")); value != "" {
		where = append(where, "po.created_at < DATE_ADD(?, INTERVAL 1 DAY)")
		args = append(args, value)
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT po.id, po.po_number, po.supplier_id, s.name, po.location_id, l.name, po.status,
		       po.total_cost, po.note, po.created_by, po.received_by, po.cancelled_by,
		       po.received_at, po.cancelled_at, po.created_at
		FROM purchase_orders po
		JOIN suppliers s ON s.id=po.supplier_id
		JOIN locations l ON l.id=po.location_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY po.id DESC
		LIMIT 200`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PurchaseOrder{}
	for rows.Next() {
		po, err := scanPurchaseOrder(rows)
		if err != nil {
			return nil, err
		}
		po.Items, _ = s.purchaseOrderItems(ctx, po.ID)
		items = append(items, po)
	}
	return items, rows.Err()
}

func (s *Server) purchaseOrderByID(ctx context.Context, id uint64) (PurchaseOrder, error) {
	po, err := scanPurchaseOrder(s.db.QueryRowContext(ctx, `
		SELECT po.id, po.po_number, po.supplier_id, s.name, po.location_id, l.name, po.status,
		       po.total_cost, po.note, po.created_by, po.received_by, po.cancelled_by,
		       po.received_at, po.cancelled_at, po.created_at
		FROM purchase_orders po
		JOIN suppliers s ON s.id=po.supplier_id
		JOIN locations l ON l.id=po.location_id
		WHERE po.id=?`, id))
	if err != nil {
		return PurchaseOrder{}, err
	}
	po.Items, err = s.purchaseOrderItems(ctx, id)
	return po, err
}

type purchaseOrderScanner interface {
	Scan(dest ...any) error
}

func scanPurchaseOrder(scanner purchaseOrderScanner) (PurchaseOrder, error) {
	var po PurchaseOrder
	err := scanner.Scan(&po.ID, &po.PONumber, &po.SupplierID, &po.SupplierName, &po.LocationID, &po.LocationName, &po.Status, &po.TotalCost, &po.Note, &po.CreatedBy, &po.ReceivedBy, &po.CancelledBy, &po.ReceivedAt, &po.CancelledAt, &po.CreatedAt)
	return po, err
}

func (s *Server) purchaseOrderItems(ctx context.Context, poID uint64) ([]PurchaseOrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT poi.id, poi.po_id, poi.product_id, p.name, p.sku, p.image_url, p.image_updated_at, poi.quantity,
		       poi.received_quantity, poi.unit_cost, poi.line_cost
		FROM purchase_order_items poi
		JOIN products p ON p.id=poi.product_id
		WHERE poi.po_id=?
		ORDER BY poi.id`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PurchaseOrderItem{}
	for rows.Next() {
		var item PurchaseOrderItem
		var imageURL sql.NullString
		var imageUpdated sql.NullTime
		if err := rows.Scan(&item.ID, &item.POID, &item.ProductID, &item.ProductName, &item.SKU, &imageURL, &imageUpdated, &item.Quantity, &item.ReceivedQuantity, &item.UnitCost, &item.LineCost); err != nil {
			return nil, err
		}
		if imageURL.Valid {
			item.ImageURL = &imageURL.String
		}
		if imageUpdated.Valid {
			item.ImageUpdated = &imageUpdated.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func purchaseErrorCode(err error) string {
	if strings.Contains(err.Error(), "cannot") || strings.Contains(err.Error(), "only ") || strings.Contains(err.Error(), "already") {
		return "PURCHASE_ORDER_STATUS_ERROR"
	}
	return "PURCHASE_ORDER_VALIDATION_FAILED"
}

func purchaseErrorStatus(err error) int {
	if strings.Contains(err.Error(), "cannot") || strings.Contains(err.Error(), "only ") || strings.Contains(err.Error(), "already") {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
