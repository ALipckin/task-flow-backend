package middleware

import (
	"TaskStorageService/logger"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, _ := metadata.FromIncomingContext(ctx)
		var reqID string
		if values := md.Get("x-request-id"); len(values) > 0 {
			reqID = values[0]
		} else {
			reqID = uuid.New().String()
		}

		ctx = context.WithValue(ctx, "requestID", reqID)
		logger.Info(ctx, "Incoming gRPC request", zap.String("method", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			logger.Error(ctx, "gRPC error", err)
		}

		return resp, err
	}
}
