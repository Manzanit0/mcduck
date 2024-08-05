package xtrace

import (
	"context"

	"connectrpc.com/connect"
	"github.com/manzanit0/mcduck/pkg/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func SpanEnhancerInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			email, ok := auth.GetInfo(ctx).(string)
			if !ok {
				return next(ctx, req)
			}

			span := trace.SpanFromContext(ctx)
			span.SetAttributes(attribute.String("mduck.user.email", email))

			return next(ctx, req)
		}
	}
}
