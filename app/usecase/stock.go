package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"warehouse-service/app/domain"
	"warehouse-service/config"
	"warehouse-service/pkg"
)

type stockUsecase struct {
	stockRepo          domain.StockRepository
	warehouseRepo      domain.WarehouseRepository
	stockPublishBroker domain.BrokerPublisher
	cfg                *config.Config
}

func NewStockUsecase(stockRepo domain.StockRepository, warehouseRepo domain.WarehouseRepository, stockPublishBroker domain.BrokerPublisher, cfg *config.Config) domain.StockService {
	return &stockUsecase{stockRepo, warehouseRepo, stockPublishBroker, cfg}
}

func (u *stockUsecase) InitStock(ctx context.Context, req domain.StockCreateRequest) ([]domain.Stock, error) {
	warehouses, err := u.warehouseRepo.GetByShopID(ctx, req.ShopID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] InitStock", "getWarehouses", err)
		return nil, err
	}

	var stocks []domain.Stock
	for _, warehouse := range warehouses {
		stocks = append(stocks, domain.Stock{
			ProductID:   req.ProductID,
			WarehouseID: warehouse.ID,
		})
	}
	err = u.stockRepo.Create(ctx, stocks)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] InitStock", "createStock", err)
		return nil, err
	}

	// Publish the stock available event to the broker
	err = u.stockPublishBroker.PublishStockAvailable(ctx, domain.StockMessage{
		ProductID: req.ProductID,
	})
	if err != nil {
		slog.WarnContext(ctx, "[stockUsecase] InitStock", "publishStockInit", err)
	}

	slog.InfoContext(ctx, "[stockUsecase] InitStock", "stocks", req)
	return stocks, nil
}

func (u *stockUsecase) GetByProductID(ctx context.Context, productID int64) ([]domain.StockResponse, error) {
	stocks, err := u.stockRepo.GetByProductID(ctx, productID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] GetByProductID", "getStocks", err)
		return nil, err
	}

	if len(stocks) == 0 {
		slog.InfoContext(ctx, "[stockUsecase] GetByProductID", "noStocksFound", nil)
		return nil, domain.ErrNotFound
	}

	var stockResponses []domain.StockResponse
	for _, stock := range stocks {
		stockResponse := domain.StockResponse{
			ProductID:   stock.ProductID,
			WarehouseID: stock.WarehouseID,
			Quantity:    stock.Quantity,
			Reserved:    stock.Reserved,
			Available:   stock.Quantity - stock.Reserved,
		}
		stockResponses = append(stockResponses, stockResponse)
	}

	return stockResponses, nil
}

func (u *stockUsecase) UpdateQuantity(ctx context.Context, id, quantity int64) error {
	stock, err := u.stockRepo.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "getStock", err)
		return err
	}

	availableStock, err := u.stockRepo.GetAvailableStockByProductID(ctx, stock.ProductID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "getAvailableStock", err)
		return err
	}

	tx, err := u.stockRepo.BeginTransaction(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "beginTransaction", err)
		return err
	}

	err = pkg.WithTransaction(tx, func() error {
		err := u.stockRepo.UpdateQuantity(ctx, id, quantity, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "updateStock", err)
			return err
		}

		change := quantity - stock.Quantity
		updatedStock := availableStock + change
		if updatedStock < 0 {
			slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "insufficientStock", nil)
			return fmt.Errorf("%w: insufficient stock", domain.ErrInvalidRequest)
		}

		err = u.stockPublishBroker.PublishStockAvailable(ctx, domain.StockMessage{
			ProductID: stock.ProductID,
			Available: updatedStock,
		})
		if err != nil {
			slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "publishStockAvailable", err)
			return err
		}
		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "transactionError", err)
		return err
	}

	slog.InfoContext(ctx, "[stockUsecase] UpdateQuantity", "quantityUpdated", quantity)
	return nil
}
