package domain

import "time"

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
