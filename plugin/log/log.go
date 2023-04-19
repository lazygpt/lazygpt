//

package log

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

// Logger is a wrapper around hclog.Logger to make it concrete.
type Logger struct {
	hclog.Logger
}

var _ hclog.Logger = (*Logger)(nil)

// NewLogger returns a new logger with the given name.
func NewLogger(name string) *Logger {
	return &Logger{
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:  name,
			Level: hclog.Trace,
		}),
	}
}

// WithName returns a new logger with the given name.
func (logger *Logger) WithName(name string) *Logger {
	return &Logger{
		Logger: logger.Logger.Named(name),
	}
}

// Trace is a convenience method for logging at the Trace level to the logger
// in the context.
func Trace(ctx context.Context, msg string, keyvals ...any) {
	AlwaysFromContext(ctx).Trace(msg, keyvals...)
}

// Debug is a convenience method for logging at the Debug level to the logger
// in the context.
func Debug(ctx context.Context, msg string, keyvals ...any) {
	AlwaysFromContext(ctx).Debug(msg, keyvals...)
}

// Info is a convenience method for logging at the Info level to the logger
// in the context.
func Info(ctx context.Context, msg string, keyvals ...any) {
	AlwaysFromContext(ctx).Info(msg, keyvals...)
}

// Warn is a convenience method for logging at the Warn level to the logger
// in the context.
func Warn(ctx context.Context, msg string, keyvals ...any) {
	AlwaysFromContext(ctx).Warn(msg, keyvals...)
}

// Error is a convenience method for logging at the Error level to the logger
// in the context. If an odd number of keyvals are provided, and the first one
// is an `error` type, the is added to the keyvals as `error`.
func Error(ctx context.Context, msg string, keyvals ...any) {
	if len(keyvals)%2 == 1 {
		if err, ok := keyvals[0].(error); ok {
			keyvals = keyvals[1:]
			keyvals = append(keyvals, "error", err)
		}
	}

	AlwaysFromContext(ctx).Error(msg, keyvals...)
}
