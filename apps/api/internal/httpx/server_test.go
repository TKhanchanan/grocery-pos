package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"grocery-pos/apps/api/internal/config"
)

func TestRoutesHandlesLoginPreflightBeforeHandler(t *testing.T) {
	server := NewServer(config.Config{
		CORSOrigins: "https://grocery-pos-front-production.up.railway.app",
		UploadDir:   t.TempDir(),
	}, nil)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://grocery-pos-front-production.up.railway.app")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://grocery-pos-front-production.up.railway.app" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}
