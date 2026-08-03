// Package observability provides shared logging helpers for all services.
package observability

import (
	"log/slog"
	"os"
)

// NewLogger returns a JSON logger. Level is read from LOG_LEVEL
// (debug|info|warn|error), defaults to info.
func NewLogger() *slog.Logger {
	var level slog.Level = slog.LevelInfo
	if raw := os.Getenv("LOG_LEVEL"); raw != "" {
		_ = level.UnmarshalText([]byte(raw))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
