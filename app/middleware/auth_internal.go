package middleware

import (
	"warehouse-service/app/domain"
	"warehouse-service/app/handler/api/response"
	"warehouse-service/config"

	"github.com/gofiber/fiber/v2"
)

type AuthInternalHeader string

const (
	AuthInternalHeaderKey AuthInternalHeader = "X-Internal-Auth"
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
