package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"grocery-pos/backend/internal/models"
)

type Repository struct {
	DB *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) UserByUsername(ctx context.Context, username string) (models.User, string, error) {
	var u models.User
	var hash string
	err := r.DB.QueryRowContext(ctx, `SELECT id, username, full_name, role, active, created_at, password_hash FROM users WHERE username=?`, username).
		Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.Active, &u.CreatedAt, &hash)
	return u, hash, err
}

func (r *Repository) UserByID(ctx context.Context, id uint64) (models.User, error) {
	var u models.User
	err := r.DB.QueryRowContext(ctx, `SELECT id, username, full_name, role, active, created_at FROM users WHERE id=?`, id).
		Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.Active, &u.CreatedAt)
	return u, err
}

func (r *Repository) Users(ctx context.Context) ([]models.User, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, username, full_name, role, active, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertUser(ctx context.Context, id uint64, in models.UserInput, hash string) (uint64, error) {
	if id == 0 {
		res, err := r.DB.ExecContext(ctx, `INSERT INTO users(username, password_hash, full_name, role, active) VALUES (?, ?, ?, ?, ?)`, in.Username, hash, in.FullName, in.Role, in.Active)
		if err != nil {
			return 0, err
		}
		newID, _ := res.LastInsertId()
		return uint64(newID), nil
	}
	if hash != "" {
		_, err := r.DB.ExecContext(ctx, `UPDATE users SET username=?, password_hash=?, full_name=?, role=?, active=? WHERE id=?`, in.Username, hash, in.FullName, in.Role, in.Active, id)
		return id, err
	}
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET username=?, full_name=?, role=?, active=? WHERE id=?`, in.Username, in.FullName, in.Role, in.Active, id)
	return id, err
}

func (r *Repository) Categories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertCategory(ctx context.Context, id uint64, name string) (uint64, error) {
	if id == 0 {
		res, err := r.DB.ExecContext(ctx, `INSERT INTO categories(name) VALUES (?)`, name)
		if err != nil {
			return 0, err
		}
		newID, _ := res.LastInsertId()
		return uint64(newID), nil
	}
	_, err := r.DB.ExecContext(ctx, `UPDATE categories SET name=? WHERE id=?`, name, id)
	return id, err
}

func (r *Repository) Locations(ctx context.Context) ([]models.Location, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, active, created_at FROM locations ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Location
	for rows.Next() {
		var l models.Location
		if err := rows.Scan(&l.ID, &l.Name, &l.Active, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertLocation(ctx context.Context, id uint64, name string, active bool) (uint64, error) {
	if id == 0 {
		res, err := r.DB.ExecContext(ctx, `INSERT INTO locations(name, active) VALUES (?, ?)`, name, active)
		if err != nil {
			return 0, err
		}
		newID, _ := res.LastInsertId()
		return uint64(newID), nil
	}
	_, err := r.DB.ExecContext(ctx, `UPDATE locations SET name=?, active=? WHERE id=?`, name, active, id)
	return id, err
}

func (r *Repository) Products(ctx context.Context, q string) ([]models.Product, error) {
	where := ""
	args := []any{}
	if q != "" {
		where = `WHERE p.name LIKE ? OR p.sku LIKE ? OR p.barcode LIKE ?`
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.id, p.category_id, p.sku, p.barcode, p.name, p.unit, p.price, p.cost,
		       p.threshold, p.reorder_point, p.active, p.created_at,
		       COALESCE(SUM(ps.quantity), 0) AS total_stock
		FROM products p
		LEFT JOIN product_stocks ps ON ps.product_id = p.id
		`+where+`
		GROUP BY p.id
		ORDER BY p.name, p.sku`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.SKU, &p.Barcode, &p.Name, &p.Unit, &p.Price, &p.Cost, &p.Threshold, &p.ReorderPoint, &p.Active, &p.CreatedAt, &p.TotalStock); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) ProductByID(ctx context.Context, id uint64) (models.Product, error) {
	var p models.Product
	err := r.DB.QueryRowContext(ctx, `
		SELECT p.id, p.category_id, p.sku, p.barcode, p.name, p.unit, p.price, p.cost,
		       p.threshold, p.reorder_point, p.active, p.created_at,
		       COALESCE(SUM(ps.quantity), 0)
		FROM products p
		LEFT JOIN product_stocks ps ON ps.product_id = p.id
		WHERE p.id=?
		GROUP BY p.id`, id).
		Scan(&p.ID, &p.CategoryID, &p.SKU, &p.Barcode, &p.Name, &p.Unit, &p.Price, &p.Cost, &p.Threshold, &p.ReorderPoint, &p.Active, &p.CreatedAt, &p.TotalStock)
	return p, err
}

