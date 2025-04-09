package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	config := DefaultCORSConfig()
	config.AllowedOrigins = []string{"https://example.com"}
	config.AllowCredentials = true

	middleware := CORS(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://test.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close() //nolint:errcheck

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "https://example.com", res.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", res.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", res.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", res.Header.Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddlewareOptionsRequest(t *testing.T) {
	config := DefaultCORSConfig()
	middleware := CORS(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "http://test.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close() //nolint:errcheck

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.Equal(t, "*", res.Header.Get("Access-Control-Allow-Origin"))
}
