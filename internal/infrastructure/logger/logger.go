// Package logger builds the app's one structured logger.
package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON slog logger at the named level ("debug", "info",
// "warn", "error"; anything else falls back to info). JSON because logs
// are meant to be scraped, not read at a terminal in production.
func New(level string) *slog.Logger {
	var lv slog.Level
	if err := lv.UnmarshalText([]byte(level)); err != nil {
		lv = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
}
