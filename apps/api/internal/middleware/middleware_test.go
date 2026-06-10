package middleware

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSPreflightForAllowedOrigin(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	handler := CORS(" http://localhost:5173, https://grocery-pos-front-production.up.railway.app ")(next)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://grocery-pos-front-production.up.railway.app")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("preflight reached the next handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	headers := map[string]string{
		"Access-Control-Allow-Origin":      "https://grocery-pos-front-production.up.railway.app",
		"Access-Control-Allow-Methods":     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		"Access-Control-Allow-Headers":     "Authorization,Content-Type,Accept,Origin,X-Requested-With",
		"Access-Control-Allow-Credentials": "true",
		"Vary":                             "Origin",
	}
	for name, want := range headers {
		if got := rec.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestCORSDoesNotAllowUnmatchedOrWildcardOrigin(t *testing.T) {
	for _, allowedOrigins := range []string{
		"https://grocery-pos-front-production.up.railway.app",
		"*",
	} {
		t.Run(allowedOrigins, func(t *testing.T) {
			handler := CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
			req.Header.Set("Origin", "https://untrusted.example")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
			}
			if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
				t.Fatalf("Access-Control-Allow-Credentials = %q, want empty", got)
			}
		})
	}
}

func TestGzipJSONCompressesJSONResponses(t *testing.T) {
	handler := GzipJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]bool
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload["success"] {
		t.Fatalf("payload = %s", body)
	}
}

func TestStatusRecorderCapturesImplicitStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	writer := &statusRecorder{ResponseWriter: rec}

	_, _ = writer.Write([]byte("ok"))

	if writer.statusCode() != http.StatusOK {
		t.Fatalf("status = %d, want %d", writer.statusCode(), http.StatusOK)
	}
}
