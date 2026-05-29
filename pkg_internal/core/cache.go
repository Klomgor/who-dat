package core

import (
	"sync"
	"time"

	"github.com/likexian/whois-parser"
)

// cacheEntry represents a cached WHOIS result
type cacheEntry struct {
	data      *whoisparser.WhoisInfo
	expiresAt time.Time
}

// Cache provides thread-safe caching with TTL
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
	stats   CacheStats
}

// CacheStats tracks cache performance
type CacheStats struct {
	Hits   int64
	Misses int64
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (*whoisparser.WhoisInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return entry.data, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value *whoisparser.WhoisInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		data:      value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// cleanupLoop periodically removes expired entries
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}
