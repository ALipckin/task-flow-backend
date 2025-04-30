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

func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	log.With(extractRequestID(ctx)...).Error(msg, append(fields, zap.Error(err))...)
}

func extractRequestID(ctx context.Context) []zap.Field {
	if reqID, ok := ctx.Value("requestID").(string); ok {
		return []zap.Field{zap.String("request_id", reqID)}
	}
	return nil
}
