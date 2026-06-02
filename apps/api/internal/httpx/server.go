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
