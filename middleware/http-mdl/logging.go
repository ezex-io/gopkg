package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging logs incoming HTTP requests and their duration
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)

			log.Printf("[%s] %s %s %dms",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				duration.Milliseconds(),
			)
		})
	}
}
