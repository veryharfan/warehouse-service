package domain

import (
	"context"
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

type ReservedStockRepository interface {
	CreateReservedStock(ctx context.Context, stockID, quantity, orderID int64) (int64, error)
	GetReservedStockByOrderID(ctx context.Context, orderID int64) (ReservedStock, error)
	GetReservedStocksByStockIDAndStatus(ctx context.Context, stockID int64, status ReservedStockStatus) ([]ReservedStock, error)
	GetTotalReservedStockByStockIDAndStatus(ctx context.Context, stockID int64, status ReservedStockStatus) (int64, error)
	UpdateReservedStockStatus(ctx context.Context, id int64, status ReservedStockStatus) error
	GetTotalReservedStockByStockIDsAndStatus(ctx context.Context, stockIDs []int64, status ReservedStockStatus) (map[int64]int64, error)
}
