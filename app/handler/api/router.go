package handler

import (
	"warehouse-service/app/middleware"
	"warehouse-service/config"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, warehousHandler *WarehouseHandler, stockHandler *StockHandler, cfg *config.Config) {

	api := app.Group("/warehouse-service").Use(middleware.Auth(cfg.Jwt.SecretKey))

	// warehouses
	api.Post("/warehouses", warehousHandler.Create)
	api.Get("/shops/:shop_id/warehouses", warehousHandler.GetByShopID)

	// stocks
	api.Patch("/stocks/:id", stockHandler.UpdateQuantity)

	// internal
	internal := app.Group("/internal/warehouse-service").Use(middleware.AuthInternal(cfg))
	// internal stocks
	internal.Post("/stocks", stockHandler.Create)
	internal.Get("/products/:product_id/stocks", stockHandler.GetByProductID)
}
