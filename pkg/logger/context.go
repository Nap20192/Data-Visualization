package logger

import (
	"context"
)

type contextKey struct{}

var ctxLoggerKey = contextKey{}

type logCtx struct {
	sessionId string
	name      string
}

func WithSessionId(ctx context.Context, sessionId string) context.Context {
	if c, ok := ctx.Value(ctxLoggerKey).(logCtx); ok {
		c.sessionId = sessionId
		return context.WithValue(ctx, ctxLoggerKey, c)
	}

	return context.WithValue(ctx, ctxLoggerKey, logCtx{
		sessionId: sessionId,
	})
}

func WithName(ctx context.Context, name string) context.Context {
	if c, ok := ctx.Value(ctxLoggerKey).(logCtx); ok {
		c.name = name
		return context.WithValue(ctx, ctxLoggerKey, c)
	}
	return context.WithValue(ctx, ctxLoggerKey, logCtx{
		name: name,
	})
}
