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
	defer log.SetOutput(nil)

	middleware := Logging()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://test.com/foo", nil)
	req.RemoteAddr = "127.0.0.1"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	logged := logBuffer.String()
	assert.Contains(t, logged, "[GET] /foo 127.0.0.1")
	assert.True(t, strings.Contains(logged, "ms"))
}
