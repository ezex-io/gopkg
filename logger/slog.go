package logger

import (
	"io"
	"log/slog"
	"os"
)

type Slog struct {
	log *slog.Logger
}

type SlogHandler func() *slog.Logger

var DefaultSlog = NewSlog(nil)

// NewSlog creates a new Slog logger using functional options.
func NewSlog(handler SlogHandler) *Slog {
	if handler == nil {
		handler = WithTextHandler(os.Stdout, slog.LevelInfo)
	}

	return &Slog{
		log: handler(),
	}
}

// WithJSONHandler returns a logger with JSON formatting and custom level.
func WithJSONHandler(w io.Writer, level slog.Level) SlogHandler {
	return func() *slog.Logger {
		if w == nil {
			w = os.Stdout
		}
		handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		})

		return slog.New(handler)
	}
}

// WithTextHandler returns a logger with text formatting and custom level.
func WithTextHandler(w io.Writer, level slog.Level) SlogHandler {
	return func() *slog.Logger {
		if w == nil {
			w = os.Stdout
		}
		handler := slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: level,
		})

		return slog.New(handler)
	}
}

func (s *Slog) Debug(msg string, args ...any) {
	s.log.Debug(msg, args...)
}

func (s *Slog) Info(msg string, args ...any) {
	s.log.Info(msg, args...)
}

func (s *Slog) Warn(msg string, args ...any) {
	s.log.Warn(msg, args...)
}

func (s *Slog) Error(msg string, args ...any) {
	s.log.Error(msg, args...)
}

func (s *Slog) Fatal(msg string, args ...any) {
	s.log.Error(msg, args...)
	//nolint:revive // exit on fatal log
	os.Exit(1)
}

func (s *Slog) With(args ...any) *Slog {
	return &Slog{
		log: s.log.With(args...),
	}
}
