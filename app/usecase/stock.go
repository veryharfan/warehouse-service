package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"warehouse-service/app/domain"
	"warehouse-service/config"
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

func (u *stockUsecase) UpdateQuantity(ctx context.Context, id, quantity, shopID int64) error {
	stock, err := u.stockRepo.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "getStock", err)
		return err
	}

	warehouse, err := u.warehouseRepo.GetByID(ctx, stock.WarehouseID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "getWarehouse", err)
		return err
	}

	if warehouse.ShopID != shopID {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "invalidShopID", "shopID unauthorized")
		return domain.ErrUnauthorized
	}

	if stock.Quantity == quantity {
		slog.InfoContext(ctx, "[stockUsecase] UpdateQuantity", "noChange", nil)
		return nil
	}

	availableStock, err := u.stockRepo.GetAvailableStockByProductID(ctx, stock.ProductID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "getAvailableStock", err)
		return err
	}

	if err = u.stockRepo.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err := u.stockRepo.UpdateQuantity(ctx, id, quantity, stock.Version, tx)
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
	}); err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] UpdateQuantity", "transactionError", err)
		return err
	}

	slog.InfoContext(ctx, "[stockUsecase] UpdateQuantity", "quantityUpdated", quantity)
	return nil
}

func (u *stockUsecase) GetListStock(ctx context.Context, shopID int64, param domain.GetListStockRequest) ([]domain.Stock, domain.Metadata, error) {
	var metadata domain.Metadata

	stocks, err := u.stockRepo.GetListStock(ctx, shopID, param)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] GetListStock", "getListStock", err)
		return nil, metadata, err
	}

	count, err := u.stockRepo.GetListStockCount(ctx, shopID, param)
	if err != nil {
		slog.ErrorContext(ctx, "[stockUsecase] GetListStock", "getListStockCount", err)
		return nil, metadata, err
	}

	if len(stocks) == 0 {
		slog.InfoContext(ctx, "[stockUsecase] GetListStock", "noStocksFound", nil)
		return nil, metadata, domain.ErrNotFound
	}

	metadata = domain.Metadata{
		TotalData: count,
		TotalPage: (count + param.Limit - 1) / param.Limit,
		Page:      param.Page,
		Limit:     param.Limit,
		SortBy:    param.SortBy,
		SortOrder: param.SortOrder,
	}

	return stocks, metadata, nil
}
