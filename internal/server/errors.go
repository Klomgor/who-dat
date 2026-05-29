package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/srcerr"
)

// Error codes returned in the error envelope.
const (
	codeInvalidDomain  = "INVALID_DOMAIN"
	codeUnsupportedTLD = "UNSUPPORTED_TLD"
	codeRateLimited    = "RATE_LIMITED"
	codeUpstreamError  = "UPSTREAM_ERROR"
	codeUpstreamTimout = "UPSTREAM_TIMEOUT"
	codeInternal       = "INTERNAL_ERROR"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Query   string `json:"query"`
}

// writeError sends the standard error envelope with the given status.
func writeError(w http.ResponseWriter, status int, code, message, query string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message, Query: query}})
}

// writeDomainError maps a domain.Parse error to a 400 envelope.
func writeDomainError(w http.ResponseWriter, err error, query string) {
	switch {
	case errors.Is(err, domain.ErrNoPublicSuffix):
		writeError(w, http.StatusBadRequest, codeInvalidDomain, "no valid public suffix", query)
	default:
		writeError(w, http.StatusBadRequest, codeInvalidDomain, "could not parse a registrable domain", query)
	}
}

// writeLookupError maps an orchestrator/source error to the right status and envelope.
func writeLookupError(w http.ResponseWriter, err error, query string) {
	var rl *srcerr.RateLimited
	switch {
	case errors.As(err, &rl):
		if rl.RetryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.RetryAfter.Seconds())))
		}
		writeError(w, http.StatusTooManyRequests, codeRateLimited, "rate limited; retry later", query)
	case errors.Is(err, srcerr.ErrTimeout), errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, codeUpstreamTimout, "registry timed out", query)
	case errors.Is(err, srcerr.ErrNoSource):
		writeError(w, http.StatusNotImplemented, codeUnsupportedTLD, "no rdap or whois source for this tld", query)
	case errors.Is(err, srcerr.ErrUpstream):
		writeError(w, http.StatusBadGateway, codeUpstreamError, "registry unreachable or returned an invalid response", query)
	default:
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error", query)
	}
}
