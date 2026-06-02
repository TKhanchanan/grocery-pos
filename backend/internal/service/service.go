package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"grocery-pos/backend/internal/config"
	"grocery-pos/backend/internal/models"
	"grocery-pos/backend/internal/repo"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Services struct {
	repo *repo.Repository
	cfg  config.Config
}

func New(r *repo.Repository, cfg config.Config) *Services {
	return &Services{repo: r, cfg: cfg}
}

func (s *Services) Login(ctx context.Context, in models.LoginRequest) (models.LoginResponse, error) {
	user, hash, err := s.repo.UserByUsername(ctx, in.Username)
	if err != nil || !user.Active {
		return models.LoginResponse{}, errors.New("invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
		return models.LoginResponse{}, errors.New("invalid username or password")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", user.ID),
		"role": string(user.Role),
		"exp":  time.Now().Add(12 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return models.LoginResponse{}, err
	}
	return models.LoginResponse{Token: tokenString, User: user}, nil
}

func (s *Services) UserFromToken(ctx context.Context, tokenString string) (models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return models.User{}, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return models.User{}, errors.New("invalid token claims")
	}
	sub, _ := claims["sub"].(string)
	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return models.User{}, errors.New("invalid token subject")
	}
	return s.repo.UserByID(ctx, id)
}

func (s *Services) Users(ctx context.Context) ([]models.User, error) {
	return s.repo.Users(ctx)
}

func (s *Services) SaveUser(ctx context.Context, id uint64, in models.UserInput) (uint64, error) {
	if strings.TrimSpace(in.Username) == "" || strings.TrimSpace(in.FullName) == "" {
		return 0, errors.New("username and full name are required")
	}
	if in.Role == "" {
		in.Role = models.RoleCashier
	}
	hash := ""
	if in.Password != "" {
		bytes, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return 0, err
		}
		hash = string(bytes)
	}
	if id == 0 && hash == "" {
		return 0, errors.New("password is required for a new user")
	}
	return s.repo.UpsertUser(ctx, id, in, hash)
}

func (s *Services) Categories(ctx context.Context) ([]models.Category, error) {
	return s.repo.Categories(ctx)
}

func (s *Services) SaveCategory(ctx context.Context, id uint64, name string) (uint64, error) {
	if strings.TrimSpace(name) == "" {
		return 0, errors.New("category name is required")
	}
	return s.repo.UpsertCategory(ctx, id, strings.TrimSpace(name))
}

func (s *Services) Locations(ctx context.Context) ([]models.Location, error) {
	return s.repo.Locations(ctx)
}

func (s *Services) SaveLocation(ctx context.Context, id uint64, name string, active bool) (uint64, error) {
	if strings.TrimSpace(name) == "" {
		return 0, errors.New("location name is required")
	}
	return s.repo.UpsertLocation(ctx, id, strings.TrimSpace(name), active)
}

func (s *Services) Products(ctx context.Context, q string) ([]models.Product, error) {
	return s.repo.Products(ctx, q)
}

func (s *Services) SaveProduct(ctx context.Context, id uint64, in models.ProductInput) (uint64, error) {
	if strings.TrimSpace(in.SKU) == "" || strings.TrimSpace(in.Name) == "" {
		return 0, errors.New("SKU and product name are required")
	}
	if in.Price < 0 || in.Cost < 0 {
		return 0, errors.New("price and cost cannot be negative")
	}
	return s.repo.UpsertProduct(ctx, id, in)
}

func (s *Services) Stocks(ctx context.Context) ([]models.ProductStock, error) {
	return s.repo.Stocks(ctx)
}

func (s *Services) Restock(ctx context.Context, user models.User, in models.StockChangeRequest) error {
	if in.Quantity <= 0 {
		return errors.New("restock quantity must be positive")
	}
	userID := user.ID
	unitCost := in.UnitCost
	return s.repo.Tx(ctx, func(tx *sql.Tx) error {
		if err := repo.ChangeStock(ctx, tx, in.ProductID, in.LocationID, in.Quantity, &unitCost, "RESTOCK", nil, in.Reason, &userID); err != nil {
			return err
		}
		if in.UnitCost > 0 {
			_, _ = tx.ExecContext(ctx, `UPDATE products SET cost=? WHERE id=?`, in.UnitCost, in.ProductID)
		}
		if err := repo.UpsertAlertsForStock(ctx, tx, in.ProductID, in.LocationID); err != nil {
			return err
		}
		return s.notifyLine(ctx, "Restocked "+strconv.Itoa(in.Quantity)+" units")
	})
}

