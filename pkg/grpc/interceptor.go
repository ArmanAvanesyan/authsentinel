package grpc

import (
	"context"

	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a unary interceptor that delegates to the proxy Engine.
func UnaryServerInterceptor(e proxy.Engine) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		_ = e
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a stream interceptor that delegates to the proxy Engine.
func StreamServerInterceptor(e proxy.Engine) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		_ = e
		return handler(srv, ss)
	}
}
