package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger creates a logging middleware (minimal - only logs errors)
func Logger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Only log errors (4xx, 5xx)
			if wrapped.statusCode >= 400 {
				duration := time.Since(start)
				log.Printf("ERROR: %s %s - Status: %d - Duration: %v",
					r.Method, r.URL.Path, wrapped.statusCode, duration)
			}
		})
	}
}
