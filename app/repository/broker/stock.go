package broker

import (
	"context"
	"encoding/json"
	"log/slog"
	"warehouse-service/app/domain"

	"github.com/nats-io/nats.go/jetstream"
)

type stockBroker struct {
	js jetstream.JetStream
}

func NewStockBrokerPublisher(stream jetstream.JetStream) domain.BrokerPublisher {
	return &stockBroker{
		js: stream,
	}
}

func (s *stockBroker) StockInit(ctx context.Context, data domain.StockMessage) error {

	msg, err := json.Marshal(data)
	if err != nil {
		slog.ErrorContext(ctx, "[stockBroker] StockInit", "json.Marshal", err)
	}

	if _, err = s.js.Publish(ctx, "stock.init", msg); err != nil {
		slog.ErrorContext(ctx, "[stockBroker] StockInit", "Publish", err)
		return err
	}

	slog.InfoContext(ctx, "[stockBroker] StockInit", "message", msg)
	return nil
}
