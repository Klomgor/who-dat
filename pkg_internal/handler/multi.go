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

// MultiHandler handles multi-domain WHOIS lookups
type MultiHandler struct {
	service    *core.Service
	timeout    time.Duration
	maxDomains int
}

// NewMultiHandler creates a new multi-domain handler
func NewMultiHandler(cfg *config.Config) *MultiHandler {
	return &MultiHandler{
		service:    core.NewService(cfg.CacheTTL, cfg.EnableCache),
		timeout:    cfg.RequestTimeout,
		maxDomains: cfg.MaxDomainsPerRequest,
	}
}

func (h *MultiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get domains from query parameter
	domainsParam := r.URL.Query().Get("domains")
	if domainsParam == "" {
		response.BadRequest(w, "Missing 'domains' query parameter")
		return
	}

	// Split domains by comma
	domainList := strings.Split(domainsParam, ",")
	if len(domainList) == 0 {
		response.BadRequest(w, "No domains provided")
		return
	}

	// Check max domains limit
	if len(domainList) > h.maxDomains {
		response.BadRequest(w, "Too many domains. Maximum allowed: "+string(rune(h.maxDomains)))
		return
	}

	// Validate all domains
	validDomains, validationErrors := core.ValidateDomains(domainList)
	if len(validDomains) == 0 {
		response.BadRequest(w, "No valid domains provided")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Perform lookups
	lookupResults := h.service.LookupMulti(ctx, validDomains)

	// Build response
	results := make([]types.WhoisResult, 0, len(validDomains))
	succeeded := 0
	failed := 0

	for domain, result := range lookupResults {
		whoisResult := types.WhoisResult{
			Domain:    domain,
			Cached:    result.Cached,
			Timestamp: time.Now(),
		}

		if result.Error != nil {
			whoisResult.Error = h.errorToResponse(result.Error)
			failed++
		} else {
			whoisResult.Data = result.Data
			succeeded++
		}

		results = append(results, whoisResult)
	}

	// Add validation errors as failed results
	for _, err := range validationErrors {
		failed++
		results = append(results, types.WhoisResult{
			Error:     types.NewErrorResponse(types.ErrCodeInvalidDomain, err.Error()),
			Timestamp: time.Now(),
		})
	}

	multiResponse := types.MultiWhoisResponse{
		Results:   results,
		Total:     len(results),
		Succeeded: succeeded,
		Failed:    failed,
	}

	response.JSON(w, http.StatusOK, multiResponse)
}

func (h *MultiHandler) errorToResponse(err error) *types.ErrorResponse {
	switch {
	case errors.Is(err, core.ErrDomainNotFound):
		return types.NewErrorResponse(types.ErrCodeDomainNotFound, "Domain not found")
	case errors.Is(err, core.ErrTimeout):
		return types.NewErrorResponse(types.ErrCodeTimeout, "Lookup timed out")
	case errors.Is(err, core.ErrRateLimit):
		return types.NewErrorResponse(types.ErrCodeRateLimit, "Rate limited")
	default:
		return types.NewErrorResponse(types.ErrCodeInternal, "Lookup failed")
	}
}
