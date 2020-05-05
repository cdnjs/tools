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

func Debugf(ctx context.Context, format string, v ...interface{}) {
	if IsDebug() {
		Printf(ctx, format, v...)
	}
}

func SetLoggerFlag(f int) {
	logger.SetFlags(f)
}
