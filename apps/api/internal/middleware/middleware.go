package middleware

import (
	"compress/gzip"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"grocery-pos/apps/api/internal/response"
)

type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, recorder.statusCode(), time.Since(start).Round(time.Millisecond))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *statusRecorder) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func GzipJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		writer := &gzipResponseWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		_ = writer.Close()
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter  *gzip.Writer
	wroteHeader bool
	compress    bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	contentType := w.Header().Get("Content-Type")
	w.compress = status != http.StatusNoContent && status != http.StatusNotModified && strings.HasPrefix(contentType, "application/json")
	if w.compress {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.compress {
		return w.gzipWriter.Write(body)
	}
	return w.ResponseWriter.Write(body)
}

func (w *gzipResponseWriter) Close() error {
	if w.gzipWriter == nil {
		return nil
	}
	return w.gzipWriter.Close()
}

func (w *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Something went wrong.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORS(allowedOrigins string) Middleware {
	origins := splitCSV(allowedOrigins)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isAllowed(origin, origins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,X-Requested-With")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				response.NoContent(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func isAllowed(origin string, allowed []string) bool {
	for _, item := range allowed {
		if item == origin {
			return true
		}
	}
	return false
}
