package handler

import (
	"net/http"

	"github.com/lissy93/who-dat/pkg_internal/config"
	"github.com/lissy93/who-dat/pkg_internal/handler"
	"github.com/lissy93/who-dat/pkg_internal/middleware"
)

// Handler is the Vercel serverless function entry point for multi-domain lookups
func Handler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Load()

	// Create handler with middleware chain (rate limit first)
	h := middleware.Chain(
		handler.NewMultiHandler(cfg),
		middleware.CORS(),
		middleware.Logger(),
		middleware.Auth(cfg),
		middleware.RateLimit(),
	)

	h.ServeHTTP(w, r)
}
