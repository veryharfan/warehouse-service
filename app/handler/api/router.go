package handler

import (
	"warehouse-service/app/middleware"
	"warehouse-service/config"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App,
	warehousHandler *WarehouseHandler,
	stockHandler *StockHandler,
	stockTransferHandler *StockTransferHandler,
	reservedStockHandler *ReservedStockHandler,
	cfg *config.Config) {

	api := app.Group("/warehouse-service", middleware.Auth(cfg.Jwt.SecretKey))
	internal := app.Group("/internal/warehouse-service", middleware.AuthInternal(cfg))
	warehouseAdmin := app.Group("/admin/warehouse-service", middleware.AuthWarehouseAdmin(cfg))

	// warehouses
	api.Post("/warehouses", warehousHandler.Create)
	api.Get("/shops/:shop_id/warehouses", warehousHandler.GetByShopID)
	api.Patch("/warehouses/:id/status", warehousHandler.UpdateStatus)

	// stocks
	api.Get("/stocks", stockHandler.GetListStock)
	api.Patch("/stocks/:id", stockHandler.UpdateQuantity)

	// internal stocks
	internal.Post("/stocks", stockHandler.Create)
	internal.Get("/products/:product_id/stocks", stockHandler.GetByProductID)

	// stock transfers
	api.Post("/stock-transfers", stockTransferHandler.Create)
	api.Get("/stock-transfers/:id", stockTransferHandler.GetByID)
	api.Get("/stock-transfers", stockTransferHandler.GetListStockTransfer)
	warehouseAdmin.Patch("/stock-transfers/:id", stockTransferHandler.UpdateStatus)

	// reserved stocks
	internal.Post("/reserved-stocks", reservedStockHandler.CreateReservedStock)
	internal.Patch("/orders/:order_id/reserved-stocks/status", reservedStockHandler.UpdateReservedStockStatus)

}