func (s *Services) AdjustStock(ctx context.Context, user models.User, in models.StockChangeRequest) error {
	if in.Quantity == 0 {
		return errors.New("adjustment quantity cannot be zero")
	}
	userID := user.ID
	return s.repo.Tx(ctx, func(tx *sql.Tx) error {
		if err := repo.ChangeStock(ctx, tx, in.ProductID, in.LocationID, in.Quantity, nil, "ADJUSTMENT", nil, in.Reason, &userID); err != nil {
			return err
		}
		return repo.UpsertAlertsForStock(ctx, tx, in.ProductID, in.LocationID)
	})
}

func (s *Services) TransferStock(ctx context.Context, user models.User, in models.StockTransferRequest) error {
	if in.Quantity <= 0 {
		return errors.New("transfer quantity must be positive")
	}
	if in.FromLocationID == in.ToLocationID {
		return errors.New("source and destination locations must differ")
	}
	userID := user.ID
	return s.repo.Tx(ctx, func(tx *sql.Tx) error {
		if err := repo.ChangeStock(ctx, tx, in.ProductID, in.FromLocationID, -in.Quantity, nil, "TRANSFER_OUT", nil, in.Reason, &userID); err != nil {
			return err
		}
		if err := repo.ChangeStock(ctx, tx, in.ProductID, in.ToLocationID, in.Quantity, nil, "TRANSFER_IN", nil, in.Reason, &userID); err != nil {
			return err
		}
		if err := repo.UpsertAlertsForStock(ctx, tx, in.ProductID, in.FromLocationID); err != nil {
			return err
		}
		return repo.UpsertAlertsForStock(ctx, tx, in.ProductID, in.ToLocationID)
	})
}

func (s *Services) Movements(ctx context.Context) ([]models.StockMovement, error) {
	return s.repo.Movements(ctx)
}

func (s *Services) Alerts(ctx context.Context) ([]models.Alert, error) {
	return s.repo.Alerts(ctx)
}

func (s *Services) CreateSale(ctx context.Context, user models.User, in models.CreateSaleRequest) (uint64, error) {
	if len(in.Items) == 0 {
		return 0, errors.New("sale requires at least one item")
	}
	if in.PaymentMethod == "" {
		in.PaymentMethod = "CASH"
	}
	var saleID uint64
	err := s.repo.Tx(ctx, func(tx *sql.Tx) error {
		type line struct {
			product models.Product
			qty     int
		}
		lines := make([]line, 0, len(in.Items))
		total, totalCost := 0.0, 0.0
		for _, item := range in.Items {
			if item.Quantity <= 0 {
				return errors.New("sale item quantity must be positive")
			}
			var p models.Product
			err := tx.QueryRowContext(ctx, `SELECT id, category_id, sku, barcode, name, unit, price, cost, threshold, reorder_point, active, created_at, 0 FROM products WHERE id=? AND active=1`, item.ProductID).
				Scan(&p.ID, &p.CategoryID, &p.SKU, &p.Barcode, &p.Name, &p.Unit, &p.Price, &p.Cost, &p.Threshold, &p.ReorderPoint, &p.Active, &p.CreatedAt, &p.TotalStock)
			if err != nil {
				return fmt.Errorf("product %d not found", item.ProductID)
			}
			currentStock, err := repo.LockedStock(ctx, tx, item.ProductID, in.LocationID)
			if err != nil {
				return err
			}
			if currentStock < item.Quantity {
				return fmt.Errorf("insufficient stock for %s", p.Name)
			}
			total += p.Price * float64(item.Quantity)
			totalCost += p.Cost * float64(item.Quantity)
			lines = append(lines, line{product: p, qty: item.Quantity})
		}
		if in.PaidAmount < total {
			return errors.New("payment is insufficient")
		}
		receiptNo := "R" + time.Now().Format("20060102150405")
		res, err := tx.ExecContext(ctx, `
			INSERT INTO sales(receipt_no, location_id, cashier_id, total_amount, total_cost, profit, payment_method, paid_amount, change_amount, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'COMPLETED')`,
			receiptNo, in.LocationID, user.ID, total, totalCost, total-totalCost, in.PaymentMethod, in.PaidAmount, in.PaidAmount-total)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		saleID = uint64(id)
		for _, line := range lines {
			lineTotal := line.product.Price * float64(line.qty)
			lineCost := line.product.Cost * float64(line.qty)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO sale_items(sale_id, product_id, product_name_snapshot, sku_snapshot, price_snapshot, cost_snapshot, quantity, line_total, line_cost)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				saleID, line.product.ID, line.product.Name, line.product.SKU, line.product.Price, line.product.Cost, line.qty, lineTotal, lineCost)
			if err != nil {
				return err
			}
			if err := repo.ChangeStock(ctx, tx, line.product.ID, in.LocationID, -line.qty, nil, "SALE", &saleID, receiptNo, &user.ID); err != nil {
				return err
			}
			if err := repo.UpsertAlertsForStock(ctx, tx, line.product.ID, in.LocationID); err != nil {
				return err
			}
		}
		return s.notifyLine(ctx, fmt.Sprintf("Sale %s total %.2f", receiptNo, total))
	})
	return saleID, err
}

