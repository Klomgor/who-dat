package middleware

import (
	"net/http"
	"strings"

	"github.com/lissy93/who-dat/pkg_internal/config"
	"github.com/lissy93/who-dat/pkg_internal/response"
)

// Auth creates an authentication middleware
func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if not configured
			if !cfg.IsAuthEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "API key required. See: github.com/lissy93/who-dat")
				return
			}

			// Remove "Bearer " prefix if present
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			if token != cfg.AuthKey {
				response.Unauthorized(w, "Invalid API key. See: github.com/lissy93/who-dat")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
