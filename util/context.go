package util

import (
	"context"
)

func ContextWithName(loggerPrefix string) context.Context {
	return context.WithValue(context.Background(), "loggerPrefix", loggerPrefix)
}
