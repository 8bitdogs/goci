package core

import "context"

type ctxKey struct{}

func storeRequestID(ctx context.Context, requestID uint64) context.Context {
	return context.WithValue(ctx, ctxKey{}, requestID)
}

func RequestID(ctx context.Context) uint64 {
	return ctx.Value(ctxKey{}).(uint64)
}
