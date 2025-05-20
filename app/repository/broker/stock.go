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

func (s *stockBroker) PublishStockAvailable(ctx context.Context, data domain.StockMessage) error {

	msg, err := json.Marshal(data)
	if err != nil {
		slog.ErrorContext(ctx, "[stockBroker] PublishStockAvailable", "json.Marshal", err)
	}

	if _, err = s.js.Publish(ctx, "stock.available", msg); err != nil {
		slog.ErrorContext(ctx, "[stockBroker] PublishStockAvailable", "Publish", err)
		return err
	}

	slog.InfoContext(ctx, "[stockBroker] PublishStockAvailable", "message", msg)
	return nil
}
