package domain

import (
	"context"
	"database/sql"
	"time"
)

type TransferStatus string

const (
	TransferStatusNotStarted TransferStatus = "not_started"
	TransferStatusInProgress TransferStatus = "in_progress"
	TransferStatusReverted   TransferStatus = "reverted"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
)

type StockTransfer struct {
	ID            int64          `json:"id"`
	ProductID     int64          `json:"product_id"`
	FromWarehouse int64          `json:"from_warehouse"`
	ToWarehouse   int64          `json:"to_warehouse"`
	Quantity      int64          `json:"quantity"`
	Status        TransferStatus `json:"status"` // "not_started", "in_progress", "reverted", "completed", "failed"
	Description   string         `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type StockTransferCreateRequest struct {
	ProductID     int64  `json:"product_id" validate:"required"`
	FromWarehouse int64  `json:"from_warehouse" validate:"required"`
	ToWarehouse   int64  `json:"to_warehouse" validate:"required"`
	Quantity      int64  `json:"quantity" validate:"required"`
	Description   string `json:"description"`
}

type StockTransferUpdateRequest struct {
	Status      TransferStatus `json:"status" validate:"required,oneof=not_started in_progress reverted completed failed"`
	Description string         `json:"description"`
}

type GetListStockTransferRequest struct {
	Page      int64  `query:"page"`
	Limit     int64  `query:"limit"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
	Status    string `query:"status"`
}

type StockTransferRepository interface {
	Create(ctx context.Context, transfer *StockTransfer) error
	GetByID(ctx context.Context, id int64) (StockTransfer, error)
	UpdateStatus(ctx context.Context, st StockTransfer, tx *sql.Tx) error
	GetListStockTransfer(ctx context.Context, shopID int64, param GetListStockTransferRequest) ([]StockTransfer, error)
	GetListStockTransferCount(ctx context.Context, shopID int64, param GetListStockTransferRequest) (int64, error)

	WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error
}

type StockTransferUsecase interface {
	CreateTransfer(ctx context.Context, shopID int64, transfer StockTransferCreateRequest) (*StockTransfer, error)
	GetTransferByID(ctx context.Context, id int64, shopID *int64) (StockTransfer, error)
	UpdateTransferStatus(ctx context.Context, id int64, req StockTransferUpdateRequest) error
	GetListStockTransfer(ctx context.Context, shopID int64, param GetListStockTransferRequest) ([]StockTransfer, Metadata, error)
}
