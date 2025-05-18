package ctxutil

import "context"

type ctxKey string

const RequestIDKey ctxKey = "request_id"

func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, reqID)
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
