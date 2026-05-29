package unit

import (
	"testing"
	"time"

	"github.com/likexian/whois-parser"
	"github.com/lissy93/who-dat/pkg_internal/core"
)

func TestCacheSetAndGet(t *testing.T) {
	cache := core.NewCache(1 * time.Hour)

	// Create test data
	testData := &whoisparser.WhoisInfo{
		Domain: &whoisparser.Domain{
			Domain: "example.com",
		},
	}

	// Test Set and Get
	cache.Set("example.com", testData)
	result, found := cache.Get("example.com")

	if !found {
		t.Error("Expected to find cached value")
	}

	if result.Domain.Domain != "example.com" {
		t.Errorf("Expected domain example.com, got %s", result.Domain.Domain)
	}
}

func TestCacheMiss(t *testing.T) {
	cache := core.NewCache(1 * time.Hour)

	result, found := cache.Get("nonexistent.com")

	if found {
		t.Error("Expected cache miss")
	}

	if result != nil {
		t.Error("Expected nil result on cache miss")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := core.NewCache(100 * time.Millisecond)

	testData := &whoisparser.WhoisInfo{
		Domain: &whoisparser.Domain{
			Domain: "example.com",
		},
	}

	// Set value
	cache.Set("example.com", testData)

	// Should be found immediately
	_, found := cache.Get("example.com")
	if !found {
		t.Error("Expected to find cached value immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("example.com")
	if found {
		t.Error("Expected cache entry to be expired")
	}
}

func TestCacheStats(t *testing.T) {
	cache := core.NewCache(1 * time.Hour)

	testData := &whoisparser.WhoisInfo{
		Domain: &whoisparser.Domain{
			Domain: "example.com",
		},
	}

	cache.Set("example.com", testData)

	// Generate hits and misses
	cache.Get("example.com")  // hit
	cache.Get("example.com")  // hit
	cache.Get("nonexistent.com")  // miss

	stats := cache.Stats()

	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestCacheClear(t *testing.T) {
	cache := core.NewCache(1 * time.Hour)

	testData := &whoisparser.WhoisInfo{
		Domain: &whoisparser.Domain{
			Domain: "example.com",
		},
	}

	cache.Set("example.com", testData)
	cache.Set("google.com", testData)

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func TestCacheDelete(t *testing.T) {
	cache := core.NewCache(1 * time.Hour)

	testData := &whoisparser.WhoisInfo{
		Domain: &whoisparser.Domain{
			Domain: "example.com",
		},
	}

	cache.Set("example.com", testData)
	cache.Delete("example.com")

	_, found := cache.Get("example.com")
	if found {
		t.Error("Expected entry to be deleted")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := core.NewCache(1 * time.Hour)

	testData := &whoisparser.WhoisInfo{
		Domain: &whoisparser.Domain{
			Domain: "example.com",
		},
	}

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cache.Set("example.com", testData)
			}
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cache.Get("example.com")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic or deadlock
	t.Log("Concurrency test passed")
}
