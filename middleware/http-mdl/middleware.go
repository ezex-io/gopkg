package middleware

import "net/http"

// Middleware defines the middleware function signature
type Middleware func(http.Handler) http.Handler

// Chain allows chaining multiple middleware functions
func Chain(mw ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			final = mw[i](final)
		}
		return final
	}
}
