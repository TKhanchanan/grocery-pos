package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"

	"grocery-pos/backend/internal/config"
	"grocery-pos/backend/internal/models"
	"grocery-pos/backend/internal/service"
)

type Server struct {
	svc *service.Services
	cfg config.Config
}

func NewServer(svc *service.Services, cfg config.Config) *Server {
	return &Server{svc: svc, cfg: cfg}
}

type routeHandler func(nethttp.ResponseWriter, *nethttp.Request, models.User) error

func (s *Server) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/api/health", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/auth/login", s.public(s.login))
	mux.HandleFunc("/api/me", s.auth(s.me))
	mux.HandleFunc("/api/users", s.auth(s.users))
	mux.HandleFunc("/api/dashboard", s.auth(s.dashboard))
	mux.HandleFunc("/api/categories", s.auth(s.categories))
	mux.HandleFunc("/api/products", s.auth(s.products))
	mux.HandleFunc("/api/products/import", s.auth(s.importProducts))
	mux.HandleFunc("/api/locations", s.auth(s.locations))
	mux.HandleFunc("/api/stocks", s.auth(s.stocks))
	mux.HandleFunc("/api/stock/restock", s.auth(s.restock))
	mux.HandleFunc("/api/stock/adjust", s.auth(s.adjust))
	mux.HandleFunc("/api/stock/transfer", s.auth(s.transfer))
	mux.HandleFunc("/api/stock/movements", s.auth(s.movements))
	mux.HandleFunc("/api/alerts", s.auth(s.alerts))
	mux.HandleFunc("/api/sales", s.auth(s.sales))
	mux.HandleFunc("/api/sales/", s.auth(s.saleByID))
	mux.HandleFunc("/api/reports/", s.auth(s.report))
	mux.HandleFunc("/api/export/", s.auth(s.export))
	mux.HandleFunc("/api/suppliers", s.auth(s.suppliers))
	mux.HandleFunc("/api/purchase-orders", s.auth(s.purchaseOrders))
	mux.HandleFunc("/api/purchase-orders/", s.auth(s.purchaseOrderByID))
	mux.HandleFunc("/api/settings", s.auth(s.settings))
	return cors(mux)
}

func cors(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == nethttp.MethodOptions {
			w.WriteHeader(nethttp.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) public(fn func(nethttp.ResponseWriter, *nethttp.Request) error) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if err := fn(w, r); err != nil {
			writeError(w, err)
		}
	}
}

func (s *Server) auth(fn routeHandler) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeJSON(w, nethttp.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}
		user, err := s.svc.UserFromToken(r.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			writeJSON(w, nethttp.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		if err := fn(w, r, user); err != nil {
			writeError(w, err)
		}
	}
}

func (s *Server) login(w nethttp.ResponseWriter, r *nethttp.Request) error {
	if r.Method != nethttp.MethodPost {
		return method()
	}
	var in models.LoginRequest
	if err := readJSON(r, &in); err != nil {
		return err
	}
	out, err := s.svc.Login(r.Context(), in)
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) me(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	writeJSON(w, nethttp.StatusOK, user)
	return nil
}

func (s *Server) users(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin); err != nil {
		return err
	}
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Users(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost, nethttp.MethodPut:
		var body struct {
			ID uint64 `json:"id"`
			models.UserInput
		}
		if err := readJSON(r, &body); err != nil {
			return err
		}
		id, err := s.svc.SaveUser(r.Context(), body.ID, body.UserInput)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) dashboard(w nethttp.ResponseWriter, r *nethttp.Request, _ models.User) error {
	out, err := s.svc.Dashboard(r.Context())
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) categories(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Categories(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost, nethttp.MethodPut:
		if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
			return err
		}
		var body struct {
			ID   uint64 `json:"id"`
			Name string `json:"name"`
		}
		if err := readJSON(r, &body); err != nil {
			return err
		}
		id, err := s.svc.SaveCategory(r.Context(), body.ID, body.Name)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) products(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Products(r.Context(), r.URL.Query().Get("q"))
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost, nethttp.MethodPut:
		if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
			return err
		}
		var body struct {
			ID uint64 `json:"id"`
			models.ProductInput
		}
		if err := readJSON(r, &body); err != nil {
			return err
		}
		id, err := s.svc.SaveProduct(r.Context(), body.ID, body.ProductInput)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) importProducts(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
		return err
	}
	if r.Method != nethttp.MethodPost {
		return method()
	}
	count, err := s.svc.ImportProductsCSV(r.Context(), r.Body)
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, map[string]int{"imported": count})
	return nil
}

func (s *Server) locations(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Locations(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost, nethttp.MethodPut:
		if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
			return err
		}
		var body struct {
			ID     uint64 `json:"id"`
			Name   string `json:"name"`
			Active bool   `json:"active"`
		}
		if err := readJSON(r, &body); err != nil {
			return err
		}
		id, err := s.svc.SaveLocation(r.Context(), body.ID, body.Name, body.Active)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) stocks(w nethttp.ResponseWriter, r *nethttp.Request, _ models.User) error {
	out, err := s.svc.Stocks(r.Context())
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) restock(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
		return err
	}
	var in models.StockChangeRequest
	if err := readJSON(r, &in); err != nil {
		return err
	}
	if err := s.svc.Restock(r.Context(), user, in); err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, map[string]bool{"ok": true})
	return nil
}

