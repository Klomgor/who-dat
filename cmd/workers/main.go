// +build js,wasm

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"syscall/js"
	"time"

	"github.com/lissy93/who-dat/pkg_internal/config"
	"github.com/lissy93/who-dat/pkg_internal/core"
	"github.com/lissy93/who-dat/pkg/types"
)

var whoisService *core.Service

func main() {
	// Initialize service
	cfg := config.Load()
	whoisService = core.NewService(cfg.CacheTTL, cfg.EnableCache)

	// Register JavaScript functions
	js.Global().Set("handleWhoisLookup", js.FuncOf(handleWhoisLookup))
	js.Global().Set("handleMultiWhoisLookup", js.FuncOf(handleMultiWhoisLookup))

	// Keep the program running
	<-make(chan struct{})
}

// handleWhoisLookup handles single domain lookups from JavaScript
func handleWhoisLookup(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "domain parameter required",
		})
	}

	domain := args[0].String()

	// Validate domain
	cleanDomain, err := core.ValidateDomain(domain)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Perform lookup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, cached, err := whoisService.Lookup(ctx, cleanDomain)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Build response
	result := types.WhoisResult{
		Domain:    cleanDomain,
		Data:      info,
		Cached:    cached,
		Timestamp: time.Now(),
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(result)
	return js.ValueOf(string(jsonData))
}

// handleMultiWhoisLookup handles multi-domain lookups from JavaScript
func handleMultiWhoisLookup(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "domains parameter required",
		})
	}

	// Parse domains array from JavaScript
	domainsJS := args[0]
	domainsLen := domainsJS.Length()
	domains := make([]string, domainsLen)
	for i := 0; i < domainsLen; i++ {
		domains[i] = domainsJS.Index(i).String()
	}

	// Validate domains
	validDomains, _ := core.ValidateDomains(domains)
	if len(validDomains) == 0 {
		return js.ValueOf(map[string]interface{}{
			"error": "no valid domains provided",
		})
	}

	// Perform lookups
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lookupResults := whoisService.LookupMulti(ctx, validDomains)

	// Build response
	results := make([]types.WhoisResult, 0, len(validDomains))
	for domain, result := range lookupResults {
		whoisResult := types.WhoisResult{
			Domain:    domain,
			Cached:    result.Cached,
			Timestamp: time.Now(),
		}

		if result.Error != nil {
			whoisResult.Error = types.NewErrorResponse(types.ErrCodeInternal, result.Error.Error())
		} else {
			whoisResult.Data = result.Data
		}

		results = append(results, whoisResult)
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(results)
	return js.ValueOf(string(jsonData))
}
