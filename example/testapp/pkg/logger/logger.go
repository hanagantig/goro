package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger interface {
	Debug(string, ...any)
	DebugContext(context.Context, string, ...any)
	Info(string, ...any)
	InfoContext(context.Context, string, ...any)
	Error(string, ...any)
	ErrorContext(context.Context, string, ...any)
	Warn(string, ...any)
	WarnContext(context.Context, string, ...any)
}

func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, nil)

	return slog.New(handler)
}
