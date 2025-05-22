package domain

import (
	"context"
	"database/sql"
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

type GetListWarehouseRequest struct {
	Page      int64  `query:"page"`
	Limit     int64  `query:"limit"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
	Active    bool   `query:"active"`
}

type WarehouseUpdateStatusRequest struct {
	Active bool `json:"active"`
}

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	GetByShopID(ctx context.Context, shopID int64) ([]Warehouse, error)
	GetByID(ctx context.Context, id int64) (Warehouse, error)
	GetListWarehouse(ctx context.Context, shopID int64, param GetListWarehouseRequest) ([]Warehouse, error)
	GetListWarehouseCount(ctx context.Context, shopID int64, param GetListWarehouseRequest) (int64, error)
	UpdateStatus(ctx context.Context, id int64, active bool, tx *sql.Tx) error

	WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error
}

type WarehouseService interface {
	Create(ctx context.Context, shopID int64, req *WarehouseCreateRequest) (*Warehouse, error)
	GetByShopID(ctx context.Context, shopID int64) ([]Warehouse, error)
	GetListWarehouse(ctx context.Context, shopID int64, param GetListWarehouseRequest) ([]Warehouse, Metadata, error)
	UpdateStatus(ctx context.Context, id, shopID int64, active WarehouseUpdateStatusRequest) error
}
