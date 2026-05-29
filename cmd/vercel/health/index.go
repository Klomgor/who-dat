package handler

import (
	"net/http"

	"github.com/lissy93/who-dat/pkg_internal/handler"
	"github.com/lissy93/who-dat/pkg_internal/middleware"
)

const version = "2.0.0"

// Handler is the Vercel serverless function entry point for health checks
func Handler(w http.ResponseWriter, r *http.Request) {
	// Health check doesn't need auth
	h := middleware.Chain(
		handler.NewHealthHandler(version),
		middleware.CORS(),
	)

	h.ServeHTTP(w, r)
}
