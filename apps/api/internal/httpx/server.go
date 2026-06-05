package httpx

import (
	"database/sql"
	"net/http"
	"os"
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
	mux.HandleFunc("GET /api/v1/profile", s.auth(s.profile))
	mux.HandleFunc("PATCH /api/v1/profile", s.auth(s.profile))
	mux.HandleFunc("POST /api/v1/profile/avatar", s.auth(s.uploadAvatar))
	mux.HandleFunc("DELETE /api/v1/profile/avatar", s.auth(s.deleteAvatar))
	mux.HandleFunc("GET /api/v1/users", s.requirePermission(s.users, "users.view"))
	mux.HandleFunc("POST /api/v1/users", s.requirePermission(s.users, "users.create"))
	mux.HandleFunc("GET /api/v1/users/{id}", s.requirePermission(s.userDetail, "users.view"))
	mux.HandleFunc("PATCH /api/v1/users/{id}", s.requirePermission(s.userDetail, "users.update"))
	mux.HandleFunc("PATCH /api/v1/users/{id}/status", s.requirePermission(s.userStatus, "users.deactivate"))
	mux.HandleFunc("GET /api/v1/users/{id}/roles", s.requirePermission(s.userRoles, "users.view"))
	mux.HandleFunc("PUT /api/v1/users/{id}/roles", s.requirePermission(s.userRoles, "users.assign_roles"))
	mux.HandleFunc("GET /api/v1/roles", s.requirePermission(s.roles, "roles.view"))
	mux.HandleFunc("POST /api/v1/roles", s.requirePermission(s.roles, "roles.create"))
	mux.HandleFunc("GET /api/v1/roles/{id}", s.requirePermission(s.roleDetail, "roles.view"))
	mux.HandleFunc("PATCH /api/v1/roles/{id}", s.requirePermission(s.roleDetail, "roles.update"))
	mux.HandleFunc("PATCH /api/v1/roles/{id}/status", s.requirePermission(s.roleStatus, "roles.deactivate"))
	mux.HandleFunc("DELETE /api/v1/roles/{id}", s.requirePermission(s.deleteRole, "roles.deactivate"))
	mux.HandleFunc("GET /api/v1/roles/{id}/permissions", s.requirePermission(s.rolePermissions, "roles.view"))
	mux.HandleFunc("PUT /api/v1/roles/{id}/permissions", s.requirePermission(s.rolePermissions, "roles.assign_permissions"))
	mux.HandleFunc("GET /api/v1/permissions", s.requirePermission(s.permissions, "permissions.view"))
	mux.HandleFunc("POST /api/v1/permissions", s.requirePermission(s.permissions, "roles.assign_permissions"))
	mux.HandleFunc("PATCH /api/v1/permissions/{id}", s.requirePermission(s.permissionDetail, "roles.assign_permissions"))
	mux.HandleFunc("GET /api/v1/permissions/grouped", s.requirePermission(s.groupedPermissions, "permissions.view"))
	mux.HandleFunc("GET /api/v1/categories", s.requirePermission(s.categories, "categories.view"))
	mux.HandleFunc("POST /api/v1/categories", s.requirePermission(s.categories, "categories.create"))
	mux.HandleFunc("PATCH /api/v1/categories/{id}", s.requirePermission(s.categoryDetail, "categories.update"))
	mux.HandleFunc("PATCH /api/v1/categories/{id}/status", s.requirePermission(s.categoryStatus, "categories.deactivate"))
	mux.HandleFunc("GET /api/v1/products", s.requirePermission(s.products, "products.view"))
	mux.HandleFunc("POST /api/v1/products", s.requirePermission(s.products, "products.create"))
	mux.HandleFunc("GET /api/v1/products/{id}", s.requirePermission(s.productDetail, "products.view"))
	mux.HandleFunc("PATCH /api/v1/products/{id}", s.requirePermission(s.productDetail, "products.update"))
	mux.HandleFunc("POST /api/v1/products/{id}/image", s.requirePermission(s.productImage, "products.update"))
	mux.HandleFunc("DELETE /api/v1/products/{id}/image", s.requirePermission(s.productImage, "products.update"))
	mux.HandleFunc("PATCH /api/v1/products/{id}/status", s.requirePermission(s.productStatus, "products.deactivate"))
	mux.HandleFunc("GET /api/v1/products/{id}/stocks", s.requirePermission(s.productStocksByProduct, "products.view"))
	mux.HandleFunc("POST /api/v1/products/{id}/restock", s.requirePermission(s.restockProduct, "stock.restock"))
	mux.HandleFunc("POST /api/v1/products/{id}/adjust-stock", s.requirePermission(s.adjustProductStock, "stock.adjust"))
	mux.HandleFunc("GET /api/v1/locations", s.requirePermission(s.locations, "locations.view"))
	mux.HandleFunc("POST /api/v1/locations", s.requirePermission(s.locations, "locations.create"))
	mux.HandleFunc("PATCH /api/v1/locations/{id}", s.requirePermission(s.locationDetail, "locations.update"))
	mux.HandleFunc("PATCH /api/v1/locations/{id}/status", s.requirePermission(s.locationStatus, "locations.deactivate"))
	mux.HandleFunc("GET /api/v1/product-stocks", s.requirePermission(s.productStocks, "stock.view"))
	mux.HandleFunc("GET /api/v1/stock-operations/options", s.requireAnyPermission(s.stockOperationOptions, "stock.restock", "stock.adjust"))
	mux.HandleFunc("GET /api/v1/stock-movements", s.requirePermission(s.stockMovements, "stock.movements.view"))
	mux.HandleFunc("GET /api/v1/stock-transfers/options", s.requireAnyPermission(s.stockTransferOptions, "transfers.view", "transfers.create"))
	mux.HandleFunc("GET /api/v1/stock-transfers", s.requirePermission(s.stockTransfers, "transfers.view"))
	mux.HandleFunc("POST /api/v1/stock-transfers", s.requirePermission(s.stockTransfers, "transfers.create"))
	mux.HandleFunc("GET /api/v1/stock-transfers/{id}", s.requirePermission(s.stockTransferDetail, "transfers.view"))
	mux.HandleFunc("POST /api/v1/stock-transfers/{id}/complete", s.requirePermission(s.completeStockTransfer, "transfers.complete"))
	mux.HandleFunc("POST /api/v1/stock-transfers/{id}/cancel", s.requirePermission(s.cancelStockTransfer, "transfers.cancel"))
	mux.HandleFunc("GET /api/v1/pos/products", s.requirePermission(s.posProducts, "pos.view"))
	mux.HandleFunc("GET /api/v1/sales", s.requirePermission(s.sales, "sales.view"))
	mux.HandleFunc("POST /api/v1/sales", s.requirePermission(s.sales, "pos.sell"))
	mux.HandleFunc("GET /api/v1/sales/{id}", s.requirePermission(s.saleDetail, "sales.view"))
	mux.HandleFunc("GET /api/v1/sales/{id}/receipt", s.requirePermission(s.saleReceipt, "sales.receipt.view"))
	mux.HandleFunc("POST /api/v1/sales/{id}/cancel", s.requirePermission(s.cancelSale, "sales.cancel"))
	mux.HandleFunc("GET /api/v1/alerts", s.requirePermission(s.alerts, "alerts.view"))
	mux.HandleFunc("PATCH /api/v1/alerts/{id}/read", s.requirePermission(s.readAlert, "alerts.mark_read"))
	mux.HandleFunc("PATCH /api/v1/alerts/read-all", s.requirePermission(s.readAllAlerts, "alerts.mark_read"))
	mux.HandleFunc("GET /api/v1/dashboard/summary", s.requirePermission(s.dashboardSummary, "dashboard.view"))
	mux.HandleFunc("GET /api/v1/reports/daily-sales", s.requirePermission(s.reportDailySales, "reports.daily_sales"))
	mux.HandleFunc("GET /api/v1/reports/monthly-sales", s.requirePermission(s.reportMonthlySales, "reports.monthly_sales"))
	mux.HandleFunc("GET /api/v1/reports/best-selling", s.requirePermission(s.reportBestSelling, "reports.best_selling"))
	mux.HandleFunc("GET /api/v1/reports/profit-by-product", s.requirePermission(s.reportProfitByProduct, "reports.profit"))
	mux.HandleFunc("GET /api/v1/reports/stock", s.requirePermission(s.reportStock, "reports.stock"))
	mux.HandleFunc("GET /api/v1/reports/inventory-valuation", s.requirePermission(s.reportInventoryValuation, "reports.inventory_valuation"))
	mux.HandleFunc("GET /api/v1/reports/payment-summary", s.requirePermission(s.reportPaymentSummary, "reports.payment_summary"))
	mux.HandleFunc("GET /api/v1/reports/low-stock", s.requirePermission(s.reportLowStock, "reports.low_stock"))
	mux.HandleFunc("GET /api/v1/reports/reorder", s.requirePermission(s.reportReorder, "reports.reorder"))
	mux.HandleFunc("GET /api/v1/exports/inventory-monthly", s.requirePermission(s.exportInventoryMonthly, "exports.inventory"))
	mux.HandleFunc("GET /api/v1/exports/products", s.requirePermission(s.exportProducts, "exports.products"))
	mux.HandleFunc("GET /api/v1/exports/sales", s.requirePermission(s.exportSales, "exports.sales"))
	mux.HandleFunc("GET /api/v1/exports/profit", s.requirePermission(s.exportProfit, "exports.profit"))
	mux.HandleFunc("GET /api/v1/imports/products/template", s.requirePermission(s.productImportTemplate, "imports.template.download"))
	mux.HandleFunc("POST /api/v1/imports/products/preview", s.requirePermission(s.productImportPreview, "imports.products.preview"))
	mux.HandleFunc("POST /api/v1/imports/products/confirm", s.requirePermission(s.productImportConfirm, "imports.products.confirm"))
	mux.HandleFunc("GET /api/v1/imports", s.requirePermission(s.imports, "imports.view"))
	mux.HandleFunc("GET /api/v1/imports/{id}", s.requirePermission(s.importDetail, "imports.history.view"))
	mux.HandleFunc("GET /api/v1/suppliers", s.requirePermission(s.suppliers, "suppliers.view"))
	mux.HandleFunc("POST /api/v1/suppliers", s.requirePermission(s.suppliers, "suppliers.create"))
	mux.HandleFunc("GET /api/v1/suppliers/{id}", s.requirePermission(s.supplierDetail, "suppliers.view"))
	mux.HandleFunc("PATCH /api/v1/suppliers/{id}", s.requirePermission(s.supplierDetail, "suppliers.update"))
	mux.HandleFunc("PATCH /api/v1/suppliers/{id}/status", s.requirePermission(s.supplierStatus, "suppliers.deactivate"))
	mux.HandleFunc("GET /api/v1/purchase-orders", s.requirePermission(s.purchaseOrders, "purchase_orders.view"))
	mux.HandleFunc("POST /api/v1/purchase-orders", s.requirePermission(s.purchaseOrders, "purchase_orders.create"))
	mux.HandleFunc("GET /api/v1/purchase-orders/{id}", s.requirePermission(s.purchaseOrderDetail, "purchase_orders.view"))
	mux.HandleFunc("PATCH /api/v1/purchase-orders/{id}", s.requirePermission(s.purchaseOrderDetail, "purchase_orders.update"))
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/send", s.requirePermission(s.sendPurchaseOrder, "purchase_orders.send"))
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/receive", s.requirePermission(s.receivePurchaseOrder, "purchase_orders.receive"))
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/cancel", s.requirePermission(s.cancelPurchaseOrder, "purchase_orders.cancel"))
	mux.HandleFunc("GET /api/v1/settings", s.requirePermission(s.settings, "settings.view"))
	mux.HandleFunc("PATCH /api/v1/settings", s.requirePermission(s.settings, "settings.update"))
	mux.HandleFunc("GET /api/v1/settings/line", s.requirePermission(s.lineSettings, "settings.line.view"))
	mux.HandleFunc("PATCH /api/v1/settings/line", s.requirePermission(s.lineSettings, "settings.line.update"))
	mux.HandleFunc("POST /api/v1/settings/line/test", s.requirePermission(s.testLineNotification, "settings.line.test"))
	mux.HandleFunc("GET /api/v1/notification-logs", s.requirePermission(s.notificationLogs, "notifications.view"))
	_ = os.MkdirAll(s.cfg.UploadDir, 0755)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.cfg.UploadDir))))
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
