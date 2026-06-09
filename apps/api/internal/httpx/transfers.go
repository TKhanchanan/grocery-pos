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

type StockTransferInput struct {
	FromLocationID uint64              `json:"from_location_id"`
	ToLocationID   uint64              `json:"to_location_id"`
	Note           string              `json:"note"`
	Items          []StockTransferItem `json:"items"`
}

type StockTransferItem struct {
	ID          uint64 `json:"id,omitempty"`
	TransferID  uint64 `json:"transfer_id,omitempty"`
	ProductID   uint64 `json:"product_id"`
	ProductName string `json:"product_name,omitempty"`
	SKU         string `json:"sku,omitempty"`
	Quantity    int    `json:"quantity"`
}

type StockTransfer struct {
	ID               uint64              `json:"id"`
	TransferNo       string              `json:"transfer_no"`
	FromLocationID   uint64              `json:"from_location_id"`
	FromLocationName string              `json:"from_location_name"`
	ToLocationID     uint64              `json:"to_location_id"`
	ToLocationName   string              `json:"to_location_name"`
	Status           string              `json:"status"`
	Note             string              `json:"note"`
	CreatedBy        *uint64             `json:"created_by"`
	CompletedAt      *time.Time          `json:"completed_at"`
	CancelledAt      *time.Time          `json:"cancelled_at"`
	CreatedAt        time.Time           `json:"created_at"`
	Items            []StockTransferItem `json:"items"`
}

type StockTransferPage struct {
	Items    []StockTransfer `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type StockTransferOptions struct {
	Products  []Product      `json:"products"`
	Locations []Location     `json:"locations"`
	Stocks    []ProductStock `json:"stocks"`
}

func (s *Server) stockTransferOptions(w http.ResponseWriter, r *http.Request) {
	products, err := s.listProducts(r.Context(), r)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load products.")
		return
	}
	locationRows, err := s.db.QueryContext(r.Context(), `SELECT id, name, description, active, created_at FROM locations ORDER BY name`)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load locations.")
		return
	}
	defer locationRows.Close()
	locations := []Location{}
	for locationRows.Next() {
		var item Location
		if err := locationRows.Scan(&item.ID, &item.Name, &item.Description, &item.Active, &item.CreatedAt); err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "SCAN_FAILED", "Could not read locations.")
			return
		}
		locations = append(locations, item)
	}
	stocks, err := s.queryProductStocks(r.Context(), nil)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load product stocks.")
		return
	}
	response.JSON(w, http.StatusOK, StockTransferOptions{
		Products:  products,
		Locations: locations,
		Stocks:    stocks,
	})
}

func (s *Server) stockTransfers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page := positiveQueryInt(r, "page", 1)
		pageSize := positiveQueryInt(r, "page_size", 20)
		if pageSize != 10 && pageSize != 20 && pageSize != 50 {
			pageSize = 20
		}
		transfers, total, err := s.listStockTransfers(r.Context(), r, page, pageSize)
		if err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load stock transfers.")
			return
		}
		response.JSON(w, http.StatusOK, StockTransferPage{
			Items:    transfers,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	case http.MethodPost:
		var body StockTransferInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		user, _ := currentUser(r.Context())
		transfer, err := s.createStockTransfer(r.Context(), user, body)
		if err != nil {
			response.ErrorJSON(w, transferErrorStatus(err), transferErrorCode(err), err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, transfer)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) stockTransferDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid transfer id.")
		return
	}
	transfer, err := s.stockTransferByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Transfer not found.")
		return
	}
	response.JSON(w, http.StatusOK, transfer)
}

func (s *Server) completeStockTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid transfer id.")
		return
	}
	user, _ := currentUser(r.Context())
	transfer, err := s.completeTransfer(r.Context(), user, id)
	if err != nil {
		response.ErrorJSON(w, transferErrorStatus(err), transferErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusOK, transfer)
}

func (s *Server) cancelStockTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid transfer id.")
		return
	}
	transfer, err := s.cancelTransfer(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, transferErrorStatus(err), transferErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusOK, transfer)
}

func (s *Server) listStockTransfers(ctx context.Context, r *http.Request, page, pageSize int) ([]StockTransfer, int, error) {
	where := []string{"1=1"}
	args := []any{}
	query := r.URL.Query()
	if value := strings.TrimSpace(query.Get("search")); value != "" {
		like := "%" + value + "%"
		where = append(where, `(st.transfer_no LIKE ? OR st.note LIKE ? OR EXISTS (
			SELECT 1
			FROM stock_transfer_items search_sti
			JOIN products search_p ON search_p.id=search_sti.product_id
			WHERE search_sti.transfer_id=st.id AND (search_p.name LIKE ? OR search_p.sku LIKE ?)
		))`)
		args = append(args, like, like, like, like)
	}
	if value := strings.TrimSpace(query.Get("product_id")); value != "" {
		where = append(where, `EXISTS (
			SELECT 1 FROM stock_transfer_items product_sti
			WHERE product_sti.transfer_id=st.id AND product_sti.product_id=?
		)`)
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("from_location_id")); value != "" {
		where = append(where, "st.from_location_id=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("to_location_id")); value != "" {
		where = append(where, "st.to_location_id=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("date_from")); value != "" {
		where = append(where, "st.created_at >= ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("date_to")); value != "" {
		where = append(where, "st.created_at < DATE_ADD(?, INTERVAL 1 DAY)")
		args = append(args, value)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM stock_transfers st
		WHERE `+strings.Join(where, " AND "), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx, `
		SELECT st.id, st.transfer_no, st.from_location_id, fl.name, st.to_location_id, tl.name,
		       st.status, st.note, st.created_by, st.completed_at, st.cancelled_at, st.created_at
		FROM stock_transfers st
		JOIN locations fl ON fl.id=st.from_location_id
		JOIN locations tl ON tl.id=st.to_location_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY st.id DESC
		LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	transfers := []StockTransfer{}
	for rows.Next() {
		transfer, err := scanStockTransfer(rows)
		if err != nil {
			return nil, 0, err
		}
		transfer.Items, _ = s.stockTransferItems(ctx, transfer.ID)
		transfers = append(transfers, transfer)
	}
	return transfers, total, rows.Err()
}

func (s *Server) createStockTransfer(ctx context.Context, user User, body StockTransferInput) (StockTransfer, error) {
	if err := validateTransferInput(body); err != nil {
		return StockTransfer{}, err
	}

	var transferID uint64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		transferNo := fmt.Sprintf("TR%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
		result, err := tx.ExecContext(ctx, `
			INSERT INTO stock_transfers(transfer_no, from_location_id, to_location_id, status, note, created_by)
			VALUES (?, ?, ?, 'DRAFT', ?, ?)`, transferNo, body.FromLocationID, body.ToLocationID, body.Note, user.ID)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		transferID = uint64(id)
		for _, item := range body.Items {
			if item.Quantity <= 0 {
				return errors.New("transfer item quantity must be greater than 0")
			}
			sourceBefore, err := lockedStock(ctx, tx, item.ProductID, body.FromLocationID)
			if err != nil {
				return err
			}
			if sourceBefore < item.Quantity {
				return errors.New("insufficient source stock")
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO stock_transfer_items(transfer_id, product_id, quantity) VALUES (?, ?, ?)`, transferID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return StockTransfer{}, err
	}
	return s.stockTransferByID(ctx, transferID)
}

