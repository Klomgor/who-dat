package core

import (
	"errors"
	"regexp"
	"strings"
)

// Domain validation constants
const (
	MinDomainLength = 3
	MaxDomainLength = 253
)

var (
	// Regex for validating domain names
	// Supports: example.com, sub.example.co.uk, xn--domain.com (IDN)
	domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

	ErrInvalidDomain     = errors.New("invalid domain format")
	ErrDomainTooShort    = errors.New("domain too short")
	ErrDomainTooLong     = errors.New("domain too long")
	ErrEmptyDomain       = errors.New("domain cannot be empty")
)

// ValidateDomain validates and normalizes a domain name
func ValidateDomain(domain string) (string, error) {
	// Trim whitespace
	domain = strings.TrimSpace(domain)

	// Check if empty
	if domain == "" {
		return "", ErrEmptyDomain
	}

	// Convert to lowercase for consistency
	domain = strings.ToLower(domain)

	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// Remove trailing dot if present
	domain = strings.TrimSuffix(domain, ".")

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Check length
	if len(domain) < MinDomainLength {
		return "", ErrDomainTooShort
	}
	if len(domain) > MaxDomainLength {
		return "", ErrDomainTooLong
	}

	// Validate format
	if !domainRegex.MatchString(domain) {
		return "", ErrInvalidDomain
	}

	return domain, nil
}

// ValidateDomains validates multiple domains
func ValidateDomains(domains []string) ([]string, []error) {
	validated := make([]string, 0, len(domains))
	errors := make([]error, 0)

	for _, domain := range domains {
		clean, err := ValidateDomain(domain)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		validated = append(validated, clean)
	}

	return validated, errors
}
