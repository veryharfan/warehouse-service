package domain

import (
	"context"
	"time"
)

type Warehouse struct {
	ID        int64     `json:"id"`
	ShopID    int64     `json:"shop_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WarehouseCreateRequest struct {
	Name     string `json:"name" validate:"required"`
	Location string `json:"location" validate:"required"`
}

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	GetByShopID(ctx context.Context, shopID int64) ([]Warehouse, error)
}

type WarehouseService interface {
	Create(ctx context.Context, shopID int64, req *WarehouseCreateRequest) (*Warehouse, error)
	GetByShopID(ctx context.Context, shopID int64) ([]Warehouse, error)
}
