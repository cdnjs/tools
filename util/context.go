package util

import (
	"context"
)

func ContextWithName(loggerPrefix string) context.Context {
	return context.WithValue(context.Background(), "loggerPrefix", loggerPrefix)
}

// Create a new context with our logging sink
// 100 logs are allowed to be buffered before blocking
func WithLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, "logs", make(chan interface{}, 100))
}
