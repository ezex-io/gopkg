package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlog_InfoLogsToBuffer(t *testing.T) {
	var buf bytes.Buffer
	log := NewSlog(WithTextHandler(&buf, slog.LevelInfo))

	log.Info("user logged in", "user_id", "123")

	output := buf.String()
	assert.Contains(t, output, "user logged in")
	assert.Contains(t, output, "user_id=123")
}

func TestSlog_DebugIsNotLoggedAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	log := NewSlog(WithTextHandler(&buf, slog.LevelInfo))

	log.Debug("debug msg", "trace_id", "abc")

	output := buf.String()
	assert.NotContains(t, output, "debug msg")
}

func TestSlog_WithAddsFields(t *testing.T) {
	var buf bytes.Buffer
	log := NewSlog(WithTextHandler(&buf, slog.LevelInfo)).
		With("module", "auth")

	log.Info("login successful", "user_id", "456")

	output := buf.String()
	assert.Contains(t, output, "login successful")
	assert.Contains(t, output, "module=auth")
	assert.Contains(t, output, "user_id=456")
}

func TestSlog_ErrorLogsToBuffer(t *testing.T) {
	var buf bytes.Buffer
	log := NewSlog(WithTextHandler(&buf, slog.LevelError))

	log.Error("something went wrong", "err", "db_failed")

	output := buf.String()
	assert.Contains(t, output, "something went wrong")
	assert.Contains(t, output, "err=db_failed")
}

func TestSlog_DefaultFallbackHandler(t *testing.T) {
	assert.NotPanics(t, func() {
		NewSlog(nil)
	})
}
