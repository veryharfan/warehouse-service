package domain

import "context"

type StockMessage struct {
	ProductID int64 `json:"product_id"`
	Available int64 `json:"available"`
}

type BrokerPublisher interface {
	PublishStockAvailable(ctx context.Context, data StockMessage) error
}
