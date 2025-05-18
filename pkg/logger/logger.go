package logger

import (
	"log/slog"
	"os"
)

func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(&RequestIDHandler{Handler: handler}))
}
