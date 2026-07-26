package grpc

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs every unary call (method, code, latency).
// Registered in NewServer — with no business RPCs yet it only fires for
// reflection calls, which is exactly how you confirm it works.
func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		log.Info("grpc call",
			"method", info.FullMethod,
			"code", status.Code(err).String(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
		return resp, err
	}
}
