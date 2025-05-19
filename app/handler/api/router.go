package handler

import (
	"warehouse-service/app/middleware"
	"warehouse-service/config"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, warehousHandler *WarehouseHandler, stockHandler *StockHandler, cfg *config.Config) {

	api := app.Group("/warehouse-service").Use(middleware.Auth(cfg.Jwt.SecretKey))

	api.Post("/warehouses", warehousHandler.Create)
	api.Get("/shops/:shop_id/warehouses", warehousHandler.GetByShopID)

	api.Get("/products/:product_id/stocks", stockHandler.GetByProductID)

	internal := app.Group("/internal/warehouse-service").Use(middleware.AuthInternal(cfg))
	internal.Post("/stocks", stockHandler.Create)
}
