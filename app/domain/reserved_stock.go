package domain

import (
	"context"
	"database/sql"
	"time"
)

type ReservedStockStatus string

const (
	ReservedStockStatusActive    ReservedStockStatus = "active"
	ReservedStockStatusCompleted ReservedStockStatus = "completed"
	ReservedStockStatusCancelled ReservedStockStatus = "cancelled"
)

type ReservedStock struct {
	ID        int64
	StockID   int64
	Quantity  int64
	OrderID   int64
	Status    ReservedStockStatus // "active", "completed", "cancelled"
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReservedStockCreateRequest struct {
	ProductID int64 `json:"product_id" validate:"required"`
	Quantity  int64 `json:"quantity" validate:"required"`
	OrderID   int64 `json:"order_id" validate:"required"`
}

type ReservedStockUpdateRequest struct {
	Status ReservedStockStatus `json:"status" validate:"required,oneof=active completed cancelled"`
}

type ReservedStockRepository interface {
	CreateReservedStock(ctx context.Context, rs *ReservedStock, tx *sql.Tx) error
	GetReservedStockByOrderID(ctx context.Context, orderID int64) (ReservedStock, error)
	GetReservedStocksByStockIDAndStatus(ctx context.Context, stockID int64, status ReservedStockStatus) ([]ReservedStock, error)
	GetTotalReservedStockByStockIDAndStatus(ctx context.Context, stockID int64, status ReservedStockStatus) (int64, error)
	UpdateReservedStockStatus(ctx context.Context, id int64, status ReservedStockStatus) error
	GetTotalReservedStockByStockIDsAndStatus(ctx context.Context, stockIDs []int64, status ReservedStockStatus) (map[int64]int64, error)
}

type ReservedStockUsecase interface {
	CreateReservedStock(ctx context.Context, req ReservedStockCreateRequest) error
	UpdateReservedStockStatusByOrderID(ctx context.Context, orderID int64, req ReservedStockUpdateRequest) error
}
