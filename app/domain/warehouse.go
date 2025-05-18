package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	ShopID    uuid.UUID `json:"shop_id"`
	Name      string    `json:"name"`
	Location  *string   `json:"location"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WarehouseCreateRequest struct {
	ShopID   uuid.UUID `json:"shop_id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	Location *string   `json:"location" validate:"required"`
}

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	GetByShopID(ctx context.Context, shopID string) ([]Warehouse, error)
}

type WarehouseService interface {
	Create(ctx context.Context, req *WarehouseCreateRequest) (*Warehouse, error)
	GetByShopID(ctx context.Context, shopID string) ([]Warehouse, error)
}
