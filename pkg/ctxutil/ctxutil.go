package ctxutil

import (
	"context"
	"errors"
)

type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	UserIDKey    ctxKey = "user_id"
	ShopIDKey    ctxKey = "shop_id"
)

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

func GetUserIDCtx(ctx context.Context) (int64, error) {
	if v := ctx.Value(UserIDKey); v != nil {
		if id, ok := v.(int64); ok {
			return id, nil
		}
	}
	return 0, errors.New("user ID not found")
}

func GetShopIDCtx(ctx context.Context) (int64, error) {
	if v := ctx.Value(ShopIDKey); v != nil {
		if id, ok := v.(int64); ok {
			return id, nil
		}
	}
	return 0, errors.New("shop ID not found")
}
