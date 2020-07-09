package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
)

// LogFunc represents a function that takes a context,
// format string and number of interface{} and logs accordingly.
type LogFunc func(context.Context, string, ...interface{})

// GetStandardLogger returns a logger that is used for debugging.
// It logs to STDERR, has no prefix, and logs the date and time in UTC.
func GetStandardLogger() *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags|log.LUTC)
}

// GetCheckerLogger returns a logger designed for the CI checker.
// It logs to STDOUT, has no prefix, and does not log any metadata.
func GetCheckerLogger() *log.Logger {
	return log.New(os.Stdout, "", 0)
}

// GetStandardEntries gets a slice of []ContextEntry given
// a prefix and *log.Logger. All LogFuncs will default to StandardDebugf
// since they are not included.
func GetStandardEntries(prefix string, logger *log.Logger) []ContextEntry {
	return []ContextEntry{
		{
			Key:   LoggerPrefix,
			Value: prefix,
		},
		{
			Key:   Logger,
			Value: logger,
		},
		{
			Key:   Warn,
			Value: LogFunc(SentryWarnf),
		},
	}
}

// GetCheckerEntries gets a slice of []ContextEntry given
// a prefix and *log.Logger. The Debug LogFunc will default to
// StandardDebugf since it is not included.
func GetCheckerEntries(prefix string, logger *log.Logger) []ContextEntry {
	return []ContextEntry{
		{
			Key:   LoggerPrefix,
			Value: prefix,
		},
		{
			Key:   Logger,
			Value: logger,
		},
		{
			Key:   Warn,
			Value: LogFunc(CheckerWarnf),
		},
		{
			Key:   Err,
			Value: LogFunc(CheckerErrf),
		},
	}
}

// SentryWarnf implements LogFunc, notifying sentry of an error.
func SentryWarnf(ctx context.Context, format string, v ...interface{}) {
	sentry.CurrentHub().RecoverWithContext(ctx, fmt.Errorf(format, v...))
	sentry.CurrentHub().Flush(SentryFlushTime)
}

// Printf is a LogFunc that uses a logger to log a formatted string.
func Printf(ctx context.Context, format string, v ...interface{}) {
	if logger, ok := ctx.Value(Logger).(*log.Logger); ok && logger != nil {
		if prefix, ok := ctx.Value(LoggerPrefix).(string); ok {
			logger.Printf(prefix+": "+format, v...)
		} else {
			logger.Printf(format, v...)
		}
	} else {
		panic("logger does not exist")
	}
}

// StandardDebugf is a LogFunc that calls Printf if the program is in DEBUG mode.
func StandardDebugf(ctx context.Context, format string, v ...interface{}) {
	if IsDebug() {
		Printf(ctx, format, v...)
	}
}

// Generic function used to call a LogFunc stored in the context using a ContextKey key,
// calling a default LogFunc if the key is not set.
func logf(ctx context.Context, key ContextKey, defaultLogf LogFunc, format string, v ...interface{}) {
	if f, ok := ctx.Value(key).(LogFunc); ok && f != nil {
		f(ctx, format, v...)
	} else {
		defaultLogf(ctx, format, v...)
	}
}

// Debugf is a LogFunc that attempts to call the Debug LogFunc in the context, defaulting
// to StandardDebugf if unset.
func Debugf(ctx context.Context, format string, v ...interface{}) {
	logf(ctx, Debug, StandardDebugf, format, v...)
}

// Warnf is a LogFunc that attempts to call the Warn LogFunc in the context, defaulting
// to StandardDebugf if unset.
func Warnf(ctx context.Context, format string, v ...interface{}) {
	logf(ctx, Warn, StandardDebugf, format, v...)
}

// Errf is a LogFunc that attempts to call the Err LogFunc in the context, defaulting
// to StandardDebugf if unset.
func Errf(ctx context.Context, format string, v ...interface{}) {
	logf(ctx, Err, StandardDebugf, format, v...)
}

// Used to determine the type of the checker's log output.
type checkerLogType string

const (
	checkerErr  checkerLogType = "error"
	checkerWarn checkerLogType = "warning"
)

// Generic function used to output a checkerLogType to STDOUT for the CI checker.
func checkerLogf(ctx context.Context, logType checkerLogType, format string, v ...interface{}) {
	if logger, ok := ctx.Value(Logger).(*log.Logger); ok && logger != nil {
		if prefix, ok := ctx.Value(LoggerPrefix).(string); ok {
			logger.Printf("::%s file=%s,line=1,col=1::%s\n", logType, prefix, escapeGitHub(fmt.Sprintf(format, v...)))
		} else {
			panic("logger prefix does not exist")
		}
	}
}

// CheckerErrf outputs an error to STDOUT for the CI checker.
func CheckerErrf(ctx context.Context, format string, v ...interface{}) {
	checkerLogf(ctx, checkerErr, format, v...)
}

// CheckerWarnf outputs a warning to STDOUT for the CI checker.
func CheckerWarnf(ctx context.Context, format string, v ...interface{}) {
	checkerLogf(ctx, checkerWarn, format, v...)
}

// escape characters
func escapeGitHub(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\n", "%0A")
	s = strings.ReplaceAll(s, "\r", "%0D")
	return s
}
