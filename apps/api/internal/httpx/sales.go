package httpx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type POSProduct struct {
	ID           uint64  `json:"id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Barcode      *string `json:"barcode"`
	SellingPrice float64 `json:"selling_price"`
	UnitCost     float64 `json:"unit_cost"`
	Unit         string  `json:"unit"`
	Threshold    int     `json:"threshold"`
	ReorderPoint int     `json:"reorder_point"`
	LocationID   uint64  `json:"location_id"`
	Stock        int     `json:"stock"`
	StockStatus  string  `json:"stock_status"`
}

type SaleInput struct {
	LocationID     uint64          `json:"location_id"`
	PaymentMethod  string          `json:"payment_method"`
	ReceivedAmount float64         `json:"received_amount"`
	Items          []SaleInputItem `json:"items"`
}

type SaleInputItem struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CancelSaleInput struct {
	Reason string `json:"reason"`
}

type Receipt struct {
	ID            uint64        `json:"id"`
	ReceiptNo     string        `json:"receipt_no"`
	LocationID    uint64        `json:"location_id"`
	LocationName  string        `json:"location_name"`
	CashierID     uint64        `json:"cashier_id"`
	CashierName   string        `json:"cashier_name"`
	Subtotal      float64       `json:"subtotal"`
	TotalAmount   float64       `json:"total_amount"`
	TotalCost     float64       `json:"total_cost"`
	Profit        float64       `json:"profit"`
	PaymentMethod string        `json:"payment_method"`
	PaidAmount    float64       `json:"paid_amount"`
	ChangeAmount  float64       `json:"change_amount"`
	Status        string        `json:"status"`
	CancelledBy   *uint64       `json:"cancelled_by"`
	CancelledAt   *time.Time    `json:"cancelled_at"`
	CancelReason  string        `json:"cancel_reason"`
	CreatedAt     time.Time     `json:"created_at"`
	Items         []ReceiptItem `json:"items"`
}

type ReceiptItem struct {
	ID          uint64  `json:"id"`
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Barcode     *string `json:"barcode"`
	Price       float64 `json:"price"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	LineTotal   float64 `json:"line_total"`
	LineCost    float64 `json:"line_cost"`
	BeforeStock int     `json:"before_stock,omitempty"`
	AfterStock  int     `json:"after_stock,omitempty"`
}

func (s *Server) posProducts(w http.ResponseWriter, r *http.Request) {
	locationID, err := parseQueryUint(r, "location_id")
	if err != nil || locationID == 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "location_id is required.")
		return
	}
	products, err := s.listPOSProducts(r.Context(), locationID, strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load POS products.")
		return
	}
	response.JSON(w, http.StatusOK, products)
}

func (s *Server) sales(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sales, err := s.listSales(r.Context(), r)
		if err != nil {
			if strings.HasPrefix(err.Error(), "invalid ") {
				response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
				return
			}
			response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load sales.")
			return
		}
		response.JSON(w, http.StatusOK, sales)
	case http.MethodPost:
		var body SaleInput
		if err := readJSON(r, &body); err != nil {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		user, _ := currentUser(r.Context())
		receipt, err := s.createSale(r.Context(), user, body)
		if err != nil {
			response.ErrorJSON(w, saleErrorStatus(err), saleErrorCode(err), err.Error())
			return
		}
		response.JSON(w, http.StatusCreated, receipt)
	default:
		response.ErrorJSON(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed.")
	}
}

func (s *Server) saleDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid sale id.")
		return
	}
	receipt, err := s.receiptByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Sale not found.")
		return
	}
	response.JSON(w, http.StatusOK, receipt)
}

func (s *Server) saleReceipt(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid sale id.")
		return
	}
	receipt, err := s.receiptByID(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Receipt not found.")
		return
	}
	response.JSON(w, http.StatusOK, receipt)
}

func (s *Server) cancelSale(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid sale id.")
		return
	}
	var body CancelSaleInput
	if err := readJSON(r, &body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	user, _ := currentUser(r.Context())
	receipt, err := s.cancelCompletedSale(r.Context(), user, id, body)
	if err != nil {
		response.ErrorJSON(w, saleErrorStatus(err), saleErrorCode(err), err.Error())
		return
	}
	response.JSON(w, http.StatusOK, receipt)
}

