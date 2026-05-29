package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lissy93/who-dat/internal/srcerr"
)

// bootstrapURL is the IANA RDAP bootstrap registry for DNS.
const bootstrapURL = "https://data.iana.org/rdap/dns.json"

// bootstrapTTL is how long a fetched registry is trusted before re-fetching.
const bootstrapTTL = 24 * time.Hour

// bootstrap resolves a TLD to its RDAP base URLs using the IANA registry, which it
// fetches once and refreshes lazily.
type bootstrap struct {
	http *http.Client
	url  string

	mu        sync.RWMutex
	services  map[string][]string
	fetchedAt time.Time
}

func newBootstrap(httpClient *http.Client) *bootstrap {
	return &bootstrap{http: httpClient, url: bootstrapURL}
}

// dnsRegistry mirrors the IANA bootstrap file: each service is [[tlds...], [urls...]].
type dnsRegistry struct {
	Services [][][]string `json:"services"`
}

// serversFor returns the RDAP base URLs for the registrable domain's TLD, picking the
// longest matching suffix. Returns srcerr.ErrNoSource when no TLD entry matches.
func (b *bootstrap) serversFor(ctx context.Context, asciiDomain string) ([]string, error) {
	if err := b.ensure(ctx); err != nil {
		return nil, err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	labels := strings.Split(asciiDomain, ".")
	for i := range labels {
		if urls, ok := b.services[strings.Join(labels[i:], ".")]; ok {
			return urls, nil
		}
	}
	return nil, srcerr.ErrNoSource
}

func (b *bootstrap) ensure(ctx context.Context) error {
	b.mu.RLock()
	fresh := b.services != nil && time.Since(b.fetchedAt) < bootstrapTTL
	b.mu.RUnlock()
	if fresh {
		return nil
	}
	return b.fetch(ctx)
}

func (b *bootstrap) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.url, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", srcerr.ErrUpstream, err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := b.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: bootstrap fetch: %v", srcerr.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: bootstrap status %d", srcerr.ErrUpstream, resp.StatusCode)
	}

	var reg dnsRegistry
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&reg); err != nil {
		return fmt.Errorf("%w: bootstrap decode: %v", srcerr.ErrUpstream, err)
	}

	services := make(map[string][]string)
	for _, svc := range reg.Services {
		if len(svc) < 2 {
			continue
		}
		for _, tld := range svc[0] {
			services[strings.ToLower(tld)] = svc[1]
		}
	}
	if len(services) == 0 {
		return fmt.Errorf("%w: empty bootstrap registry", srcerr.ErrUpstream)
	}

	b.mu.Lock()
	b.services = services
	b.fetchedAt = time.Now()
	b.mu.Unlock()
	return nil
}
