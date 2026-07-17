package logger

import (
	"log/slog"
	"os"
)

func InitialiseLogger() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: nil})

	logger := slog.New(handler)

	// set as the default logger
	slog.SetDefault(logger)
}
