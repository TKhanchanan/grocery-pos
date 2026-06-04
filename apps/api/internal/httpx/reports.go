package httpx

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type DashboardSummary struct {
	TodaySales           float64                `json:"today_sales"`
	TodayReceipts        int                    `json:"today_receipts"`
	GrossProfitThisMonth float64                `json:"gross_profit_this_month"`
	TopProductThisMonth  *ProductSalesReport    `json:"top_product_this_month"`
	LowStockCount        int                    `json:"low_stock_count"`
	OutOfStockCount      int                    `json:"out_of_stock_count"`
	ReorderCount         int                    `json:"reorder_count"`
	PaymentSummary       []PaymentSummaryReport `json:"payment_method_summary"`
	RecentSales          []Receipt              `json:"recent_sales"`
	LowStockItems        []StockReport          `json:"low_stock_items"`
	TopProducts          []ProductSalesReport   `json:"top_products"`
}

type SalesPeriodReport struct {
	Period       string  `json:"period"`
	ReceiptCount int     `json:"receipt_count"`
	Revenue      float64 `json:"revenue"`
	Cost         float64 `json:"cost"`
	Profit       float64 `json:"profit"`
}

type ProductSalesReport struct {
	ProductID    uint64     `json:"product_id"`
	ProductName  string     `json:"product_name"`
	SKU          string     `json:"sku"`
	ImageURL     *string    `json:"image_url"`
	ImageUpdated *time.Time `json:"image_updated_at"`
	Quantity     int        `json:"quantity"`
	Revenue      float64    `json:"revenue"`
	Cost         float64    `json:"cost"`
	Profit       float64    `json:"profit"`
}

type StockReport struct {
	ProductID    uint64     `json:"product_id"`
	ProductName  string     `json:"product_name"`
	SKU          string     `json:"sku"`
	ImageURL     *string    `json:"image_url"`
	ImageUpdated *time.Time `json:"image_updated_at"`
	LocationID   uint64     `json:"location_id"`
	LocationName string     `json:"location_name"`
	Quantity     int        `json:"quantity"`
	UnitCost     float64    `json:"unit_cost"`
	TotalValue   float64    `json:"total_value"`
	Threshold    int        `json:"threshold"`
	ReorderPoint int        `json:"reorder_point"`
	StockStatus  string     `json:"stock_status"`
}

type PaymentSummaryReport struct {
	PaymentMethod string  `json:"payment_method"`
	ReceiptCount  int     `json:"receipt_count"`
	Revenue       float64 `json:"revenue"`
}

func (s *Server) dashboardSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.buildDashboardSummary(r.Context())
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load dashboard summary.")
		return
	}
	response.JSON(w, http.StatusOK, summary)
}

