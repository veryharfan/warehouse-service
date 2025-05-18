package logger

import (
	"context"
	"log/slog"
	"warehouse-service/pkg/ctxutil"
)

type RequestIDHandler struct {
	slog.Handler
}

func (h *RequestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	requestID := ctxutil.GetRequestID(ctx)
	if requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}
	return h.Handler.Handle(ctx, r)
}