func (s *Server) completeTransfer(ctx context.Context, user User, transferID uint64) (StockTransfer, error) {
	type affectedStock struct {
		productID  uint64
		locationID uint64
	}
	affected := []affectedStock{}
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var status string
		var fromLocationID, toLocationID uint64
		if err := tx.QueryRowContext(ctx, `SELECT status, from_location_id, to_location_id FROM stock_transfers WHERE id=? FOR UPDATE`, transferID).Scan(&status, &fromLocationID, &toLocationID); err != nil {
			return err
		}
		if status != "DRAFT" {
			return errors.New("only draft transfers can be completed")
		}
		rows, err := tx.QueryContext(ctx, `SELECT product_id, quantity FROM stock_transfer_items WHERE transfer_id=?`, transferID)
		if err != nil {
			return err
		}
		defer rows.Close()
		type line struct {
			productID uint64
			quantity  int
		}
		lines := []line{}
		for rows.Next() {
			var item line
			if err := rows.Scan(&item.productID, &item.quantity); err != nil {
				return err
			}
			lines = append(lines, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(lines) == 0 {
			return errors.New("transfer requires at least one item")
		}
		for _, item := range lines {
			sourceBefore, err := lockedStock(ctx, tx, item.productID, fromLocationID)
			if err != nil {
				return err
			}
			if sourceBefore < item.quantity {
				return errors.New("insufficient source stock")
			}
			sourceAfter := sourceBefore - item.quantity
			if sourceAfter < 0 {
				return errors.New("stock cannot become negative")
			}
			destBefore, err := lockedStock(ctx, tx, item.productID, toLocationID)
			if err != nil {
				return err
			}
			destAfter := destBefore + item.quantity
			if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, sourceAfter, item.productID, fromLocationID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, destAfter, item.productID, toLocationID); err != nil {
				return err
			}
			if _, err := insertTransferMovement(ctx, tx, item.productID, fromLocationID, "TRANSFER_OUT", -item.quantity, sourceBefore, sourceAfter, transferID, user.ID); err != nil {
				return err
			}
			if _, err := insertTransferMovement(ctx, tx, item.productID, toLocationID, "TRANSFER_IN", item.quantity, destBefore, destAfter, transferID, user.ID); err != nil {
				return err
			}
			if err := recalculateAlerts(ctx, tx, item.productID, fromLocationID); err != nil {
				return err
			}
			if err := recalculateAlerts(ctx, tx, item.productID, toLocationID); err != nil {
				return err
			}
			affected = append(affected,
				affectedStock{productID: item.productID, locationID: fromLocationID},
				affectedStock{productID: item.productID, locationID: toLocationID},
			)
		}
		_, err = tx.ExecContext(ctx, `UPDATE stock_transfers SET status='COMPLETED', completed_at=NOW() WHERE id=?`, transferID)
		return err
	})
	if err != nil {
		return StockTransfer{}, err
	}
	transfer, err := s.stockTransferByID(ctx, transferID)
	if err != nil {
		return StockTransfer{}, err
	}
	s.notifyEvent(ctx, "TRANSFER_COMPLETED", "Transfer completed "+transfer.TransferNo+" from "+transfer.FromLocationName+" to "+transfer.ToLocationName, map[string]any{
		"transfer_id":      transfer.ID,
		"transfer_no":      transfer.TransferNo,
		"from_location_id": transfer.FromLocationID,
		"to_location_id":   transfer.ToLocationID,
	})
	for _, item := range affected {
		s.notifyActiveStockAlerts(ctx, item.productID, item.locationID)
	}
	return transfer, nil
}

