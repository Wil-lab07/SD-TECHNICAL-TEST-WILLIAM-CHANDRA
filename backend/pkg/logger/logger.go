package logger

import (
	"log/slog"
	"os"
)

func Setup() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	
	logger := slog.New(handler)
	
	slog.SetDefault(logger)
	
	return logger
}
