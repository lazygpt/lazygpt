//

package log

import "context"

// loggerKey is the key used to store the logger in the context.
type loggerKey struct{}

// AlwaysFromContext returns the logger from the context. If there is no logger
// in the context or the context is `nil`, it returns a new logger.
func AlwaysFromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return NewLogger("unknown")
	}

	logger := FromContext(ctx)
	if logger == nil {
		return NewLogger("unknown")
	}

	return logger
}

// FromContext returns the logger from the context. If there is no logger in
// the context or the context is `nil`, it returns `nil`.
func FromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return nil
	}

	logger, ok := ctx.Value(loggerKey{}).(*Logger)
	if !ok {
		return nil
	}

	return logger
}

// NewContext returns a new context with the logger.
func NewContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// ContextWithName returns a new context with the logger and the name.
func ContextWithName(ctx context.Context, name string) context.Context {
	logger := FromContext(ctx)
	if logger == nil {
		logger = NewLogger(name)
	} else {
		logger.Logger = logger.Named(name)
	}

	return NewContext(ctx, logger)
}

// ContextWithValues returns a new context with the logger and the key/value
// pairs.
func ContextWith(ctx context.Context, keyvals ...any) context.Context {
	logger := AlwaysFromContext(ctx)
	logger.Logger = logger.With(keyvals...)

	return NewContext(ctx, logger)
}

// ContextWithNameAndValues returns a new context with the logger, the name
// and the key/value pairs.
func ContextWithNameAndValues(
	ctx context.Context,
	name string,
	keyvals ...any,
) context.Context {
	logger := AlwaysFromContext(ctx)
	logger.Logger = logger.Named(name).With(keyvals...)

	return NewContext(ctx, logger)
}
