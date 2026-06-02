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
	mux.HandleFunc("GET /api/v1/locations", s.auth(s.locations))
	mux.HandleFunc("POST /api/v1/locations", s.requireRoles(s.locations, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/locations/{id}", s.requireRoles(s.locationDetail, RoleAdmin, RoleManager))
	mux.HandleFunc("PATCH /api/v1/locations/{id}/status", s.requireRoles(s.locationStatus, RoleAdmin, RoleManager))
	mux.HandleFunc("GET /api/v1/product-stocks", s.auth(s.productStocks))
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
