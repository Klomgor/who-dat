package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lissy93/who-dat/pkg_internal/config"
	"github.com/lissy93/who-dat/pkg_internal/core"
	"github.com/lissy93/who-dat/pkg_internal/response"
	"github.com/lissy93/who-dat/pkg/types"
)

// SingleHandler handles single domain WHOIS lookups
type SingleHandler struct {
	service *core.Service
	timeout time.Duration
}

// NewSingleHandler creates a new single domain handler
func NewSingleHandler(cfg *config.Config) *SingleHandler {
	return &SingleHandler{
		service: core.NewService(cfg.CacheTTL, cfg.EnableCache),
		timeout: cfg.RequestTimeout,
	}
}

func (h *SingleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract domain from URL path
	domain := strings.TrimPrefix(r.URL.Path, "/")
	domain = strings.TrimSpace(domain)

	// If empty, show help message
	if domain == "" || domain == "/" {
		h.showHelp(w)
		return
	}

	// Validate domain
	cleanDomain, err := core.ValidateDomain(domain)
	if err != nil {
		response.BadRequest(w, "Invalid domain: "+err.Error())
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Perform lookup
	info, cached, err := h.service.Lookup(ctx, cleanDomain)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Build response
	result := types.WhoisResult{
		Domain:    cleanDomain,
		Data:      info,
		Cached:    cached,
		Timestamp: time.Now(),
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *SingleHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrDomainNotFound):
		response.NotFound(w, "Domain not found or no WHOIS data available")
	case errors.Is(err, core.ErrTimeout):
		response.GatewayTimeout(w, "WHOIS lookup timed out")
	case errors.Is(err, core.ErrRateLimit):
		response.TooManyRequests(w, "Rate limited by WHOIS server")
	default:
		response.InternalError(w, "Failed to perform WHOIS lookup")
	}
}

func (h *SingleHandler) showHelp(w http.ResponseWriter) {
	help := map[string]interface{}{
		"service": "Who-Dat WHOIS Lookup API",
		"usage": map[string]string{
			"single": "GET /{domain}",
			"multi":  "GET /multi?domains=domain1.com,domain2.com",
			"health": "GET /ping",
		},
		"example": "GET /google.com",
		"docs":    "https://github.com/lissy93/who-dat",
	}
	response.JSON(w, http.StatusOK, help)
}
