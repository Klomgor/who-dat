package unit

import (
	"testing"

	"github.com/lissy93/who-dat/pkg_internal/core"
)

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{"valid domain", "example.com", "example.com", false},
		{"valid subdomain", "api.example.com", "api.example.com", false},
		{"uppercase", "EXAMPLE.COM", "example.com", false},
		{"with whitespace", "  example.com  ", "example.com", false},
		{"with http", "http://example.com", "example.com", false},
		{"with https", "https://example.com", "example.com", false},
		{"with www", "www.example.com", "example.com", false},
		{"with path", "example.com/path", "example.com", false},
		{"with trailing dot", "example.com.", "example.com", false},
		{"ccTLD", "example.co.uk", "example.co.uk", false},
		{"new gTLD", "example.dev", "example.dev", false},
		{"empty string", "", "", true},
		{"only whitespace", "   ", "", true},
		{"too short", "a.b", "", true},
		{"no TLD", "example", "", true},
		{"invalid chars", "exam ple.com", "", true},
		{"starts with dash", "-example.com", "", true},
		{"ends with dash", "example-.com", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := core.ValidateDomain(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %q, got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestValidateDomains(t *testing.T) {
	tests := []struct {
		name           string
		input          []string
		expectedValid  int
		expectedErrors int
	}{
		{
			"all valid",
			[]string{"example.com", "google.com", "github.com"},
			3,
			0,
		},
		{
			"mixed valid and invalid",
			[]string{"example.com", "invalid", "google.com"},
			2,
			1,
		},
		{
			"all invalid",
			[]string{"invalid", "no-tld", ""},
			0,
			3,
		},
		{
			"empty list",
			[]string{},
			0,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errors := core.ValidateDomains(tt.input)

			if len(valid) != tt.expectedValid {
				t.Errorf("Expected %d valid domains, got %d", tt.expectedValid, len(valid))
			}
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(errors))
			}
		})
	}
}
