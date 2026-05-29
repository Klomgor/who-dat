package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lissy93/who-dat/pkg_internal/response"
)

// Rate limit configuration
const (
	MaxRequestsPerMinute = 10
	MaxRequestsPerHour   = 100
	MaxRequestsPerDay    = 1000
)

// requestRecord tracks requests for an IP address
type requestRecord struct {
	minuteCount   int
	hourCount     int
	dayCount      int
	minuteReset   time.Time
	hourReset     time.Time
	dayReset      time.Time
	lastCleanup   time.Time
}

// rateLimiter manages rate limiting state
type rateLimiter struct {
	mu      sync.RWMutex
	records map[string]*requestRecord
}

var limiter = &rateLimiter{
	records: make(map[string]*requestRecord),
}

// RateLimit creates a rate limiting middleware
func RateLimit() func(http.Handler) http.Handler {
	// Start cleanup goroutine
	go limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			// Check if rate limit exceeded
			if !limiter.allow(ip) {
				// Get remaining limits for response
				remaining := limiter.getRemaining(ip)

				w.Header().Set("X-RateLimit-Limit-Minute", fmt.Sprintf("%d", MaxRequestsPerMinute))
				w.Header().Set("X-RateLimit-Limit-Hour", fmt.Sprintf("%d", MaxRequestsPerHour))
				w.Header().Set("X-RateLimit-Limit-Day", fmt.Sprintf("%d", MaxRequestsPerDay))
				w.Header().Set("X-RateLimit-Remaining-Minute", fmt.Sprintf("%d", remaining.minute))
				w.Header().Set("X-RateLimit-Remaining-Hour", fmt.Sprintf("%d", remaining.hour))
				w.Header().Set("X-RateLimit-Remaining-Day", fmt.Sprintf("%d", remaining.day))
				w.Header().Set("Retry-After", "60")

				response.TooManyRequests(w, "Rate limit exceeded. Limits: 10/min, 100/hour, 1000/day")
				return
			}

			// Add rate limit headers to successful responses
			remaining := limiter.getRemaining(ip)
			w.Header().Set("X-RateLimit-Limit-Minute", fmt.Sprintf("%d", MaxRequestsPerMinute))
			w.Header().Set("X-RateLimit-Limit-Hour", fmt.Sprintf("%d", MaxRequestsPerHour))
			w.Header().Set("X-RateLimit-Limit-Day", fmt.Sprintf("%d", MaxRequestsPerDay))
			w.Header().Set("X-RateLimit-Remaining-Minute", fmt.Sprintf("%d", remaining.minute))
			w.Header().Set("X-RateLimit-Remaining-Hour", fmt.Sprintf("%d", remaining.hour))
			w.Header().Set("X-RateLimit-Remaining-Day", fmt.Sprintf("%d", remaining.day))

			next.ServeHTTP(w, r)
		})
	}
}

// allow checks if a request from the given IP is allowed
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	record := rl.getOrCreateRecord(ip, now)

	// Reset counters if windows have expired
	if now.After(record.minuteReset) {
		record.minuteCount = 0
		record.minuteReset = now.Add(time.Minute)
	}
	if now.After(record.hourReset) {
		record.hourCount = 0
		record.hourReset = now.Add(time.Hour)
	}
	if now.After(record.dayReset) {
		record.dayCount = 0
		record.dayReset = now.Add(24 * time.Hour)
	}

	// Check limits
	if record.minuteCount >= MaxRequestsPerMinute {
		return false
	}
	if record.hourCount >= MaxRequestsPerHour {
		return false
	}
	if record.dayCount >= MaxRequestsPerDay {
		return false
	}

	// Increment counters
	record.minuteCount++
	record.hourCount++
	record.dayCount++

	return true
}

// getRemaining returns remaining requests for each time window
func (rl *rateLimiter) getRemaining(ip string) struct{ minute, hour, day int } {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	record := rl.getOrCreateRecordUnsafe(ip, now)

	var remaining struct{ minute, hour, day int }

	// Calculate remaining based on current counts and resets
	if now.After(record.minuteReset) {
		remaining.minute = MaxRequestsPerMinute
	} else {
		remaining.minute = MaxRequestsPerMinute - record.minuteCount
	}

	if now.After(record.hourReset) {
		remaining.hour = MaxRequestsPerHour
	} else {
		remaining.hour = MaxRequestsPerHour - record.hourCount
	}

	if now.After(record.dayReset) {
		remaining.day = MaxRequestsPerDay
	} else {
		remaining.day = MaxRequestsPerDay - record.dayCount
	}

	if remaining.minute < 0 {
		remaining.minute = 0
	}
	if remaining.hour < 0 {
		remaining.hour = 0
	}
	if remaining.day < 0 {
		remaining.day = 0
	}

	return remaining
}

// getOrCreateRecord gets or creates a record for an IP (requires lock)
func (rl *rateLimiter) getOrCreateRecord(ip string, now time.Time) *requestRecord {
	record, exists := rl.records[ip]
	if !exists {
		record = &requestRecord{
			minuteReset: now.Add(time.Minute),
			hourReset:   now.Add(time.Hour),
			dayReset:    now.Add(24 * time.Hour),
			lastCleanup: now,
		}
		rl.records[ip] = record
	}
	return record
}

// getOrCreateRecordUnsafe gets or creates a record without locking (for read-only access)
func (rl *rateLimiter) getOrCreateRecordUnsafe(ip string, now time.Time) *requestRecord {
	record, exists := rl.records[ip]
	if !exists {
		return &requestRecord{
			minuteReset: now.Add(time.Minute),
			hourReset:   now.Add(time.Hour),
			dayReset:    now.Add(24 * time.Hour),
		}
	}
	return record
}

// cleanup removes expired records periodically
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for ip, record := range rl.records {
			// Remove records that haven't been used in 25 hours (1 day + buffer)
			if now.Sub(record.lastCleanup) > 25*time.Hour {
				delete(rl.records, ip)
			}
		}

		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (Vercel sets this)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, c := range xff {
					if c == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return xff[:commaIdx]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
