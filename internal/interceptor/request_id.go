package interceptor

import (
	"context"

	"github.com/cathudson/order-service/internal/requestid"
	"github.com/cathudson/order-service/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestIDInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestID := extractOrDefault(ctx)
	ctx = requestid.WithRequestID(ctx, requestID)
	return handler(ctx, req)
}

func extractOrDefault(ctx context.Context) string {
	var requestID string
	incomingMD, exists := metadata.FromIncomingContext(ctx)
	if exists {
		if values := incomingMD.Get("x-request-id"); len(values) > 0 {
			requestID = values[0]
		}
	}
	if requestID == "" {
		requestID = util.UUID().String()
	}
	return requestID
}
