package httpx

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

func (s *Server) exportInventoryMonthly(w http.ResponseWriter, r *http.Request) {
	if err := requireCSVFormat(r); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if month == "" {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "month is required.")
		return
	}
	rows, err := s.inventoryExportRows(r.Context(), month)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "EXPORT_FAILED", "Could not export inventory.")
		return
	}
	writeCSV(w, "inventory-monthly-"+month+".csv", rows)
}

func (s *Server) exportProducts(w http.ResponseWriter, r *http.Request) {
	if err := requireCSVFormat(r); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	rows, err := s.productExportRows(r.Context())
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "EXPORT_FAILED", "Could not export products.")
		return
	}
	writeCSV(w, "products.csv", rows)
}

func (s *Server) exportSales(w http.ResponseWriter, r *http.Request) {
	if err := requireCSVFormat(r); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	rows, err := s.salesExportRows(r.Context(), r)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid ") {
			response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "EXPORT_FAILED", "Could not export sales.")
		return
	}
	writeCSV(w, "sales.csv", rows)
}

func (s *Server) exportProfit(w http.ResponseWriter, r *http.Request) {
	if err := requireCSVFormat(r); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if month == "" {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "month is required.")
		return
	}
	rows, err := s.profitExportRows(r.Context(), r, month)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "EXPORT_FAILED", "Could not export profit.")
		return
	}
	writeCSV(w, "profit-"+month+".csv", rows)
}

func (s *Server) inventoryExportRows(ctx context.Context, month string) ([][]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.sku, p.name, COALESCE(c.name, ''), l.name, ps.quantity, p.cost, ps.quantity * p.cost,
		       p.threshold, p.reorder_point
		FROM product_stocks ps
		JOIN products p ON p.id=ps.product_id
		LEFT JOIN categories c ON c.id=p.category_id
		JOIN locations l ON l.id=ps.location_id
		ORDER BY p.name, l.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := [][]string{{"month", "sku", "product_name", "category", "location", "quantity", "unit_cost", "total_value", "threshold", "reorder_point"}}
	for rows.Next() {
		var sku, productName, categoryName, locationName string
		var quantity, threshold, reorderPoint int
		var unitCost, totalValue float64
		if err := rows.Scan(&sku, &productName, &categoryName, &locationName, &quantity, &unitCost, &totalValue, &threshold, &reorderPoint); err != nil {
			return nil, err
		}
		out = append(out, []string{thaiCSVMonth(month), sku, productName, categoryName, locationName, strconv.Itoa(quantity), moneyString(unitCost), moneyString(totalValue), strconv.Itoa(threshold), strconv.Itoa(reorderPoint)})
	}
	return out, rows.Err()
}

func (s *Server) productExportRows(ctx context.Context) ([][]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.sku, p.name, p.barcode, COALESCE(c.name, ''), p.unit, p.price, p.cost,
		       p.threshold, p.reorder_point, p.active, COALESCE(SUM(ps.quantity), 0)
		FROM products p
		LEFT JOIN categories c ON c.id=p.category_id
		LEFT JOIN product_stocks ps ON ps.product_id=p.id
		GROUP BY p.id, c.name
		ORDER BY p.name, p.sku`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := [][]string{{"sku", "product_name", "barcode", "category", "unit", "selling_price", "unit_cost", "threshold", "reorder_point", "is_active", "total_stock"}}
	for rows.Next() {
		var sku, productName, categoryName, unit string
		var barcode sql.NullString
		var active bool
		var threshold, reorderPoint, totalStock int
		var price, cost float64
		if err := rows.Scan(&sku, &productName, &barcode, &categoryName, &unit, &price, &cost, &threshold, &reorderPoint, &active, &totalStock); err != nil {
			return nil, err
		}
		out = append(out, []string{sku, productName, barcode.String, categoryName, unit, moneyString(price), moneyString(cost), strconv.Itoa(threshold), strconv.Itoa(reorderPoint), strconv.FormatBool(active), strconv.Itoa(totalStock)})
	}
	return out, rows.Err()
}

func (s *Server) salesExportRows(ctx context.Context, r *http.Request) ([][]string, error) {
	where, args, err := reportSalesWhere(r)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.receipt_no, s.created_at, l.name, u.full_name, s.payment_method,
		       s.total_amount, s.total_cost, s.profit
		FROM sales s
		JOIN locations l ON l.id=s.location_id
		JOIN users u ON u.id=s.cashier_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY s.created_at DESC, s.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := [][]string{{"receipt_no", "created_at", "location", "cashier", "payment_method", "revenue", "cost", "profit"}}
	for rows.Next() {
		var receiptNo, locationName, cashierName, paymentMethod string
		var createdAt time.Time
		var revenue, cost, profit float64
		if err := rows.Scan(&receiptNo, &createdAt, &locationName, &cashierName, &paymentMethod, &revenue, &cost, &profit); err != nil {
			return nil, err
		}
		out = append(out, []string{receiptNo, thaiCSVDateTime(createdAt), locationName, cashierName, paymentMethod, moneyString(revenue), moneyString(cost), moneyString(profit)})
	}
	return out, rows.Err()
}

func (s *Server) profitExportRows(ctx context.Context, r *http.Request, month string) ([][]string, error) {
	where := []string{"s.status='COMPLETED'", "DATE_FORMAT(s.created_at, '%Y-%m')=?"}
	args := []any{month}
	if value := strings.TrimSpace(r.URL.Query().Get("location_id")); value != "" {
		id, err := parseFilterUint(value, "location_id")
		if err != nil {
			return nil, err
		}
		where = append(where, "s.location_id=?")
		args = append(args, id)
	}
	reports, err := s.productSalesReportForWhere(ctx, where, args, "profit DESC", 1000)
	if err != nil {
		return nil, err
	}
	out := [][]string{{"month", "product_id", "product_name", "sku", "quantity", "revenue", "cost", "profit"}}
	for _, row := range reports {
		out = append(out, []string{thaiCSVMonth(month), strconv.FormatUint(row.ProductID, 10), row.ProductName, row.SKU, strconv.Itoa(row.Quantity), moneyString(row.Revenue), moneyString(row.Cost), moneyString(row.Profit)})
	}
	return out, nil
}

func requireCSVFormat(r *http.Request) error {
	format := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("format")))
	if format == "" || format == "csv" {
		return nil
	}
	return errors.New("only csv format is supported")
}

func writeCSV(w http.ResponseWriter, filename string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(w)
	_ = writer.WriteAll(rows)
	writer.Flush()
}

func moneyString(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func thaiCSVDate(value time.Time) string {
	return fmt.Sprintf("%02d/%02d/%d", value.Day(), value.Month(), value.Year()+543)
}

func thaiCSVDateTime(value time.Time) string {
	return fmt.Sprintf("%s %02d:%02d:%02d", thaiCSVDate(value), value.Hour(), value.Minute(), value.Second())
}

func thaiCSVMonth(month string) string {
	value, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	return thaiCSVDate(value)
}