func (s *Server) listPOSProducts(ctx context.Context, locationID uint64, query string) ([]POSProduct, error) {
	where := []string{"p.active=TRUE", "l.active=TRUE", "l.id=?"}
	args := []any{locationID}
	if query != "" {
		where = append(where, "(p.name LIKE ? OR p.sku LIKE ? OR p.barcode LIKE ?)")
		like := "%" + query + "%"
		args = append(args, like, like, like)
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.sku, p.name, p.barcode, p.price, p.cost, p.unit, p.threshold, p.reorder_point,
		       l.id, COALESCE(ps.quantity, 0)
		FROM products p
		JOIN locations l ON l.id=?
		LEFT JOIN product_stocks ps ON ps.product_id=p.id AND ps.location_id=l.id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY p.name, p.sku
		LIMIT 120`, append([]any{locationID}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []POSProduct{}
	for rows.Next() {
		var item POSProduct
		if err := rows.Scan(&item.ID, &item.SKU, &item.Name, &item.Barcode, &item.SellingPrice, &item.UnitCost, &item.Unit, &item.Threshold, &item.ReorderPoint, &item.LocationID, &item.Stock); err != nil {
			return nil, err
		}
		item.StockStatus = stockStatus(item.Stock, item.Threshold, item.ReorderPoint)
		products = append(products, item)
	}
	return products, rows.Err()
}

func (s *Server) listSales(ctx context.Context, r *http.Request) ([]Receipt, error) {
	where := []string{"1=1"}
	args := []any{}
	query := r.URL.Query()

	if value := strings.TrimSpace(query.Get("date_from")); value != "" {
		where = append(where, "s.created_at >= ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("date_to")); value != "" {
		where = append(where, "s.created_at < DATE_ADD(?, INTERVAL 1 DAY)")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("cashier_id")); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, errors.New("invalid cashier_id")
		}
		where = append(where, "s.cashier_id=?")
		args = append(args, id)
	}
	if value := strings.TrimSpace(query.Get("location_id")); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, errors.New("invalid location_id")
		}
		where = append(where, "s.location_id=?")
		args = append(args, id)
	}
	if value := strings.TrimSpace(query.Get("payment_method")); value != "" {
		where = append(where, "s.payment_method=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("status")); value != "" {
		where = append(where, "s.status=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("receipt_no")); value != "" {
		where = append(where, "s.receipt_no LIKE ?")
		args = append(args, "%"+value+"%")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.receipt_no, s.location_id, l.name, s.cashier_id, u.full_name,
		       s.subtotal, s.total_amount, s.total_cost, s.profit, s.payment_method,
		       s.paid_amount, s.change_amount, s.status, s.cancelled_by, s.cancelled_at,
		       s.cancel_reason, s.created_at
		FROM sales s
		JOIN locations l ON l.id=s.location_id
		JOIN users u ON u.id=s.cashier_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY s.id DESC
		LIMIT 300`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []Receipt{}
	for rows.Next() {
		sale, err := scanReceiptHeader(rows)
		if err != nil {
			return nil, err
		}
		sales = append(sales, sale)
	}
	return sales, rows.Err()
}

func (s *Server) createSale(ctx context.Context, user User, body SaleInput) (Receipt, error) {
	if err := validateSaleInput(body); err != nil {
		return Receipt{}, err
	}

	var saleID uint64
	var receiptNo string
	var totalAmount float64
	affectedProducts := []uint64{}
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var locationActive bool
		if err := tx.QueryRowContext(ctx, `SELECT active FROM locations WHERE id=? FOR UPDATE`, body.LocationID).Scan(&locationActive); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("location not found")
			}
			return err
		}
		if !locationActive {
			return errors.New("location is inactive")
		}

		quantities := map[uint64]int{}
		for _, item := range body.Items {
			quantities[item.ProductID] += item.Quantity
		}

		lines := []ReceiptItem{}
		var subtotal float64
		var totalCost float64
		for productID, quantity := range quantities {
			if quantity <= 0 {
				return errors.New("cart item quantity must be greater than 0")
			}
			line, before, err := lockedSaleProduct(ctx, tx, productID, body.LocationID)
			if err != nil {
				return err
			}
			if before < quantity {
				return errors.New("insufficient stock")
			}
			line.Quantity = quantity
			line.LineTotal = money(line.Price * float64(quantity))
			line.LineCost = money(line.Cost * float64(quantity))
			line.BeforeStock = before
			line.AfterStock = before - quantity
			subtotal = money(subtotal + line.LineTotal)
			totalCost = money(totalCost + line.LineCost)
			lines = append(lines, line)
		}

		totalAmount = subtotal
		if body.ReceivedAmount+0.0001 < totalAmount {
			return errors.New("received amount is insufficient")
		}
		profit := money(totalAmount - totalCost)
		changeAmount := money(body.ReceivedAmount - totalAmount)
		now := time.Now()
		receiptNo = fmt.Sprintf("RC%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
		result, err := tx.ExecContext(ctx, `
			INSERT INTO sales(receipt_no, location_id, cashier_id, subtotal, total_amount, total_cost, profit, payment_method, paid_amount, change_amount, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'COMPLETED')`,
			receiptNo, body.LocationID, user.ID, subtotal, totalAmount, totalCost, profit, body.PaymentMethod, body.ReceivedAmount, changeAmount)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		saleID = uint64(id)

		for _, line := range lines {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO sale_items(sale_id, product_id, product_name_snapshot, sku_snapshot, barcode_snapshot, price_snapshot, cost_snapshot, quantity, line_total, line_cost)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				saleID, line.ProductID, line.ProductName, line.SKU, line.Barcode, line.Price, line.Cost, line.Quantity, line.LineTotal, line.LineCost); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, line.AfterStock, line.ProductID, body.LocationID); err != nil {
				return err
			}
			note := fmt.Sprintf("sale %s", receiptNo)
			if _, err := insertSaleMovement(ctx, tx, line.ProductID, body.LocationID, -line.Quantity, line.BeforeStock, line.AfterStock, saleID, note, user.ID); err != nil {
				return err
			}
			if err := recalculateAlerts(ctx, tx, line.ProductID, body.LocationID); err != nil {
				return err
			}
			affectedProducts = append(affectedProducts, line.ProductID)
		}
		return nil
	})
	if err != nil {
		return Receipt{}, err
	}
	receipt, err := s.receiptByID(ctx, saleID)
	if err != nil {
		return Receipt{}, err
	}
	s.notifyEvent(ctx, "SALE_COMPLETED", fmt.Sprintf("Sale completed %s total %.2f", receiptNo, totalAmount), map[string]any{
		"sale_id":     saleID,
		"receipt_no":  receiptNo,
		"location_id": body.LocationID,
		"total":       totalAmount,
	})
	for _, productID := range affectedProducts {
		s.notifyActiveStockAlerts(ctx, productID, body.LocationID)
	}
	return receipt, nil
}

