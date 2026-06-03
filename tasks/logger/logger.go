package logger

import (
	"context"

	"go.uber.org/zap"
)

// Logger wraps zap for request-scoped structured logging.
type Logger struct {
	log *zap.Logger
}

func New(l *zap.Logger) *Logger {
	return &Logger{log: l}
}

func NewProduction() (*Logger, error) {
	z, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return New(z), nil
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.With(extractRequestID(ctx)...).Info(msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.With(extractRequestID(ctx)...).Warn(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	l.log.With(extractRequestID(ctx)...).Error(msg, append(fields, zap.Error(err))...)
}

func ZapUint(key string, v uint) zap.Field { return zap.Uint64(key, uint64(v)) }
func ZapError(err error) zap.Field         { return zap.Error(err) }

func extractRequestID(ctx context.Context) []zap.Field {
	if reqID, ok := ctx.Value("requestID").(string); ok {
		return []zap.Field{zap.String("request_id", reqID)}
	}
	return nil
}
