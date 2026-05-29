package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/lissy93/who-dat/pkg_internal/config"
	"github.com/lissy93/who-dat/pkg_internal/handler"
	"github.com/lissy93/who-dat/pkg_internal/middleware"
)

//go:embed dist
var distFS embed.FS

const version = "2.0.0"

var startTime = time.Now()

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Who-Dat server v%s on port %s", version, cfg.Port)
	log.Printf("Cache enabled: %v (TTL: %v)", cfg.EnableCache, cfg.CacheTTL)
	log.Printf("Auth enabled: %v", cfg.IsAuthEnabled())

	// Create handlers (reuse for caching)
	healthHandler := handler.NewHealthHandler(version)
	multiHandler := handler.NewMultiHandler(cfg)
	singleHandler := handler.NewSingleHandler(cfg)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.Handle("/ping", middleware.Chain(
		healthHandler,
		middleware.CORS(),
		middleware.Logger(),
	))

	// Multi-domain endpoint (with auth)
	mux.Handle("/multi", middleware.Chain(
		multiHandler,
		middleware.CORS(),
		middleware.Logger(),
		middleware.Auth(cfg),
	))

	// Static files (frontend)
	distSubFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Printf("Warning: Could not load frontend assets: %v", err)
	} else {
		fileServer := http.FileServer(http.FS(distSubFS))
		mux.Handle("/assets/", fileServer)
		mux.Handle("/favicon.ico", fileServer)

		// Create middleware chain for single domain handler
		singleHandlerWithMiddleware := middleware.Chain(
			singleHandler,
			middleware.CORS(),
			middleware.Logger(),
			middleware.Auth(cfg),
		)

		// Serve index.html for root
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// If it's the root path, serve index.html
			if r.URL.Path == "/" {
				indexHTML, err := distSubFS.Open("index.html")
				if err == nil {
					defer indexHTML.Close()
					content, _ := io.ReadAll(indexHTML)
					w.Header().Set("Content-Type", "text/html")
					w.Write(content)
					return
				}
			}

			// Otherwise, it's a domain lookup (with auth)
			singleHandlerWithMiddleware.ServeHTTP(w, r)
		})
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