func (s *Server) cancelCompletedSale(ctx context.Context, user User, saleID uint64, body CancelSaleInput) (Receipt, error) {
	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		return Receipt{}, errors.New("cancel reason is required")
	}
	if len(reason) > 255 {
		reason = reason[:255]
	}

	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var receiptNo, status string
		var locationID uint64
		if err := tx.QueryRowContext(ctx, `SELECT receipt_no, location_id, status FROM sales WHERE id=? FOR UPDATE`, saleID).Scan(&receiptNo, &locationID, &status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("sale not found")
			}
			return err
		}
		if status != "COMPLETED" {
			return errors.New("only completed sale can be cancelled")
		}

		rows, err := tx.QueryContext(ctx, `SELECT product_id, quantity, cost_snapshot FROM sale_items WHERE sale_id=? ORDER BY id`, saleID)
		if err != nil {
			return err
		}
		defer rows.Close()
		type cancelLine struct {
			productID uint64
			quantity  int
			cost      float64
		}
		lines := []cancelLine{}
		for rows.Next() {
			var line cancelLine
			if err := rows.Scan(&line.productID, &line.quantity, &line.cost); err != nil {
				return err
			}
			lines = append(lines, line)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(lines) == 0 {
			return errors.New("sale has no items")
		}

		for _, line := range lines {
			before, err := lockedStock(ctx, tx, line.productID, locationID)
			if err != nil {
				return err
			}
			after := before + line.quantity
			if _, err := tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, after, line.productID, locationID); err != nil {
				return err
			}
			note := fmt.Sprintf("cancel sale %s: %s", receiptNo, reason)
			if _, err := insertCancelSaleMovement(ctx, tx, line.productID, locationID, line.quantity, before, after, line.cost, saleID, note, user.ID); err != nil {
				return err
			}
			if err := recalculateAlerts(ctx, tx, line.productID, locationID); err != nil {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, `UPDATE sales SET status='CANCELLED', cancelled_by=?, cancelled_at=NOW(), cancel_reason=? WHERE id=?`, user.ID, reason, saleID)
		return err
	})
	if err != nil {
		return Receipt{}, err
	}
	return s.receiptByID(ctx, saleID)
}

func lockedSaleProduct(ctx context.Context, tx *sql.Tx, productID, locationID uint64) (ReceiptItem, int, error) {
	if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO product_stocks(product_id, location_id, quantity) VALUES (?, ?, 0)`, productID, locationID); err != nil {
		return ReceiptItem{}, 0, err
	}
	var line ReceiptItem
	var active bool
	var stock int
	if err := tx.QueryRowContext(ctx, `SELECT quantity FROM product_stocks WHERE product_id=? AND location_id=? FOR UPDATE`, productID, locationID).Scan(&stock); err != nil {
		return ReceiptItem{}, 0, err
	}
	err := tx.QueryRowContext(ctx, `
		SELECT id, name, sku, barcode, price, cost, active
		FROM products
		WHERE id=?`, productID).Scan(&line.ProductID, &line.ProductName, &line.SKU, &line.Barcode, &line.Price, &line.Cost, &active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ReceiptItem{}, 0, errors.New("product not found")
		}
		return ReceiptItem{}, 0, err
	}
	if !active {
		return ReceiptItem{}, 0, errors.New("product is inactive")
	}
	return line, stock, nil
}

func insertSaleMovement(ctx context.Context, tx *sql.Tx, productID, locationID uint64, delta, before, after int, saleID uint64, note string, userID uint64) (uint64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO stock_movements(product_id, location_id, reference_type, reference_id, quantity_change, before_stock, after_stock, quantity_after, note, created_by)
		VALUES (?, ?, 'SALE', ?, ?, ?, ?, ?, ?, ?)`,
		productID, locationID, saleID, delta, before, after, after, note, userID)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

