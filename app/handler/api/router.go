package handler

import (
	"warehouse-service/app/middleware"
	"warehouse-service/config"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, handler *WarehouseHandler, cfg *config.Config) {

	warehouseAPIGroup := app.Group("/warehouse-service")

	warehouseAPIGroup.Use(middleware.AuthInternal(cfg))

	warehouseAPIGroup.Post("/warehouse", handler.Create)
	warehouseAPIGroup.Get("/warehouse/:shop_id", handler.GetByShopID)
}