func (s *Services) Sales(ctx context.Context) ([]models.Sale, error) {
	sales, err := s.repo.Sales(ctx)
	if err != nil {
		return nil, err
	}
	for i := range sales {
		items, _ := s.repo.SaleItems(ctx, sales[i].ID)
		sales[i].Items = items
	}
	return sales, nil
}

func (s *Services) Sale(ctx context.Context, id uint64) (models.Sale, error) {
	var sale models.Sale
	err := s.repo.DB.QueryRowContext(ctx, `
		SELECT s.id, s.receipt_no, s.location_id, l.name, s.cashier_id, s.total_amount, s.total_cost, s.profit,
		       s.payment_method, s.paid_amount, s.change_amount, s.status, s.cancelled_at, s.created_at
		FROM sales s JOIN locations l ON l.id=s.location_id WHERE s.id=?`, id).
		Scan(&sale.ID, &sale.ReceiptNo, &sale.LocationID, &sale.LocationName, &sale.CashierID, &sale.TotalAmount, &sale.TotalCost, &sale.Profit, &sale.PaymentMethod, &sale.PaidAmount, &sale.ChangeAmount, &sale.Status, &sale.CancelledAt, &sale.CreatedAt)
	if err != nil {
		return models.Sale{}, err
	}
	sale.Items, _ = s.repo.SaleItems(ctx, id)
	return sale, nil
}

