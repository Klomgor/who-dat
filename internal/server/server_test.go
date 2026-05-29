package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lissy93/who-dat/internal/config"
	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/lookup"
	"github.com/lissy93/who-dat/internal/model"
	"github.com/lissy93/who-dat/internal/srcerr"
)

type fakeSource struct {
	result *model.Result
	err    error
}

func (f *fakeSource) Lookup(context.Context, domain.Name) (*model.Result, error) {
	return f.result, f.err
}

// newTestServer builds a handler whose RDAP source returns the given result/err. The WHOIS
// source always reports no source, so it never masks the RDAP outcome under test.
func newTestServer(result *model.Result, err error) http.Handler {
	cfg := &config.Config{LookupTimeout: 2_000_000_000, MaxDomains: 5} // 2s, no rate limit/auth
	svc := lookup.NewService(&fakeSource{result: result, err: err}, &fakeSource{err: srcerr.ErrNoSource}, nil)
	return New(cfg, svc)
}

func registered() *model.Result {
	r := model.New("example.com", "example.com", "com")
	r.IsRegistered = true
	r.Meta = model.Meta{Source: model.SourceRDAP}
	r.Raw = []byte("RAW-RDAP-BODY")
	r.RawContentType = "application/rdap+json"
	return r
}

func TestLookupStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		result     *model.Result
		err        error
		wantStatus int
		wantCode   string
	}{
		{"ok", "/v1/whois/example.com", registered(), nil, http.StatusOK, ""},
		{"invalid domain", "/v1/whois/bad_label.com", nil, nil, http.StatusBadRequest, codeInvalidDomain},
		{"upstream", "/v1/whois/example.com", nil, srcerr.ErrUpstream, http.StatusBadGateway, codeUpstreamError},
		{"timeout", "/v1/whois/example.com", nil, srcerr.ErrTimeout, http.StatusGatewayTimeout, codeUpstreamTimout},
		{"unsupported", "/v1/whois/example.com", nil, srcerr.ErrNoSource, http.StatusNotImplemented, codeUnsupportedTLD},
		{"rate limited", "/v1/whois/example.com", nil, &srcerr.RateLimited{}, http.StatusTooManyRequests, codeRateLimited},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newTestServer(tt.result, tt.err).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if tt.wantCode != "" {
				var body errorBody
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				if body.Error.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", body.Error.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestLookupSuccessShape(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(registered(), nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/whois/example.com", nil))

	var res model.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Query != "example.com" || !res.IsRegistered || res.Meta.Source != model.SourceRDAP {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestRawPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(registered(), nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/whois/example.com?raw=true", nil))

	if got := rec.Header().Get("Content-Type"); got != "application/rdap+json" {
		t.Errorf("content type = %q", got)
	}
	if rec.Body.String() != "RAW-RDAP-BODY" {
		t.Errorf("raw body = %q", rec.Body.String())
	}
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
