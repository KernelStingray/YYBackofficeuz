package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

type contextKey string

const TraceIDKey contextKey = "trace_id"

// ContextHandler wraps a slog.Handler to inject trace_id from context
type ContextHandler struct {
	slog.Handler
}

// Handle adds the trace_id from the context to the log record's attributes.
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	return h.Handler.Handle(ctx, r)
}

func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(attrs)}
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(name)}
}

// InitLogger initializes the structured logger
func InitLogger(env string) *slog.Logger {
	var handler slog.Handler

	if env == "production" {
		// В продакшене используем JSON формат для логов
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		// В разработке используем текстовый формат с поддержкой дебага
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	logger := slog.New(ContextHandler{Handler: handler})
	slog.SetDefault(logger)

	return logger
}

// InitSentry initializes the Sentry SDK
func InitSentry(dsn string, environment string) error {
	if dsn == "" {
		slog.Warn("[WARNING] SENTRY_DSN is not set in .env! Errors will NOT be tracked.")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		slog.Error("Sentry initialization failed", "error", err)
		return err
	}
	slog.Info("Sentry initialized successfully")
	return nil
}

// FlushSentry flushes queued events to Sentry before shutdown
func FlushSentry(timeout time.Duration) {
	sentry.Flush(timeout)
}
