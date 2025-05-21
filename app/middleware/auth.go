package middleware

import (
	"log/slog"
	"warehouse-service/app/domain"
	"warehouse-service/app/handler/api/response"
	"warehouse-service/config"
	"warehouse-service/pkg"
	"warehouse-service/pkg/ctxutil"

	"github.com/gofiber/fiber/v2"
)

type AuthInternalHeader string

const (
	AuthInternalHeaderKey       AuthInternalHeader = "X-Internal-Auth"
	AuthWarehouseAdminHeaderKey AuthInternalHeader = "X-Warehouse-Admin-Auth"
)

func AuthInternal(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the auth header from the request
		authHeader := c.Get(string(AuthInternalHeaderKey))
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}
		// Check if the auth header is valid (you can implement your own logic here)
		if authHeader != cfg.InternalAuthHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}

		return c.Next()
	}
}

func AuthWarehouseAdmin(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the auth header from the request
		authHeader := c.Get(string(AuthWarehouseAdminHeaderKey))
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}
		// Check if the auth header is valid (you can implement your own logic here)
		if authHeader != cfg.WarehouseAdminAuthHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}

		return c.Next()
	}
}

func Auth(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		token, err := pkg.GetTokenFromHeaders(c.Get("Authorization"))
		if err != nil {
			slog.ErrorContext(c.Context(), "[middleware] Auth", "GetTokenFromHeaders", err)
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}

		claims, err := pkg.ParseJwtToken(token, secretKey)
		if err != nil {
			slog.ErrorContext(c.Context(), "[middleware] Auth", "ParseJwtToken", err)
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}

		if claims.UID == 0 {
			slog.ErrorContext(c.Context(), "[middleware] Auth", "userID", "0")
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}

		if claims.SID == nil {
			slog.ErrorContext(c.Context(), "[middleware] Auth", "shopID", "nil")
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error(domain.ErrUnauthorized))
		}

		c.Locals(ctxutil.UserIDKey, claims.UID)
		c.Locals(ctxutil.ShopIDKey, *claims.SID)
		return c.Next()
	}
}
