package logger

import (
	"context"
	"os"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

var (
	globalLogger *Logger
	once         sync.Once
	tmpLogger    = NewTmpLogger()
)

// Logger struct holds the zap SugaredLogger.
type Logger struct {
	l *zap.SugaredLogger
}

// NewTmpLogger
func NewTmpLogger() *Logger {
	return &Logger{l: zap.NewExample().Sugar()}
}

// NewLogger initializes new Logger instance.
func NewLogger(_ context.Context, debug bool, errorOutputPaths []string, serviceName string, graylogAddr *string) *Logger {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(zap.InfoLevel)
	if debug {
		config.Level.SetLevel(zap.DebugLevel)
	}

	// Build the console core
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	consoleWS := zapcore.Lock(os.Stdout)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWS, config.Level)

	var cores []zapcore.Core
	cores = append(cores, consoleCore)

	if graylogAddr != nil && *graylogAddr != "" {
		graylogWriter, err := gelf.NewUDPWriter(*graylogAddr)
		if err != nil {
			zap.L().Error("Failed to create Graylog writer", zap.Error(err))
		} else {
			graylogWS := zapcore.AddSync(graylogWriter)
			graylogEncoder := zapcore.NewJSONEncoder(config.EncoderConfig)
			graylogCore := zapcore.NewCore(graylogEncoder, graylogWS, config.Level)
			cores = append(cores, graylogCore)
		}
	}

	// Combine cores
	combinedCore := zapcore.NewTee(cores...)

	// Open error output paths
	errorOutput, _, err := zap.Open(errorOutputPaths...)
	if err != nil {
		panic(err)
	}

	// Build logger with combined core
	l := zap.New(
		combinedCore,
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("service", serviceName)),
		zap.ErrorOutput(errorOutput),
	)

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
		return tmpLogger
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

// GetZapLogger returns zap.Logger
func GetZapLogger() *zap.Logger {
	return globalLogger.GetZapLogger()
}

// GetZapLogger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.l.Desugar()
}
