package logger

import (
	"context"
	"fmt"
	"log/slog"
)

type HandlerMiddleware struct {
	log slog.Handler
}

func NewHandlerMiddleware(log slog.Handler) *HandlerMiddleware {
	return &HandlerMiddleware{
		log: log,
	}
}

func (h *HandlerMiddleware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.log.Enabled(ctx, rec)
}

func (h *HandlerMiddleware) Handle(ctx context.Context, rec slog.Record) error {
	if c, ok := ctx.Value(ctxLoggerKey).(logCtx); ok {
		if c.sessionId != "" {
			rec.Add(slog.String("session_id", fmt.Sprintf("%.4s", c.sessionId)))
		}
		if c.name != "" {
			rec.Add(slog.String("name", c.name))
		}
	}
	return h.log.Handle(ctx, rec)
}

func (h *HandlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.log.WithAttrs(attrs)
}

func (h *HandlerMiddleware) WithGroup(name string) slog.Handler {
	return h.log.WithGroup(name)
}
