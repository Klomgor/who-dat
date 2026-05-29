package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lissy93/who-dat/pkg_internal/config"
	"github.com/lissy93/who-dat/pkg_internal/handler"
)

func TestHealthHandler(t *testing.T) {
	handler := handler.NewHealthHandler("2.0.0-test")

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "ok") {
		t.Errorf("Expected response to contain 'ok', got %s", body)
	}

	if !strings.Contains(body, "2.0.0-test") {
		t.Errorf("Expected response to contain version, got %s", body)
	}
}

func TestSingleHandlerInvalidDomain(t *testing.T) {
	cfg := &config.Config{
		CacheTTL:             1 * time.Hour,
		RequestTimeout:       5 * time.Second,
		MaxDomainsPerRequest: 10,
		EnableCache:          true,
	}

	handler := handler.NewSingleHandler(cfg)

	// Test invalid domain
	req := httptest.NewRequest("GET", "/invalid-domain", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Invalid domain") {
		t.Errorf("Expected error message about invalid domain, got %s", body)
	}
}

func TestSingleHandlerHelp(t *testing.T) {
	cfg := &config.Config{
		CacheTTL:             1 * time.Hour,
		RequestTimeout:       5 * time.Second,
		MaxDomainsPerRequest: 10,
		EnableCache:          true,
	}

	handler := handler.NewSingleHandler(cfg)

	// Test root path
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "usage") {
		t.Errorf("Expected help message with usage, got %s", body)
	}
}

func TestMultiHandlerMissingDomains(t *testing.T) {
	cfg := &config.Config{
		CacheTTL:             1 * time.Hour,
		RequestTimeout:       5 * time.Second,
		MaxDomainsPerRequest: 10,
		EnableCache:          true,
	}

	handler := handler.NewMultiHandler(cfg)

	// Test without domains parameter
	req := httptest.NewRequest("GET", "/multi", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "domains") {
		t.Errorf("Expected error message about missing domains, got %s", body)
	}
}

func TestMultiHandlerInvalidDomains(t *testing.T) {
	cfg := &config.Config{
		CacheTTL:             1 * time.Hour,
		RequestTimeout:       5 * time.Second,
		MaxDomainsPerRequest: 10,
		EnableCache:          true,
	}

	handler := handler.NewMultiHandler(cfg)

	// Test with invalid domains
	req := httptest.NewRequest("GET", "/multi?domains=invalid,no-tld", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "valid") {
		t.Errorf("Expected error message about valid domains, got %s", body)
	}
}
