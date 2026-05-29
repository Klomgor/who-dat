package types

import (
	"time"

	"github.com/likexian/whois-parser"
)

// WhoisResult wraps the WHOIS data with metadata
type WhoisResult struct {
	Domain    string                  `json:"domain"`
	Data      *whoisparser.WhoisInfo  `json:"data,omitempty"`
	Error     *ErrorResponse          `json:"error,omitempty"`
	Cached    bool                    `json:"cached"`
	Timestamp time.Time               `json:"timestamp"`
}

// MultiWhoisRequest represents a request for multiple domains
type MultiWhoisRequest struct {
	Domains []string `json:"domains"`
}

// MultiWhoisResponse represents a response for multiple domains
type MultiWhoisResponse struct {
	Results   []WhoisResult `json:"results"`
	Total     int           `json:"total"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
}

// ErrorResponse represents a standardized error
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"`
}

// Error codes
const (
	ErrCodeInvalidDomain   = "INVALID_DOMAIN"
	ErrCodeDomainNotFound  = "DOMAIN_NOT_FOUND"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeRateLimit       = "RATE_LIMIT"
	ErrCodeTooManyDomains  = "TOO_MANY_DOMAINS"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeInternal        = "INTERNAL_ERROR"
)

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}