func (r *Repository) UpsertProduct(ctx context.Context, id uint64, in models.ProductInput) (uint64, error) {
	if in.Unit == "" {
		in.Unit = "ชิ้น"
	}
	if id == 0 {
		res, err := r.DB.ExecContext(ctx, `
			INSERT INTO products(category_id, sku, barcode, name, unit, price, cost, threshold, reorder_point, active)
			VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?)`,
			in.CategoryID, in.SKU, nullableString(in.Barcode), in.Name, in.Unit, in.Price, in.Cost, in.Threshold, in.ReorderPoint, in.Active)
		if err != nil {
			return 0, err
		}
		newID, _ := res.LastInsertId()
		return uint64(newID), nil
	}
	_, err := r.DB.ExecContext(ctx, `
		UPDATE products SET category_id=?, sku=?, barcode=NULLIF(?, ''), name=?, unit=?, price=?, cost=?, threshold=?, reorder_point=?, active=?
		WHERE id=?`,
		in.CategoryID, in.SKU, nullableString(in.Barcode), in.Name, in.Unit, in.Price, in.Cost, in.Threshold, in.ReorderPoint, in.Active, id)
	return id, err
}

func nullableString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func (r *Repository) Stocks(ctx context.Context) ([]models.ProductStock, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.id, l.id, p.name, p.sku, l.name, COALESCE(ps.quantity, 0)
		FROM products p
		CROSS JOIN locations l
		LEFT JOIN product_stocks ps ON ps.product_id=p.id AND ps.location_id=l.id
		WHERE p.active=1 AND l.active=1
		ORDER BY l.id, p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ProductStock
	for rows.Next() {
		var s models.ProductStock
		if err := rows.Scan(&s.ProductID, &s.LocationID, &s.ProductName, &s.SKU, &s.LocationName, &s.Quantity); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func EnsureStockRow(ctx context.Context, tx *sql.Tx, productID, locationID uint64) error {
	_, err := tx.ExecContext(ctx, `INSERT IGNORE INTO product_stocks(product_id, location_id, quantity) VALUES (?, ?, 0)`, productID, locationID)
	return err
}

func LockedStock(ctx context.Context, tx *sql.Tx, productID, locationID uint64) (int, error) {
	if err := EnsureStockRow(ctx, tx, productID, locationID); err != nil {
		return 0, err
	}
	var qty int
	err := tx.QueryRowContext(ctx, `SELECT quantity FROM product_stocks WHERE product_id=? AND location_id=? FOR UPDATE`, productID, locationID).Scan(&qty)
	return qty, err
}

func ChangeStock(ctx context.Context, tx *sql.Tx, productID, locationID uint64, delta int, unitCost *float64, refType string, refID *uint64, note string, userID *uint64) error {
	qty, err := LockedStock(ctx, tx, productID, locationID)
	if err != nil {
		return err
	}
	if qty+delta < 0 {
		return fmt.Errorf("stock cannot become negative")
	}
	_, err = tx.ExecContext(ctx, `UPDATE product_stocks SET quantity=? WHERE product_id=? AND location_id=?`, qty+delta, productID, locationID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_movements(product_id, location_id, reference_type, reference_id, quantity_change, unit_cost, note, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, productID, locationID, refType, refID, delta, unitCost, note, userID)
	return err
}

func (r *Repository) Movements(ctx context.Context) ([]models.StockMovement, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT sm.id, sm.product_id, sm.location_id, sm.reference_type, sm.reference_id, sm.quantity_change,
		       sm.unit_cost, sm.note, sm.created_by, sm.created_at, p.name, l.name
		FROM stock_movements sm
		JOIN products p ON p.id=sm.product_id
		JOIN locations l ON l.id=sm.location_id
		ORDER BY sm.id DESC
		LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.StockMovement
	for rows.Next() {
		var m models.StockMovement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.LocationID, &m.ReferenceType, &m.ReferenceID, &m.QuantityChange, &m.UnitCost, &m.Note, &m.CreatedBy, &m.CreatedAt, &m.ProductName, &m.LocationName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) Alerts(ctx context.Context) ([]models.Alert, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT a.id, a.product_id, a.location_id, a.type, a.message, a.resolved_at, a.created_at,
		       p.name, l.name, ps.quantity, p.threshold, p.reorder_point
		FROM alerts a
		JOIN products p ON p.id=a.product_id
		JOIN locations l ON l.id=a.location_id
		JOIN product_stocks ps ON ps.product_id=a.product_id AND ps.location_id=a.location_id
		WHERE a.resolved_at IS NULL
		ORDER BY FIELD(a.type, 'OUT_OF_STOCK', 'REORDER_POINT', 'LOW_STOCK'), a.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.ID, &a.ProductID, &a.LocationID, &a.Type, &a.Message, &a.ResolvedAt, &a.CreatedAt, &a.ProductName, &a.LocationName, &a.CurrentStock, &a.Threshold, &a.ReorderPoint); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func UpsertAlertsForStock(ctx context.Context, tx *sql.Tx, productID, locationID uint64) error {
	var name string
	var qty, threshold, reorderPoint int
	err := tx.QueryRowContext(ctx, `
		SELECT p.name, ps.quantity, p.threshold, p.reorder_point
		FROM products p
		JOIN product_stocks ps ON ps.product_id=p.id
		WHERE p.id=? AND ps.location_id=?`, productID, locationID).
		Scan(&name, &qty, &threshold, &reorderPoint)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE alerts SET resolved_at=NOW() WHERE product_id=? AND location_id=? AND resolved_at IS NULL`, productID, locationID)
	create := func(t string) error {
		msg := fmt.Sprintf("%s %s at location %d: current stock %d", t, name, locationID, qty)
		_, err := tx.ExecContext(ctx, `INSERT INTO alerts(product_id, location_id, type, message) VALUES (?, ?, ?, ?)`, productID, locationID, t, msg)
		return err
	}
	if qty == 0 {
		if err := create("OUT_OF_STOCK"); err != nil {
			return err
		}
	}
	if threshold > 0 && qty <= threshold {
		if err := create("LOW_STOCK"); err != nil {
			return err
		}
	}
	if reorderPoint > 0 && qty <= reorderPoint {
		if err := create("REORDER_POINT"); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) Sales(ctx context.Context) ([]models.Sale, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT s.id, s.receipt_no, s.location_id, l.name, s.cashier_id, s.total_amount, s.total_cost,
		       s.profit, s.payment_method, s.paid_amount, s.change_amount, s.status, s.cancelled_at, s.created_at
		FROM sales s
		JOIN locations l ON l.id=s.location_id
		ORDER BY s.id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Sale
	for rows.Next() {
		var s models.Sale
		if err := rows.Scan(&s.ID, &s.ReceiptNo, &s.LocationID, &s.LocationName, &s.CashierID, &s.TotalAmount, &s.TotalCost, &s.Profit, &s.PaymentMethod, &s.PaidAmount, &s.ChangeAmount, &s.Status, &s.CancelledAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) SaleItems(ctx context.Context, saleID uint64) ([]models.SaleItem, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, sale_id, product_id, product_name_snapshot, sku_snapshot, price_snapshot, cost_snapshot, quantity, line_total, line_cost
		FROM sale_items WHERE sale_id=? ORDER BY id`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SaleItem
	for rows.Next() {
		var item models.SaleItem
		if err := rows.Scan(&item.ID, &item.SaleID, &item.ProductID, &item.ProductNameSnapshot, &item.SKUSnapshot, &item.PriceSnapshot, &item.CostSnapshot, &item.Quantity, &item.LineTotal, &item.LineCost); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) Suppliers(ctx context.Context) ([]models.Supplier, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, phone, email, address, created_at FROM suppliers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Supplier
	for rows.Next() {
		var s models.Supplier
		if err := rows.Scan(&s.ID, &s.Name, &s.Phone, &s.Email, &s.Address, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertSupplier(ctx context.Context, s models.Supplier) (uint64, error) {
	if s.ID == 0 {
		res, err := r.DB.ExecContext(ctx, `INSERT INTO suppliers(name, phone, email, address) VALUES (?, ?, ?, ?)`, s.Name, s.Phone, s.Email, s.Address)
		if err != nil {
			return 0, err
		}
		id, _ := res.LastInsertId()
		return uint64(id), nil
	}
	_, err := r.DB.ExecContext(ctx, `UPDATE suppliers SET name=?, phone=?, email=?, address=? WHERE id=?`, s.Name, s.Phone, s.Email, s.Address, s.ID)
	return s.ID, err
}

func (r *Repository) PurchaseOrders(ctx context.Context) ([]models.PurchaseOrder, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT po.id, po.po_number, po.supplier_id, s.name, po.location_id, l.name, po.status, po.total_cost, po.created_at, po.received_at
		FROM purchase_orders po
		JOIN suppliers s ON s.id=po.supplier_id
		JOIN locations l ON l.id=po.location_id
		ORDER BY po.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.PurchaseOrder
	for rows.Next() {
		var po models.PurchaseOrder
		if err := rows.Scan(&po.ID, &po.PONumber, &po.SupplierID, &po.Supplier, &po.LocationID, &po.Location, &po.Status, &po.TotalCost, &po.CreatedAt, &po.ReceivedAt); err != nil {
			return nil, err
		}
		out = append(out, po)
	}
	return out, rows.Err()
}

func (r *Repository) Settings(ctx context.Context) ([]models.Setting, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT setting_key, setting_value FROM settings ORDER BY setting_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Setting
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.Key, &s.Value); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) SetSetting(ctx context.Context, key, value string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO settings(setting_key, setting_value) VALUES (?, ?) ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value)`, key, value)
	return err
}
