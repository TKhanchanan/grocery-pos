package httpx

import (
	"database/sql"
	"net/http"
	"time"

	"grocery-pos/apps/api/internal/config"
	"grocery-pos/apps/api/internal/middleware"
	"grocery-pos/apps/api/internal/response"
)

type Server struct {
	cfg config.Config
	db  *sql.DB
}

func NewServer(cfg config.Config, db *sql.DB) *Server {
	return &Server{cfg: cfg, db: db}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/version", s.version)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/logout", s.auth(s.logout))
	mux.HandleFunc("GET /api/v1/auth/me", s.auth(s.me))
	mux.HandleFunc("/api/v1/users", s.requireRoles(s.users, RoleAdmin))
	mux.HandleFunc("GET /api/v1/users/{id}", s.requireRoles(s.userDetail, RoleAdmin))
	mux.HandleFunc("PATCH /api/v1/users/{id}", s.requireRoles(s.userDetail, RoleAdmin))
	mux.HandleFunc("PATCH /api/v1/users/{id}/status", s.requireRoles(s.userStatus, RoleAdmin))
	mux.HandleFunc("GET /api/v1/categories", s.auth(s.categories))
	mux.HandleFunc("POST /api/v1/categories", s.requireRoles(s.categories, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/categories/{id}", s.requireRoles(s.categoryDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/products", s.auth(s.products))
	mux.HandleFunc("POST /api/v1/products", s.requireRoles(s.products, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/products/{id}", s.auth(s.productDetail))
	mux.HandleFunc("PATCH /api/v1/products/{id}", s.requireRoles(s.productDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/products/{id}/status", s.requireRoles(s.productStatus, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/products/{id}/stocks", s.auth(s.productStocksByProduct))
	mux.HandleFunc("POST /api/v1/products/{id}/restock", s.requireRoles(s.restockProduct, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/products/{id}/adjust-stock", s.requireRoles(s.adjustProductStock, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/locations", s.auth(s.locations))
	mux.HandleFunc("POST /api/v1/locations", s.requireRoles(s.locations, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/locations/{id}", s.requireRoles(s.locationDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/locations/{id}/status", s.requireRoles(s.locationStatus, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/product-stocks", s.auth(s.productStocks))
	mux.HandleFunc("GET /api/v1/stock-movements", s.requireRoles(s.stockMovements, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/stock-transfers", s.requireRoles(s.stockTransfers, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/stock-transfers", s.requireRoles(s.stockTransfers, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/stock-transfers/{id}", s.requireRoles(s.stockTransferDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/stock-transfers/{id}/complete", s.requireRoles(s.completeStockTransfer, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/stock-transfers/{id}/cancel", s.requireRoles(s.cancelStockTransfer, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/pos/products", s.requireRoles(s.posProducts, RoleAdmin, RoleCashier))
	mux.HandleFunc("GET /api/v1/sales", s.requireRoles(s.sales, RoleAdmin, RoleManager, RoleCashier))
	mux.HandleFunc("POST /api/v1/sales", s.requireRoles(s.sales, RoleAdmin, RoleCashier))
	mux.HandleFunc("GET /api/v1/sales/{id}", s.requireRoles(s.saleDetail, RoleAdmin, RoleManager, RoleCashier))
	mux.HandleFunc("GET /api/v1/sales/{id}/receipt", s.requireRoles(s.saleReceipt, RoleAdmin, RoleManager, RoleCashier))
	mux.HandleFunc("POST /api/v1/sales/{id}/cancel", s.requireRoles(s.cancelSale, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/alerts", s.requireRoles(s.alerts, RoleAdmin, RoleManager, RoleCashier))
	mux.HandleFunc("PATCH /api/v1/alerts/{id}/read", s.requireRoles(s.readAlert, RoleAdmin, RoleManager, RoleCashier))
	mux.HandleFunc("PATCH /api/v1/alerts/read-all", s.requireRoles(s.readAllAlerts, RoleAdmin, RoleManager, RoleCashier))
	mux.HandleFunc("GET /api/v1/dashboard/summary", s.auth(s.dashboardSummary))
	mux.HandleFunc("GET /api/v1/reports/daily-sales", s.requireRoles(s.reportDailySales, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/monthly-sales", s.requireRoles(s.reportMonthlySales, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/best-selling", s.requireRoles(s.reportBestSelling, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/profit-by-product", s.requireRoles(s.reportProfitByProduct, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/stock", s.requireRoles(s.reportStock, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/inventory-valuation", s.requireRoles(s.reportInventoryValuation, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/payment-summary", s.requireRoles(s.reportPaymentSummary, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/low-stock", s.requireRoles(s.reportLowStock, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/reports/reorder", s.requireRoles(s.reportReorder, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/exports/inventory-monthly", s.requireRoles(s.exportInventoryMonthly, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/exports/products", s.requireRoles(s.exportProducts, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/exports/sales", s.requireRoles(s.exportSales, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/exports/profit", s.requireRoles(s.exportProfit, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/imports/products/template", s.requireRoles(s.productImportTemplate, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/imports/products/preview", s.requireRoles(s.productImportPreview, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/imports/products/confirm", s.requireRoles(s.productImportConfirm, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/imports", s.requireRoles(s.imports, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/imports/{id}", s.requireRoles(s.importDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/suppliers", s.requireRoles(s.suppliers, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/suppliers", s.requireRoles(s.suppliers, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/suppliers/{id}", s.requireRoles(s.supplierDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/suppliers/{id}", s.requireRoles(s.supplierDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/suppliers/{id}/status", s.requireRoles(s.supplierStatus, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/purchase-orders", s.requireRoles(s.purchaseOrders, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/purchase-orders", s.requireRoles(s.purchaseOrders, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/purchase-orders/{id}", s.requireRoles(s.purchaseOrderDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/purchase-orders/{id}", s.requireRoles(s.purchaseOrderDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/send", s.requireRoles(s.sendPurchaseOrder, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/receive", s.requireRoles(s.receivePurchaseOrder, RoleAdmin, RoleManager))
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/cancel", s.requireRoles(s.cancelPurchaseOrder, RoleAdmin, RoleManager))
	mux.HandleFunc("/", s.notFound)

	return middleware.Chain(
		mux,
		middleware.Recover,
		middleware.RequestLogger,
		middleware.CORS(s.cfg.CORSOrigins),
	)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "ok"
	dbStatus := "ok"

	if err := s.db.PingContext(ctx); err != nil {
		status = "degraded"
		dbStatus = "error"
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status":    status,
		"database":  dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"name":    s.cfg.AppName,
		"version": s.cfg.AppVersion,
		"env":     s.cfg.Env,
	})
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	response.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Route not found.")
}