func (s *Server) cancelTransfer(ctx context.Context, transferID uint64) (StockTransfer, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE stock_transfers SET status='CANCELLED', cancelled_at=NOW() WHERE id=? AND status='DRAFT'`, transferID)
	if err != nil {
		return StockTransfer{}, err
	}
	return s.stockTransferByID(ctx, transferID)
}

func (s *Server) stockTransferByID(ctx context.Context, id uint64) (StockTransfer, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT st.id, st.transfer_no, st.from_location_id, fl.name, st.to_location_id, tl.name,
		       st.status, st.note, st.created_by, st.completed_at, st.cancelled_at, st.created_at
		FROM stock_transfers st
		JOIN locations fl ON fl.id=st.from_location_id
		JOIN locations tl ON tl.id=st.to_location_id
		WHERE st.id=?`, id)
	transfer, err := scanStockTransfer(row)
	if err != nil {
		return StockTransfer{}, err
	}
	transfer.Items, err = s.stockTransferItems(ctx, id)
	return transfer, err
}

type stockTransferScanner interface {
	Scan(dest ...any) error
}

func scanStockTransfer(scanner stockTransferScanner) (StockTransfer, error) {
	var transfer StockTransfer
	err := scanner.Scan(&transfer.ID, &transfer.TransferNo, &transfer.FromLocationID, &transfer.FromLocationName, &transfer.ToLocationID, &transfer.ToLocationName, &transfer.Status, &transfer.Note, &transfer.CreatedBy, &transfer.CompletedAt, &transfer.CancelledAt, &transfer.CreatedAt)
	return transfer, err
}

func (s *Server) stockTransferItems(ctx context.Context, transferID uint64) ([]StockTransferItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT sti.id, sti.transfer_id, sti.product_id, p.name, p.sku, sti.quantity
		FROM stock_transfer_items sti
		JOIN products p ON p.id=sti.product_id
		WHERE sti.transfer_id=?
		ORDER BY sti.id`, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []StockTransferItem{}
	for rows.Next() {
		var item StockTransferItem
		if err := rows.Scan(&item.ID, &item.TransferID, &item.ProductID, &item.ProductName, &item.SKU, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func validateTransferInput(body StockTransferInput) error {
	if body.FromLocationID == 0 || body.ToLocationID == 0 {
		return errors.New("source and destination locations are required")
	}
	if body.FromLocationID == body.ToLocationID {
		return errors.New("source and destination locations must be different")
	}
	if len(body.Items) == 0 {
		return errors.New("transfer requires at least one item")
	}
	for _, item := range body.Items {
		if item.ProductID == 0 || item.Quantity <= 0 {
			return errors.New("transfer item product and positive quantity are required")
		}
	}
	return nil
}

func insertTransferMovement(ctx context.Context, tx *sql.Tx, productID, locationID uint64, movementType string, delta, before, after int, transferID, userID uint64) (uint64, error) {
	note := fmt.Sprintf("stock transfer %d", transferID)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO stock_movements(product_id, location_id, reference_type, reference_id, quantity_change, before_stock, after_stock, quantity_after, note, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		productID, locationID, movementType, transferID, delta, before, after, after, note, userID)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

func transferErrorCode(err error) string {
	if err.Error() == "insufficient source stock" {
		return "INSUFFICIENT_STOCK"
	}
	if err.Error() == "stock cannot become negative" {
		return "NEGATIVE_STOCK"
	}
	return "TRANSFER_VALIDATION_FAILED"
}

func transferErrorStatus(err error) int {
	if err.Error() == "insufficient source stock" || err.Error() == "stock cannot become negative" {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
