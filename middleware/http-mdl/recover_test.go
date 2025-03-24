package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverMiddleware(t *testing.T) {
	middleware := Recover()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected error")
	}))

	req := httptest.NewRequest(http.MethodGet, "http://test.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "Internal Server Error\n", w.Body.String())
}

func TestRecoverMiddleware_NoPanic(t *testing.T) {
	middleware := Recover()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("All Good"))
	}))

	req := httptest.NewRequest(http.MethodGet, "http://test.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "All Good", w.Body.String())
}
