package domain

import (
	"context"
	"database/sql"
	"time"
)

type Stock struct {
	ID          int64     `json:"id"`
	ProductID   int64     `json:"product_id"`
	WarehouseID int64     `json:"warehouse_id"`
	Quantity    int64     `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StockCreateRequest struct {
	ShopID    int64 `json:"shop_id" validate:"required"`
	ProductID int64 `json:"product_id" validate:"required"`
}

type UpdateQuantityRequest struct {
	Quantity int64 `json:"quantity"`
}

type StockResponse struct {
	ProductID   int64 `json:"product_id"`
	WarehouseID int64 `json:"warehouse_id"`
	Quantity    int64 `json:"quantity"`
	Reserved    int64 `json:"reserved"`
	Available   int64 `json:"available"`
}

type GetListStockRequest struct {
	ProductID   int64  `query:"product_id"`
	WarehouseID int64  `query:"warehouse_id"`
	Page        int64  `query:"page"`
	Limit       int64  `query:"limit"`
	SortOrder   string `query:"sort_order"`
	SortBy      string `query:"sort_by"`
}

type Metadata struct {
	TotalData int64  `json:"total_data"`
	TotalPage int64  `json:"total_page"`
	Page      int64  `json:"page"`
	Limit     int64  `json:"limit"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

type StockReservation struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type AvailableStock struct {
	ProductID      int64 `json:"product_id"`
	AvailableStock int64 `json:"available_stock"`
}

type StockRepository interface {
	Create(ctx context.Context, stock []Stock) error
	GetByProductID(ctx context.Context, productID int64) ([]Stock, error)
	GetByID(ctx context.Context, id int64) (Stock, error)
	UpdateQuantity(ctx context.Context, id, quantity int64, tx *sql.Tx) error
	GetAvailableStockByProductID(ctx context.Context, productID int64) (int64, error)
	GetByProductIDAndWarehouseID(ctx context.Context, productID, warehouseID int64) (Stock, error)
	GetListStock(ctx context.Context, shopID int64, param GetListStockRequest) ([]Stock, error)
	GetListStockCount(ctx context.Context, shopID int64, param GetListStockRequest) (int64, error)
	GetByWarehouseID(ctx context.Context, warehouseID int64) ([]Stock, error)
	GetAvailableStockByProductIDs(ctx context.Context, productIDs []int64) (map[int64]int64, error)
	// UpdateReservedStocks(ctx context.Context, id, reservedQuantity, version int64, tx *sql.Tx) error

	LockForUpdate(ctx context.Context, id int64, tx *sql.Tx) (Stock, error)
	WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error
}

type StockService interface {
	InitStock(ctx context.Context, req StockCreateRequest) ([]Stock, error)
	GetAvailableStockByProductID(ctx context.Context, productID int64) (AvailableStock, error)
	UpdateQuantity(ctx context.Context, id, shopID int64, req UpdateQuantityRequest) error
	GetListStock(ctx context.Context, shopID int64, param GetListStockRequest) ([]Stock, Metadata, error)
}
