package logger

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// Logger struct holds the zap SugaredLogger.
type Logger struct {
	l *zap.SugaredLogger
}

// NewLogger initializes new Logger instance.
func NewLogger(_ context.Context, debug bool, errorOutputPaths []string, serviceName string) *Logger {
	config := zap.NewProductionConfig()
	config.ErrorOutputPaths = errorOutputPaths
	config.Level.SetLevel(zap.InfoLevel)
	if debug {
		config.Level.SetLevel(zap.DebugLevel)
	}

	l, err := config.Build(zap.AddCallerSkip(1), zap.Fields(zap.String("service", serviceName)))
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
func ToContext(ctx context.Context, logger *Logger) context.Context {
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

// WithContext adds trace_id and span_id to the logger if available.
func (l *Logger) WithContext(ctx context.Context) *zap.SugaredLogger {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return l.l
	}
	spanContext := span.SpanContext()
	if !spanContext.IsValid() {
		return l.l
	}

	return l.l.With(
		"trace_id", spanContext.TraceID().String(),
		"span_id", spanContext.SpanID().String(),
	)
}

// Infow logs informational message with structured key-value pairs.
func Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	FromContext(ctx).WithContext(ctx).Infow(msg, keysAndValues...)
}

// Errorw logs error message with structured key-value pairs.
func Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	FromContext(ctx).WithContext(ctx).Errorw(msg, keysAndValues...)
}

// Debugw logs debug message with structured key-value pairs.
func Debugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	FromContext(ctx).WithContext(ctx).Debugw(msg, keysAndValues...)
}

// Sync flushes any buffered log entries to the output.
func (l *Logger) Sync() error {
	return globalLogger.l.Sync()
}
