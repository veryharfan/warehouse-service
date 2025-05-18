package middleware

import (
	"log/slog"
	"warehouse-service/pkg/ctxutil"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
)

func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			uuidV4, err := uuid.NewV4()
			if err != nil {
				slog.WarnContext(c.Context(), "[RequestIDMiddleware] Error generating UUID", "error", err)
			}
			reqID = uuidV4.String()
		}
		c.Locals(ctxutil.RequestIDKey, reqID)
		return c.Next()
	}
}
