package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	flags  = log.LstdFlags | log.LUTC
	logger = log.New(os.Stderr, "", flags)
)

// Printf uses the logger to log a formatted string.
func Printf(ctx context.Context, format string, v ...interface{}) {
	if prefix, ok := ctx.Value("loggerPrefix").(string); ok {
		logger.Printf(prefix+": "+format, v...)
	} else {
		logger.Printf(format, v...)
	}
}

// Debugf calls Printf if the program is in DEBUG mode.
func Debugf(ctx context.Context, format string, v ...interface{}) {
	if IsDebug() {
		Printf(ctx, format, v...)
	}
}

// Warnf is used to output a warning, either to STDOUT in logger format for the CI checker
// or STDERR for debugging.
func Warnf(ctx context.Context, runFromChecker bool, format string, v ...interface{}) {
	if runFromChecker {
		CheckerWarn(ctx, fmt.Sprintf(format, v...))
	} else {
		Debugf(ctx, format, v...)
	}
}

// CheckerErr outputs an error to STDOUT for the CI checker.
func CheckerErr(ctx context.Context, s string) {
	if prefix, ok := ctx.Value("loggerPrefix").(string); ok {
		fmt.Printf("::error file=%s,line=1,col=1::%s\n", prefix, escapeGitHub(s))
	} else {
		panic("unreachable")
	}
}

// CheckerWarn outputs a warning to STDOUT for the CI checker.
func CheckerWarn(ctx context.Context, s string) {
	if prefix, ok := ctx.Value("loggerPrefix").(string); ok {
		fmt.Printf("::warning file=%s,line=1,col=1::%s\n", prefix, escapeGitHub(s))
	} else {
		panic("unreachable")
	}
}

// escape characters
func escapeGitHub(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\n", "%0A")
	s = strings.ReplaceAll(s, "\r", "%0D")
	return s
}

// SetLoggerFlag will update the logger's output flags.
func SetLoggerFlag(f int) {
	logger.SetFlags(f)
}
