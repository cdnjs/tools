package util

import (
	"context"
	"log"
	"os"
)

var (
	flags  = log.LstdFlags | log.LUTC
	logger = log.New(os.Stderr, "", flags)
)

func Printf(ctx context.Context, format string, v ...interface{}) {
	if prefix, ok := ctx.Value("loggerPrefix").(string); ok {
		logger.Printf(prefix+": "+format, v...)
	} else {
		logger.Printf(format, v...)
	}
}

func ContextWithName(loggerPrefix string) context.Context {
	return context.WithValue(context.Background(), "loggerPrefix", loggerPrefix)
}
