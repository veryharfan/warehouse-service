package domain

import "context"

type StockMessage struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Reserved  int64 `json:"reserved"`
	Available int64 `json:"available"`
}

type BrokerPublisher interface {
	StockInit(ctx context.Context, data StockMessage) error
}
