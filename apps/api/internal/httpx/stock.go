package httpx

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type RestockInput struct {
	LocationID uint64   `json:"location_id"`
	Quantity   int      `json:"quantity"`
	TotalCost  *float64 `json:"total_cost"`
	UnitCost   float64  `json:"unit_cost"`
	Note       string   `json:"note"`
}

type AdjustmentInput struct {
	LocationID uint64 `json:"location_id"`
	Quantity   int    `json:"quantity"`
	Note       string `json:"note"`
}

type StockMovement struct {
	ID             uint64    `json:"id"`
	ProductID      uint64    `json:"product_id"`
	ProductName    string    `json:"product_name"`
	SKU            string    `json:"sku"`
	LocationID     uint64    `json:"location_id"`
	LocationName   string    `json:"location_name"`
	ReferenceType  string    `json:"reference_type"`
	ReferenceID    *uint64   `json:"reference_id"`
	QuantityChange int       `json:"quantity_change"`
	BeforeStock    int       `json:"before_stock"`
	AfterStock     int       `json:"after_stock"`
	UnitCost       *float64  `json:"unit_cost"`
	Note           string    `json:"note"`
	CreatedBy      *uint64   `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

func (s *Server) restockProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid product id.")
		return
	}
	var body RestockInput
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	user, _ := currentUser(r.Context())
	movement, err := s.restock(r.Context(), user, productID, body)
	if err != nil {
		response.ErrorJSON(w, stockErrorStatus(err), stockErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, movement)
}

func (s *Server) adjustProductStock(w http.ResponseWriter, r *http.Request) {
	productID, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid product id.")
		return
	}
	var body AdjustmentInput
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	user, _ := currentUser(r.Context())
	movement, err := s.adjustStock(r.Context(), user, productID, body)
	if err != nil {
		response.ErrorJSON(w, stockErrorStatus(err), stockErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, movement)
}

func (s *Server) stockMovements(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT sm.id, sm.product_id, p.name, p.sku, sm.location_id, l.name, sm.reference_type, sm.reference_id,
		       sm.quantity_change, sm.before_stock, sm.after_stock, sm.unit_cost, sm.note, sm.created_by, sm.created_at
		FROM stock_movements sm
		JOIN products p ON p.id=sm.product_id
		JOIN locations l ON l.id=sm.location_id
		ORDER BY sm.id DESC
		LIMIT 300`)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load stock movements.")
		return
	}
	defer rows.Close()
	movements := []StockMovement{}
	for rows.Next() {
		var item StockMovement
		if err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.SKU, &item.LocationID, &item.LocationName, &item.ReferenceType, &item.ReferenceID, &item.QuantityChange, &item.BeforeStock, &item.AfterStock, &item.UnitCost, &item.Note, &item.CreatedBy, &item.CreatedAt); err != nil {
			response.ErrorJSON(w, http.StatusInternalServerError, "SCAN_FAILED", "Could not read stock movements.")
			return
		}
		movements = append(movements, item)
	}
	response.JSON(w, http.StatusOK, movements)
}

func (s *Server) restock(ctx context.Context, user User, productID uint64, body RestockInput) (StockMovement, error) {
	if body.Quantity <= 0 {
		return StockMovement{}, errors.New("quantity must be greater than 0")
	}
	unitCost := body.UnitCost
	if body.TotalCost != nil {
		if *body.TotalCost < 0 {
			return StockMovement{}, errors.New("total cost must be greater than or equal to 0")
		}
		unitCost = *body.TotalCost / float64(body.Quantity)
	}
	if unitCost < 0 {
		return StockMovement{}, errors.New("unit cost must be greater than or equal to 0")
	}

	var movementID uint64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		before, err := lockedStock(ctx, tx, productID, body.LocationID)
		if err != nil {
			return err
		}
		after := before + body.Quantity
		if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, after, productID, body.LocationID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE products SET cost=? WHERE id=?`, unitCost, productID); err != nil {
			return err
		}
		id, err := insertStockMovement(ctx, tx, productID, body.LocationID, "RESTOCK", body.Quantity, before, after, &unitCost, body.Note, user.ID)
		if err != nil {
			return err
		}
		movementID = id
		return recalculateAlerts(ctx, tx, productID, body.LocationID)
	})
	if err != nil {
		return StockMovement{}, err
	}
	return s.stockMovementByID(ctx, movementID)
}

func (s *Server) adjustStock(ctx context.Context, user User, productID uint64, body AdjustmentInput) (StockMovement, error) {
	if body.Quantity == 0 {
		return StockMovement{}, errors.New("adjustment quantity cannot be 0")
	}
	if body.Note == "" {
		return StockMovement{}, errors.New("note is required")
	}

	var movementID uint64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		before, err := lockedStock(ctx, tx, productID, body.LocationID)
		if err != nil {
			return err
		}
		after := before + body.Quantity
		if after < 0 {
			return errors.New("stock cannot become negative")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, after, productID, body.LocationID); err != nil {
			return err
		}
		id, err := insertStockMovement(ctx, tx, productID, body.LocationID, "ADJUSTMENT", body.Quantity, before, after, nil, body.Note, user.ID)
		if err != nil {
			return err
		}
		movementID = id
		return recalculateAlerts(ctx, tx, productID, body.LocationID)
	})
	if err != nil {
		return StockMovement{}, err
	}
	return s.stockMovementByID(ctx, movementID)
}

func (s *Server) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func lockedStock(ctx context.Context, tx *sql.Tx, productID, locationID uint64) (int, error) {
	if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO product_stocks(product_id, location_id, quantity) VALUES (?, ?, 0)`, productID, locationID); err != nil {
		return 0, err
	}
	var quantity int
	err := tx.QueryRowContext(ctx, `SELECT quantity FROM product_stocks WHERE product_id=? AND location_id=? FOR UPDATE`, productID, locationID).Scan(&quantity)
	return quantity, err
}

func insertStockMovement(ctx context.Context, tx *sql.Tx, productID, locationID uint64, movementType string, delta, before, after int, unitCost *float64, note string, userID uint64) (uint64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO stock_movements(product_id, location_id, reference_type, quantity_change, before_stock, after_stock, quantity_after, unit_cost, note, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		productID, locationID, movementType, delta, before, after, after, unitCost, note, userID)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

func (s *Server) stockMovementByID(ctx context.Context, id uint64) (StockMovement, error) {
	var item StockMovement
	err := s.db.QueryRowContext(ctx, `
		SELECT sm.id, sm.product_id, p.name, p.sku, sm.location_id, l.name, sm.reference_type, sm.reference_id,
		       sm.quantity_change, sm.before_stock, sm.after_stock, sm.unit_cost, sm.note, sm.created_by, sm.created_at
		FROM stock_movements sm
		JOIN products p ON p.id=sm.product_id
		JOIN locations l ON l.id=sm.location_id
		WHERE sm.id=?`, id).Scan(&item.ID, &item.ProductID, &item.ProductName, &item.SKU, &item.LocationID, &item.LocationName, &item.ReferenceType, &item.ReferenceID, &item.QuantityChange, &item.BeforeStock, &item.AfterStock, &item.UnitCost, &item.Note, &item.CreatedBy, &item.CreatedAt)
	return item, err
}

func stockErrorCode(err error) string {
	if err.Error() == "stock cannot become negative" {
		return "NEGATIVE_STOCK"
	}
	return "STOCK_VALIDATION_FAILED"
}

func stockErrorStatus(err error) int {
	if err.Error() == "stock cannot become negative" {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
