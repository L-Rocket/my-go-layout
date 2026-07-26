// Package grpc is the gRPC transport. It serves no business RPCs yet —
// only server reflection, so the port is real and grpcurl-explorable.
// Contracts will live in api/proto (see its README).
package grpc

import (
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewServer builds a grpc.Server with the logging interceptor and
// reflection registered. Kept tiny on purpose: services get registered
// here as protos land.
func NewServer(log *slog.Logger) *grpc.Server {
	srv := grpc.NewServer(grpc.UnaryInterceptor(LoggingInterceptor(log)))
	reflection.Register(srv)
	return srv
}
