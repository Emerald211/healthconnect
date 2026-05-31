package logger

import (
	"log/slog"
	"os"
)

func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: env != "production",
	}
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
