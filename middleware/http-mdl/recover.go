package middleware

import (
	"log/slog"
	"net/http"
	"runtime"
	"strings"
)

// Recover middleware with structured stack trace logging
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := captureStackTrace(3) // Skip 3 frames to start at panic origin
					slog.Error("panic recovered",
						"error", err,
						"stack", stack,
					)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// captureStackTrace formats the stack trace in a structured and readable way
func captureStackTrace(skip int) []map[string]any {
	var pcs [32]uintptr
	n := runtime.Callers(skip, pcs[:])

	var stackTrace []map[string]any
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		// Skip runtime internal frames
		if !strings.Contains(frame.File, "runtime/") {
			stackTrace = append(stackTrace, map[string]any{
				"function": frame.Function,
				"file":     frame.File,
				"line":     frame.Line,
			})
		}
		if !more {
			break
		}
	}

	return stackTrace
}