func (s *Services) CancelSale(ctx context.Context, user models.User, id uint64) error {
	return s.repo.Tx(ctx, func(tx *sql.Tx) error {
		var status string
		var locationID uint64
		err := tx.QueryRowContext(ctx, `SELECT status, location_id FROM sales WHERE id=? FOR UPDATE`, id).Scan(&status, &locationID)
		if err != nil {
			return err
		}
		if status == "CANCELLED" {
			return errors.New("sale is already cancelled")
		}
		rows, err := tx.QueryContext(ctx, `SELECT product_id, quantity FROM sale_items WHERE sale_id=?`, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var productID uint64
			var qty int
			if err := rows.Scan(&productID, &qty); err != nil {
				return err
			}
			if err := repo.ChangeStock(ctx, tx, productID, locationID, qty, nil, "SALE_CANCEL", &id, "sale cancelled", &user.ID); err != nil {
				return err
			}
			if err := repo.UpsertAlertsForStock(ctx, tx, productID, locationID); err != nil {
				return err
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE sales SET status='CANCELLED', cancelled_at=NOW() WHERE id=?`, id)
		return err
	})
}

func (s *Services) Suppliers(ctx context.Context) ([]models.Supplier, error) {
	return s.repo.Suppliers(ctx)
}

func (s *Services) SaveSupplier(ctx context.Context, supplier models.Supplier) (uint64, error) {
	if strings.TrimSpace(supplier.Name) == "" {
		return 0, errors.New("supplier name is required")
	}
	return s.repo.UpsertSupplier(ctx, supplier)
}

func (s *Services) PurchaseOrders(ctx context.Context) ([]models.PurchaseOrder, error) {
	return s.repo.PurchaseOrders(ctx)
}

func (s *Services) CreatePurchaseOrder(ctx context.Context, in models.PurchaseOrderInput) (uint64, error) {
	if len(in.Items) == 0 {
		return 0, errors.New("purchase order requires items")
	}
	var id uint64
	err := s.repo.Tx(ctx, func(tx *sql.Tx) error {
		total := 0.0
		for _, item := range in.Items {
			if item.Quantity <= 0 || item.UnitCost < 0 {
				return errors.New("invalid purchase order item")
			}
			total += float64(item.Quantity) * item.UnitCost
		}
		poNumber := "PO" + time.Now().Format("20060102150405")
		res, err := tx.ExecContext(ctx, `INSERT INTO purchase_orders(po_number, supplier_id, location_id, status, total_cost) VALUES (?, ?, ?, 'OPEN', ?)`, poNumber, in.SupplierID, in.LocationID, total)
		if err != nil {
			return err
		}
		newID, _ := res.LastInsertId()
		id = uint64(newID)
		for _, item := range in.Items {
			_, err := tx.ExecContext(ctx, `INSERT INTO purchase_order_items(po_id, product_id, quantity, unit_cost, line_cost) VALUES (?, ?, ?, ?, ?)`, id, item.ProductID, item.Quantity, item.UnitCost, float64(item.Quantity)*item.UnitCost)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return id, err
}

func (s *Services) ReceivePurchaseOrder(ctx context.Context, user models.User, id uint64) error {
	return s.repo.Tx(ctx, func(tx *sql.Tx) error {
		var status string
		var locationID uint64
		if err := tx.QueryRowContext(ctx, `SELECT status, location_id FROM purchase_orders WHERE id=? FOR UPDATE`, id).Scan(&status, &locationID); err != nil {
			return err
		}
		if status == "RECEIVED" {
			return errors.New("purchase order is already received")
		}
		rows, err := tx.QueryContext(ctx, `SELECT product_id, quantity, unit_cost FROM purchase_order_items WHERE po_id=?`, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var productID uint64
			var qty int
			var cost float64
			if err := rows.Scan(&productID, &qty, &cost); err != nil {
				return err
			}
			if err := repo.ChangeStock(ctx, tx, productID, locationID, qty, &cost, "PO_RECEIVE", &id, "purchase order received", &user.ID); err != nil {
				return err
			}
			_, _ = tx.ExecContext(ctx, `UPDATE products SET cost=? WHERE id=?`, cost, productID)
			if err := repo.UpsertAlertsForStock(ctx, tx, productID, locationID); err != nil {
				return err
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE purchase_orders SET status='RECEIVED', received_at=NOW() WHERE id=?`, id)
		return err
	})
}

func (s *Services) Dashboard(ctx context.Context) (models.ReportSummary, error) {
	var out models.ReportSummary
	_ = s.repo.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_amount),0), COALESCE(SUM(profit),0), COUNT(*) FROM sales WHERE status='COMPLETED' AND DATE(created_at)=CURRENT_DATE`).Scan(&out.Revenue, &out.Profit, &out.SalesCount)
	_ = s.repo.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity),0) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.status='COMPLETED' AND DATE(s.created_at)=CURRENT_DATE`).Scan(&out.ItemsSold)
	_ = s.repo.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(ps.quantity * p.cost),0) FROM product_stocks ps JOIN products p ON p.id=ps.product_id`).Scan(&out.InventoryVal)
	_ = s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE resolved_at IS NULL AND type='LOW_STOCK'`).Scan(&out.LowAlerts)
	_ = s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE resolved_at IS NULL AND type='OUT_OF_STOCK'`).Scan(&out.OutAlerts)
	return out, nil
}

func (s *Services) Report(ctx context.Context, name string) ([]models.ReportRow, error) {
	queries := map[string]string{
		"daily-sales":     `SELECT DATE(created_at) label, COUNT(*) sales, SUM(total_amount) revenue, SUM(profit) profit FROM sales WHERE status='COMPLETED' GROUP BY DATE(created_at) ORDER BY label DESC`,
		"monthly-sales":   `SELECT DATE_FORMAT(created_at, '%Y-%m') label, COUNT(*) sales, SUM(total_amount) revenue, SUM(profit) profit FROM sales WHERE status='COMPLETED' GROUP BY label ORDER BY label DESC`,
		"best-selling":    `SELECT si.product_name_snapshot product, SUM(si.quantity) quantity, SUM(si.line_total) revenue FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.status='COMPLETED' GROUP BY si.product_id, si.product_name_snapshot ORDER BY quantity DESC`,
		"profit-products": `SELECT si.product_name_snapshot product, SUM(si.line_total) revenue, SUM(si.line_cost) cost, SUM(si.line_total-si.line_cost) profit FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.status='COMPLETED' GROUP BY si.product_id, si.product_name_snapshot ORDER BY profit DESC`,
		"stock":           `SELECT l.name location, p.name product, p.sku sku, ps.quantity quantity, p.threshold threshold, p.reorder_point reorder_point FROM product_stocks ps JOIN products p ON p.id=ps.product_id JOIN locations l ON l.id=ps.location_id ORDER BY l.name, p.name`,
		"valuation":       `SELECT l.name location, p.name product, ps.quantity quantity, p.cost cost, ps.quantity*p.cost value FROM product_stocks ps JOIN products p ON p.id=ps.product_id JOIN locations l ON l.id=ps.location_id ORDER BY value DESC`,
		"payments":        `SELECT payment_method method, COUNT(*) sales, SUM(total_amount) revenue FROM sales WHERE status='COMPLETED' GROUP BY payment_method`,
	}
	query, ok := queries[name]
	if !ok {
		return nil, errors.New("unknown report")
	}
	rows, err := s.repo.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []models.ReportRow
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := models.ReportRow{}
		for i, col := range cols {
			switch v := values[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Services) ExportCSV(ctx context.Context, reportName string) ([]byte, error) {
	rows, err := s.Report(ctx, reportName)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if len(rows) == 0 {
		_ = w.Write([]string{"empty"})
		w.Flush()
		return buf.Bytes(), w.Error()
	}
	headers := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		headers = append(headers, k)
	}
	_ = w.Write(headers)
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = fmt.Sprint(row[h])
		}
		_ = w.Write(record)
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func (s *Services) ImportProductsCSV(ctx context.Context, body io.Reader) (int, error) {
	r := csv.NewReader(body)
	r.TrimLeadingSpace = true
	headers, err := r.Read()
	if err != nil {
		return 0, err
	}
	index := map[string]int{}
	for i, h := range headers {
		index[strings.ToLower(strings.TrimSpace(h))] = i
	}
	get := func(row []string, key string) string {
		if i, ok := index[key]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	count := 0
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return count, err
		}
		price, _ := strconv.ParseFloat(get(row, "price"), 64)
		cost, _ := strconv.ParseFloat(get(row, "cost"), 64)
		threshold, _ := strconv.Atoi(get(row, "threshold"))
		reorder, _ := strconv.Atoi(get(row, "reorder_point"))
		barcode := get(row, "barcode")
		var barcodePtr *string
		if barcode != "" {
			barcodePtr = &barcode
		}
		_, err = s.SaveProduct(ctx, 0, models.ProductInput{
			SKU: get(row, "sku"), Barcode: barcodePtr, Name: get(row, "name"), Unit: get(row, "unit"),
			Price: price, Cost: cost, Threshold: threshold, ReorderPoint: reorder, Active: true,
		})
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Services) Settings(ctx context.Context) ([]models.Setting, error) {
	return s.repo.Settings(ctx)
}

func (s *Services) SetSetting(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.New("setting key is required")
	}
	return s.repo.SetSetting(ctx, key, value)
}

func (s *Services) notifyLine(ctx context.Context, message string) error {
	if s.cfg.LineToken == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/v2/bot/message/broadcast", strings.NewReader(`{"messages":[{"type":"text","text":`+strconv.Quote(message)+`}]}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.LineToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}
