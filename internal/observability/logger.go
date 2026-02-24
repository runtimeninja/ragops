package observability

import (
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	h := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(h).With("service", "ragops", "env", env)
}
