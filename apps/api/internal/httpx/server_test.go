package httpx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestRoutesServesUploadsAndMissingFilesReturnNotFound(t *testing.T) {
	uploadDir := t.TempDir()
	productDir := filepath.Join(uploadDir, "products")
	if err := os.MkdirAll(productDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(productDir, "sample.png"), []byte("image-bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	server := NewServer(config.Config{UploadDir: uploadDir}, nil)

	t.Run("serves persisted upload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/uploads/products/sample.png", nil)
		rec := httptest.NewRecorder()

		server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Body.String(); got != "image-bytes" {
			t.Fatalf("body = %q, want image bytes", got)
		}
	})

	t.Run("missing upload does not crash", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/uploads/products/missing.png", nil)
		rec := httptest.NewRecorder()

		server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}
