package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var (
	globalLogger *Logger
	once         sync.Once
)

type Logger struct {
	l *zap.SugaredLogger
}

// NewLogger initializes new Logger instance.
func NewLogger(_ context.Context, debug bool, errorOutputPaths []string) *Logger {
	config := zap.NewProductionConfig()
	config.ErrorOutputPaths = errorOutputPaths
	config.Level.SetLevel(zap.InfoLevel)
	if debug {
		config.Level.SetLevel(zap.DebugLevel)
	}

	l, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

	once.Do(func() {
		globalLogger = &Logger{l: l.Sugar()}
	})

	return globalLogger
}

type ctxKey struct{}

// ToContext embeds Logger instance into given context.
func (l *Logger) ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext extracts Logger instance from given context.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok && l != nil {
		return l
	}
	if globalLogger == nil {
		panic("global logger is nil")
	}
	return globalLogger
}

// Infow logs informational message with structured key-value pairs.
func (l *Logger) Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	FromContext(ctx).l.Infow(msg, keysAndValues...)
}

// Errorw logs error message with structured key-value pairs.
func (l *Logger) Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	FromContext(ctx).l.Errorw(msg, keysAndValues...)
}

// Debugw logs debug message with structured key-value pairs.
func (l *Logger) Debugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	FromContext(ctx).l.Debugw(msg, keysAndValues...)
}

// Sync flushes any buffered log entries to the output.
func (l *Logger) Sync() error {
	return globalLogger.l.Sync()
}