func (s *Server) adjust(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
		return err
	}
	var in models.StockChangeRequest
	if err := readJSON(r, &in); err != nil {
		return err
	}
	if err := s.svc.AdjustStock(r.Context(), user, in); err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, map[string]bool{"ok": true})
	return nil
}

func (s *Server) transfer(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
		return err
	}
	var in models.StockTransferRequest
	if err := readJSON(r, &in); err != nil {
		return err
	}
	if err := s.svc.TransferStock(r.Context(), user, in); err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, map[string]bool{"ok": true})
	return nil
}

func (s *Server) movements(w nethttp.ResponseWriter, r *nethttp.Request, _ models.User) error {
	out, err := s.svc.Movements(r.Context())
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) alerts(w nethttp.ResponseWriter, r *nethttp.Request, _ models.User) error {
	out, err := s.svc.Alerts(r.Context())
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) sales(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Sales(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost:
		var in models.CreateSaleRequest
		if err := readJSON(r, &in); err != nil {
			return err
		}
		id, err := s.svc.CreateSale(r.Context(), user, in)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) saleByID(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	id, suffix, err := idFromPath(r.URL.Path, "/api/sales/")
	if err != nil {
		return err
	}
	if suffix == "/cancel" {
		if r.Method != nethttp.MethodPost {
			return method()
		}
		if err := s.svc.CancelSale(r.Context(), user, id); err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]bool{"ok": true})
		return nil
	}
	out, err := s.svc.Sale(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) report(w nethttp.ResponseWriter, r *nethttp.Request, _ models.User) error {
	name := strings.TrimPrefix(r.URL.Path, "/api/reports/")
	out, err := s.svc.Report(r.Context(), name)
	if err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, out)
	return nil
}

func (s *Server) export(w nethttp.ResponseWriter, r *nethttp.Request, _ models.User) error {
	name := strings.TrimPrefix(r.URL.Path, "/api/export/")
	data, err := s.svc.ExportCSV(r.Context(), name)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+name+".csv")
	_, _ = w.Write(data)
	return nil
}

func (s *Server) suppliers(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Suppliers(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost, nethttp.MethodPut:
		if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
			return err
		}
		var in models.Supplier
		if err := readJSON(r, &in); err != nil {
			return err
		}
		id, err := s.svc.SaveSupplier(r.Context(), in)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) purchaseOrders(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.PurchaseOrders(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost:
		if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
			return err
		}
		var in models.PurchaseOrderInput
		if err := readJSON(r, &in); err != nil {
			return err
		}
		id, err := s.svc.CreatePurchaseOrder(r.Context(), in)
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]uint64{"id": id})
	default:
		return method()
	}
	return nil
}

func (s *Server) purchaseOrderByID(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin, models.RoleManager); err != nil {
		return err
	}
	id, suffix, err := idFromPath(r.URL.Path, "/api/purchase-orders/")
	if err != nil {
		return err
	}
	if suffix != "/receive" || r.Method != nethttp.MethodPost {
		return method()
	}
	if err := s.svc.ReceivePurchaseOrder(r.Context(), user, id); err != nil {
		return err
	}
	writeJSON(w, nethttp.StatusOK, map[string]bool{"ok": true})
	return nil
}

func (s *Server) settings(w nethttp.ResponseWriter, r *nethttp.Request, user models.User) error {
	if err := require(user, models.RoleAdmin); err != nil {
		return err
	}
	switch r.Method {
	case nethttp.MethodGet:
		out, err := s.svc.Settings(r.Context())
		if err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, out)
	case nethttp.MethodPost, nethttp.MethodPut:
		var in models.Setting
		if err := readJSON(r, &in); err != nil {
			return err
		}
		if err := s.svc.SetSetting(r.Context(), in.Key, in.Value); err != nil {
			return err
		}
		writeJSON(w, nethttp.StatusOK, map[string]bool{"ok": true})
	default:
		return method()
	}
	return nil
}

func require(user models.User, roles ...models.Role) error {
	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}
	return errors.New("permission denied")
}

func idFromPath(path, prefix string) (uint64, string, error) {
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(rest, "/", 2)
	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, "", errors.New("invalid id")
	}
	suffix := ""
	if len(parts) == 2 {
		suffix = "/" + parts[1]
	}
	return id, suffix, nil
}

func readJSON(r *nethttp.Request, dst any) error {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("request body is required")
	}
	return json.Unmarshal(body, dst)
}

func writeJSON(w nethttp.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w nethttp.ResponseWriter, err error) {
	status := nethttp.StatusBadRequest
	if strings.Contains(err.Error(), "permission") {
		status = nethttp.StatusForbidden
	}
	if strings.Contains(err.Error(), "not found") {
		status = nethttp.StatusNotFound
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func method() error {
	return fmt.Errorf("method not allowed")
}
