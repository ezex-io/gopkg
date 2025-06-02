package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	middleware := Logging()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://test.com/foo", http.NoBody)
	req.RemoteAddr = "127.0.0.1"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	res := w.Result()
	defer func() { _ = res.Body.Close() }()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	logged := logBuffer.String()
	assert.Contains(t, logged, "[GET] /foo 127.0.0.1")
	assert.True(t, strings.Contains(logged, "ms"))
}
