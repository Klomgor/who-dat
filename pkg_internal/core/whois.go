package core

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

var (
	ErrDomainNotFound = errors.New("domain not found")
	ErrTimeout        = errors.New("lookup timeout")
	ErrRateLimit      = errors.New("rate limited by WHOIS server")
)

// Service provides WHOIS lookup functionality
type Service struct {
	cache   *Cache
	enabled bool
}

// NewService creates a new WHOIS service
func NewService(cacheTTL time.Duration, enableCache bool) *Service {
	var cache *Cache
	if enableCache {
		cache = NewCache(cacheTTL)
	}

	return &Service{
		cache:   cache,
		enabled: enableCache,
	}
}

// Lookup performs a WHOIS lookup for a single domain
func (s *Service) Lookup(ctx context.Context, domain string) (*whoisparser.WhoisInfo, bool, error) {
	// Check cache first
	if s.enabled && s.cache != nil {
		if cached, found := s.cache.Get(domain); found {
			return cached, true, nil
		}
	}

	// Perform lookup with context
	info, err := s.fetchWhois(ctx, domain)
	if err != nil {
		return nil, false, err
	}

	// Cache the result
	if s.enabled && s.cache != nil {
		s.cache.Set(domain, info)
	}

	return info, false, nil
}

// LookupMulti performs WHOIS lookups for multiple domains concurrently
func (s *Service) LookupMulti(ctx context.Context, domains []string) map[string]*LookupResult {
	results := make(map[string]*LookupResult, len(domains))
	resultsCh := make(chan *domainResult, len(domains))

	// Launch goroutines for each domain
	for _, domain := range domains {
		go func(d string) {
			info, cached, err := s.Lookup(ctx, d)
			resultsCh <- &domainResult{
				domain: d,
				info:   info,
				cached: cached,
				err:    err,
			}
		}(domain)
	}

	// Collect results
	for i := 0; i < len(domains); i++ {
		select {
		case result := <-resultsCh:
			results[result.domain] = &LookupResult{
				Data:   result.info,
				Cached: result.cached,
				Error:  result.err,
			}
		case <-ctx.Done():
			// Timeout - return partial results
			return results
		}
	}

	return results
}

// fetchWhois performs the actual WHOIS lookup with retry logic
func (s *Service) fetchWhois(ctx context.Context, domain string) (*whoisparser.WhoisInfo, error) {
	var lastErr error
	maxRetries := 2

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context
		if ctx.Err() != nil {
			return nil, ErrTimeout
		}

		// Perform lookup
		raw, err := whois.Whois(domain)
		if err != nil {
			lastErr = err

			// Check for specific errors
			if isNotFoundError(err) {
				return nil, ErrDomainNotFound
			}
			if isRateLimitError(err) {
				return nil, ErrRateLimit
			}

			// Retry on other errors
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
				continue
			}
			return nil, err
		}

		// Parse the result
		info, err := whoisparser.Parse(raw)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
				continue
			}
			return nil, err
		}

		return &info, nil
	}

	return nil, lastErr
}

// CacheStats returns cache statistics
func (s *Service) CacheStats() *CacheStats {
	if s.cache != nil {
		stats := s.cache.Stats()
		return &stats
	}
	return nil
}

// Helper types

type domainResult struct {
	domain string
	info   *whoisparser.WhoisInfo
	cached bool
	err    error
}

// LookupResult represents the result of a WHOIS lookup
type LookupResult struct {
	Data   *whoisparser.WhoisInfo
	Cached bool
	Error  error
}

// Helper functions

func isNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "no match") ||
		strings.Contains(errMsg, "no data found") ||
		strings.Contains(errMsg, "no entries found")
}

func isRateLimitError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "too many requests") ||
		strings.Contains(errMsg, "quota exceeded")
}
