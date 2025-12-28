package lf

import (
	"log/slog"
	"os"
)

const (
	LevelTrace = slog.Level(-8)
)

func newLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	// return slog.Default()
}
