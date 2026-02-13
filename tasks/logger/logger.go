package logger

import (
	"context"

	"go.uber.org/zap"
)

var log *zap.Logger

func Init() {
	log, _ = zap.NewProduction()
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	log.With(extractRequestID(ctx)...).Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	log.With(extractRequestID(ctx)...).Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	log.With(extractRequestID(ctx)...).Error(msg, append(fields, zap.Error(err))...)
}

// ZapUint and ZapError are helpers to create zap fields for use outside the logger package
func ZapUint(key string, v uint) zap.Field { return zap.Uint64(key, uint64(v)) }
func ZapError(err error) zap.Field         { return zap.Error(err) }

func extractRequestID(ctx context.Context) []zap.Field {
	if reqID, ok := ctx.Value("requestID").(string); ok {
		return []zap.Field{zap.String("request_id", reqID)}
	}
	return nil
}
