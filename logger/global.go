package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	globLogger Logger
	once       sync.Once
)

// InitGlobalLogger init global logger.
func InitGlobalLogger() {
	once.Do(func() {
		globLogger = NewSlog(WithTextHandler(os.Stdout, slog.LevelDebug))
	})
}

func Debug(msg string, args ...any) {
	log(msg, slog.LevelDebug, args...)
}

func Info(msg string, args ...any) {
	log(msg, slog.LevelInfo, args...)
}

func Warn(msg string, args ...any) {
	log(msg, slog.LevelWarn, args...)
}

func Error(msg string, args ...any) {
	log(msg, slog.LevelError, args...)
}

func Fatal(msg string, args ...any) {
	log(msg, slog.LevelError, args...)
	//nolint:revive // exit on fatal log
	os.Exit(1)
}

func log(msg string, level slog.Level, args ...any) {
	if globLogger == nil {
		InitGlobalLogger()
	}

	switch level {
	case slog.LevelDebug:
		globLogger.Debug(msg, args...)
	case slog.LevelInfo:
		globLogger.Info(msg, args...)
	case slog.LevelWarn:
		globLogger.Warn(msg, args...)
	case slog.LevelError:
		globLogger.Error(msg, args...)
	}
}