func (s *Server) reportDailySales(w http.ResponseWriter, r *http.Request) {
	rows, err := s.salesByPeriod(r.Context(), r, "DATE(s.created_at)")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportMonthlySales(w http.ResponseWriter, r *http.Request) {
	rows, err := s.salesByPeriod(r.Context(), r, "DATE_FORMAT(s.created_at, '%Y-%m')")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportBestSelling(w http.ResponseWriter, r *http.Request) {
	rows, err := s.productSalesReport(r.Context(), r, "SUM(si.quantity) DESC")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportProfitByProduct(w http.ResponseWriter, r *http.Request) {
	rows, err := s.productSalesReport(r.Context(), r, "profit DESC")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportStock(w http.ResponseWriter, r *http.Request) {
	rows, err := s.stockReport(r.Context(), r, "1=1")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportInventoryValuation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.stockReport(r.Context(), r, "1=1")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportPaymentSummary(w http.ResponseWriter, r *http.Request) {
	rows, err := s.paymentSummaryReport(r.Context(), r)
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportLowStock(w http.ResponseWriter, r *http.Request) {
	rows, err := s.stockReport(r.Context(), r, "p.threshold > 0 AND ps.quantity <= p.threshold")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) reportReorder(w http.ResponseWriter, r *http.Request) {
	rows, err := s.stockReport(r.Context(), r, "p.reorder_point > 0 AND ps.quantity <= p.reorder_point")
	if err != nil {
		reportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rows)
}

func (s *Server) buildDashboardSummary(ctx context.Context) (DashboardSummary, error) {
	var summary DashboardSummary
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM sales
		WHERE status='COMPLETED' AND DATE(created_at)=CURDATE()`).Scan(&summary.TodaySales, &summary.TodayReceipts); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(profit), 0)
		FROM sales
		WHERE status='COMPLETED' AND created_at >= DATE_FORMAT(CURDATE(), '%Y-%m-01')`).Scan(&summary.GrossProfitThisMonth); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE resolved_at IS NULL AND type='LOW_STOCK'`).Scan(&summary.LowStockCount); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE resolved_at IS NULL AND type='OUT_OF_STOCK'`).Scan(&summary.OutOfStockCount); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE resolved_at IS NULL AND type='REORDER_POINT'`).Scan(&summary.ReorderCount); err != nil {
		return DashboardSummary{}, err
	}

	topProducts, err := s.productSalesReportForWhere(ctx, []string{"s.status='COMPLETED'", "s.created_at >= DATE_FORMAT(CURDATE(), '%Y-%m-01')"}, nil, "SUM(si.quantity) DESC", 5)
	if err != nil {
		return DashboardSummary{}, err
	}
	summary.TopProducts = topProducts
	if len(topProducts) > 0 {
		summary.TopProductThisMonth = &topProducts[0]
	}
	summary.PaymentSummary, err = s.paymentSummaryForWhere(ctx, []string{"s.status='COMPLETED'", "DATE(s.created_at)=CURDATE()"}, nil)
	if err != nil {
		return DashboardSummary{}, err
	}
	summary.RecentSales, err = s.recentCompletedSales(ctx, 5)
	if err != nil {
		return DashboardSummary{}, err
	}
	summary.LowStockItems, err = s.stockReportForWhere(ctx, []string{"p.threshold > 0", "ps.quantity <= p.threshold"}, nil, 6)
	if err != nil {
		return DashboardSummary{}, err
	}
	return summary, nil
}

func (s *Server) salesByPeriod(ctx context.Context, r *http.Request, periodExpr string) ([]SalesPeriodReport, error) {
	where, args, err := reportSalesWhere(r)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+periodExpr+` AS period, COUNT(DISTINCT s.id), COALESCE(SUM(s.total_amount), 0),
		       COALESCE(SUM(s.total_cost), 0), COALESCE(SUM(s.profit), 0)
		FROM sales s
		WHERE `+strings.Join(where, " AND ")+`
		GROUP BY period
		ORDER BY period DESC
		LIMIT 120`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reports := []SalesPeriodReport{}
	for rows.Next() {
		var report SalesPeriodReport
		if err := rows.Scan(&report.Period, &report.ReceiptCount, &report.Revenue, &report.Cost, &report.Profit); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

func (s *Server) productSalesReport(ctx context.Context, r *http.Request, orderBy string) ([]ProductSalesReport, error) {
	where, args, err := reportSalesWhere(r)
	if err != nil {
		return nil, err
	}
	return s.productSalesReportForWhere(ctx, where, args, orderBy, 120)
}

func (s *Server) productSalesReportForWhere(ctx context.Context, where []string, args []any, orderBy string, limit int) ([]ProductSalesReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT si.product_id, si.product_name_snapshot, si.sku_snapshot, COALESCE(SUM(si.quantity), 0),
		       COALESCE(SUM(si.line_total), 0), COALESCE(SUM(si.line_cost), 0),
		       COALESCE(SUM(si.line_total - si.line_cost), 0) AS profit,
		       p.image_url, p.image_updated_at
		FROM sale_items si
		JOIN sales s ON s.id=si.sale_id
		LEFT JOIN products p ON p.id=si.product_id
		WHERE `+strings.Join(where, " AND ")+`
		GROUP BY si.product_id, si.product_name_snapshot, si.sku_snapshot, p.image_url, p.image_updated_at
		ORDER BY `+orderBy+`
		LIMIT ?`, append(args, limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reports := []ProductSalesReport{}
	for rows.Next() {
		var report ProductSalesReport
		var imageURL sql.NullString
		var imageUpdated sql.NullTime
		if err := rows.Scan(&report.ProductID, &report.ProductName, &report.SKU, &report.Quantity, &report.Revenue, &report.Cost, &report.Profit, &imageURL, &imageUpdated); err != nil {
			return nil, err
		}
		if imageURL.Valid {
			report.ImageURL = &imageURL.String
		}
		if imageUpdated.Valid {
			report.ImageUpdated = &imageUpdated.Time
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

func (s *Server) stockReport(ctx context.Context, r *http.Request, extraWhere string) ([]StockReport, error) {
	where, args, err := reportStockWhere(r)
	if err != nil {
		return nil, err
	}
	if extraWhere != "" {
		where = append(where, extraWhere)
	}
	return s.stockReportForWhere(ctx, where, args, 300)
}

func (s *Server) stockReportForWhere(ctx context.Context, where []string, args []any, limit int) ([]StockReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.sku, p.image_url, p.image_updated_at, l.id, l.name, ps.quantity, p.cost,
		       ps.quantity * p.cost, p.threshold, p.reorder_point
		FROM product_stocks ps
		JOIN products p ON p.id=ps.product_id
		JOIN locations l ON l.id=ps.location_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY p.name, l.name
		LIMIT ?`, append(args, limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reports := []StockReport{}
	for rows.Next() {
		var report StockReport
		var imageURL sql.NullString
		var imageUpdated sql.NullTime
		if err := rows.Scan(&report.ProductID, &report.ProductName, &report.SKU, &imageURL, &imageUpdated, &report.LocationID, &report.LocationName, &report.Quantity, &report.UnitCost, &report.TotalValue, &report.Threshold, &report.ReorderPoint); err != nil {
			return nil, err
		}
		if imageURL.Valid {
			report.ImageURL = &imageURL.String
		}
		if imageUpdated.Valid {
			report.ImageUpdated = &imageUpdated.Time
		}
		report.StockStatus = stockStatus(report.Quantity, report.Threshold, report.ReorderPoint)
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

func (s *Server) paymentSummaryReport(ctx context.Context, r *http.Request) ([]PaymentSummaryReport, error) {
	where, args, err := reportSalesWhere(r)
	if err != nil {
		return nil, err
	}
	return s.paymentSummaryForWhere(ctx, where, args)
}

func (s *Server) paymentSummaryForWhere(ctx context.Context, where []string, args []any) ([]PaymentSummaryReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.payment_method, COUNT(*), COALESCE(SUM(s.total_amount), 0) AS revenue
		FROM sales s
		WHERE `+strings.Join(where, " AND ")+`
		GROUP BY s.payment_method
		ORDER BY revenue DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reports := []PaymentSummaryReport{}
	for rows.Next() {
		var report PaymentSummaryReport
		if err := rows.Scan(&report.PaymentMethod, &report.ReceiptCount, &report.Revenue); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

func (s *Server) recentCompletedSales(ctx context.Context, limit int) ([]Receipt, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.receipt_no, s.location_id, l.name, s.cashier_id, u.full_name,
		       s.subtotal, s.total_amount, s.total_cost, s.profit, s.payment_method,
		       s.paid_amount, s.change_amount, s.status, s.cancelled_by, s.cancelled_at,
		       s.cancel_reason, s.created_at
		FROM sales s
		JOIN locations l ON l.id=s.location_id
		JOIN users u ON u.id=s.cashier_id
		WHERE s.status='COMPLETED'
		ORDER BY s.id DESC
		LIMIT ?`, limit)
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

func reportSalesWhere(r *http.Request) ([]string, []any, error) {
	where := []string{"s.status='COMPLETED'"}
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
	if value := strings.TrimSpace(query.Get("month")); value != "" {
		where = append(where, "DATE_FORMAT(s.created_at, '%Y-%m')=?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Get("location_id")); value != "" {
		id, err := parseFilterUint(value, "location_id")
		if err != nil {
			return nil, nil, err
		}
		where = append(where, "s.location_id=?")
		args = append(args, id)
	}
	return where, args, nil
}

func reportStockWhere(r *http.Request) ([]string, []any, error) {
	where := []string{"1=1"}
	args := []any{}
	if value := strings.TrimSpace(r.URL.Query().Get("location_id")); value != "" {
		id, err := parseFilterUint(value, "location_id")
		if err != nil {
			return nil, nil, err
		}
		where = append(where, "ps.location_id=?")
		args = append(args, id)
	}
	return where, args, nil
}

func parseFilterUint(value, name string) (uint64, error) {
	id, err := parseUintParam(value, name)
	if err != nil {
		return 0, errors.New("invalid " + name)
	}
	return id, nil
}

func reportError(w http.ResponseWriter, err error) {
	if strings.HasPrefix(err.Error(), "invalid ") {
		response.ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	response.ErrorJSON(w, http.StatusInternalServerError, "QUERY_FAILED", "Could not load report.")
}
