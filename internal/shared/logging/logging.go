package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerCtxKey struct{}

var defaultLogger = zap.New(
	zapcore.NewCore(
		zapcore.NewJSONEncoder(
			zapcore.EncoderConfig{
				MessageKey:     "message",
				LevelKey:       "level",
				TimeKey:        "timestamp",
				NameKey:        "name",
				CallerKey:      "caller",
				FunctionKey:    "function",
				StacktraceKey:  "stacktrace",
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.RFC3339TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zapcore.InfoLevel),
	),
)

// FromContext returns the configured logger or the default logger if not configured.
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerCtxKey{}).(*zap.Logger); ok {
		return logger
	}
	return defaultLogger
}

// With adds the passed logger to the context.
func With(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

// WithOptions allows to customise the logger with extra zap options.
func WithOptions(ctx context.Context, opts ...zap.Option) context.Context {
	if len(opts) == 0 {
		return ctx
	}
	return With(ctx, FromContext(ctx).WithOptions(opts...))
}

// Info shortcut for zap's Info method.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// Error shortcut for zap's Info method.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}

// Debug shortcut for zap's Info method.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// Warn shortcut for zap's Info method.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// Fatal shortcut for zap's Fatal method.
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Fatal(msg, fields...)
}
