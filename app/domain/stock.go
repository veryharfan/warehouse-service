package domain

import (
	"context"
	"time"
)

type Stock struct {
	ID          int64     `json:"id"`
	ProductID   int64     `json:"product_id"`
	WarehouseID int64     `json:"warehouse_id"`
	Quantity    int64     `json:"quantity"`
	Reserved    int64     `json:"reserved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StockCreateRequest struct {
	ShopID    int64 `json:"shop_id" validate:"required"`
	ProductID int64 `json:"product_id" validate:"required"`
}

type ProductQuantityOnWarehouse struct {
	WarehouseID int64 `json:"warehouse_id"`
	Quantity    int64 `json:"quantity"`
}

type StockResponse struct {
	ProductID   int64 `json:"product_id"`
	WarehouseID int64 `json:"warehouse_id"`
	Quantity    int64 `json:"quantity"`
	Reserved    int64 `json:"reserved"`
	Available   int64 `json:"available"`
}

type StockRepository interface {
	Create(ctx context.Context, stock []Stock) error
	GetByProductID(ctx context.Context, productID int64) ([]Stock, error)
}

type StockService interface {
	InitStock(ctx context.Context, req StockCreateRequest) ([]Stock, error)
	GetByProductID(ctx context.Context, productID int64) ([]StockResponse, error)
}
