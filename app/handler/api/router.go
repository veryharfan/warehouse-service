package handler

import (
	"warehouse-service/app/middleware"
	"warehouse-service/config"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, warehousHandler *WarehouseHandler, stockHandler *StockHandler, cfg *config.Config) {

	api := app.Group("/internal/warehouse-service")

	api.Use(middleware.AuthInternal(cfg))

	api.Post("/warehouses", warehousHandler.Create)
	api.Get("/shops/:shop_id/warehouses", warehousHandler.GetByShopID)

	api.Post("/stocks", stockHandler.Create)
	api.Get("/products/:product_id/stocks", stockHandler.GetByProductID)
}