func insertCancelSaleMovement(ctx context.Context, tx *sql.Tx, productID, locationID uint64, delta, before, after int, unitCost float64, saleID uint64, note string, userID uint64) (uint64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO stock_movements(product_id, location_id, reference_type, reference_id, quantity_change, before_stock, after_stock, quantity_after, unit_cost, note, created_by)
		VALUES (?, ?, 'CANCEL_SALE', ?, ?, ?, ?, ?, ?, ?, ?)`,
		productID, locationID, saleID, delta, before, after, after, unitCost, note, userID)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

func (s *Server) receiptByID(ctx context.Context, id uint64) (Receipt, error) {
	receipt, err := scanReceiptHeader(s.db.QueryRowContext(ctx, `
		SELECT s.id, s.receipt_no, s.location_id, l.name, s.cashier_id, u.full_name,
		       s.subtotal, s.total_amount, s.total_cost, s.profit, s.payment_method,
		       s.paid_amount, s.change_amount, s.status, s.cancelled_by, s.cancelled_at,
		       s.cancel_reason, s.created_at
		FROM sales s
		JOIN locations l ON l.id=s.location_id
		JOIN users u ON u.id=s.cashier_id
		WHERE s.id=?`, id))
	if err != nil {
		return Receipt{}, err
	}
	receipt.Items, err = s.receiptItems(ctx, id)
	return receipt, err
}

type receiptHeaderScanner interface {
	Scan(dest ...any) error
}

func scanReceiptHeader(scanner receiptHeaderScanner) (Receipt, error) {
	var receipt Receipt
	err := scanner.Scan(&receipt.ID, &receipt.ReceiptNo, &receipt.LocationID, &receipt.LocationName, &receipt.CashierID, &receipt.CashierName, &receipt.Subtotal, &receipt.TotalAmount, &receipt.TotalCost, &receipt.Profit, &receipt.PaymentMethod, &receipt.PaidAmount, &receipt.ChangeAmount, &receipt.Status, &receipt.CancelledBy, &receipt.CancelledAt, &receipt.CancelReason, &receipt.CreatedAt)
	return receipt, err
}

func (s *Server) receiptItems(ctx context.Context, saleID uint64) ([]ReceiptItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, product_id, product_name_snapshot, sku_snapshot, barcode_snapshot,
		       price_snapshot, cost_snapshot, quantity, line_total, line_cost
		FROM sale_items
		WHERE sale_id=?
		ORDER BY id`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ReceiptItem{}
	for rows.Next() {
		var item ReceiptItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.SKU, &item.Barcode, &item.Price, &item.Cost, &item.Quantity, &item.LineTotal, &item.LineCost); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func validateSaleInput(body SaleInput) error {
	if body.LocationID == 0 {
		return errors.New("location_id is required")
	}
	if body.PaymentMethod != "CASH" && body.PaymentMethod != "QR" {
		return errors.New("payment_method must be CASH or QR")
	}
	if body.ReceivedAmount < 0 {
		return errors.New("received amount must be greater than or equal to 0")
	}
	if len(body.Items) == 0 {
		return errors.New("cart requires at least one item")
	}
	for _, item := range body.Items {
		if item.ProductID == 0 || item.Quantity <= 0 {
			return errors.New("cart item product and positive quantity are required")
		}
	}
	return nil
}

func saleErrorCode(err error) string {
	switch err.Error() {
	case "insufficient stock":
		return "INSUFFICIENT_STOCK"
	case "received amount is insufficient":
		return "INSUFFICIENT_PAYMENT"
	case "only completed sale can be cancelled":
		return "SALE_ALREADY_CANCELLED"
	case "product not found", "location not found", "sale not found":
		return "NOT_FOUND"
	}
	return "SALE_VALIDATION_FAILED"
}

func saleErrorStatus(err error) int {
	switch err.Error() {
	case "insufficient stock", "received amount is insufficient", "only completed sale can be cancelled":
		return http.StatusConflict
	case "product not found", "location not found", "sale not found":
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func parseQueryUint(r *http.Request, name string) (uint64, error) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

func money(value float64) float64 {
	return math.Round(value*100) / 100
}
