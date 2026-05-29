package middleware

import "net/http"

// Chain applies multiple middleware to a handler
func Chain(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	// Apply middleware in reverse order so they execute in the order specified
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}
