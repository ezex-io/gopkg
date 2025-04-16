package logger

import (
	"log/slog"
	"os"
)

func ExampleNewSlog() {
	log := NewSlog(WithTextHandler(os.Stdout, slog.LevelDebug))
	log.Info("foobar")

	// foobar  (or whatever your logger outputs)
}
