package middleware

import (
	"net/http"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// DefaultCORSConfig returns a default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}
}

// CORS creates middleware to handle CORS requests.
func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		originHeader := strings.Join(config.AllowedOrigins, ", ")
		methodsHeader := strings.Join(config.AllowedMethods, ", ")
		headersHeader := strings.Join(config.AllowedHeaders, ", ")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", originHeader)
			w.Header().Set("Access-Control-Allow-Methods", methodsHeader)
			w.Header().Set("Access-Control-Allow-Headers", headersHeader)

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
