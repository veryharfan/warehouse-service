package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"warehouse-service/app/domain"
	"warehouse-service/config"
)

type warehouseUsecase struct {
	warehouseRepo      domain.WarehouseRepository
	stockRepo          domain.StockRepository
	reservedStockRepo  domain.ReservedStockRepository
	stockPublishBroker domain.BrokerPublisher
	cfg                *config.Config
}

func NewWarehouseUsecase(warehouseRepo domain.WarehouseRepository, stockRepo domain.StockRepository, reservedStockRepo domain.ReservedStockRepository, stockPublishBroker domain.BrokerPublisher, cfg *config.Config) domain.WarehouseService {
	return &warehouseUsecase{warehouseRepo, stockRepo, reservedStockRepo, stockPublishBroker, cfg}
}

func (u *warehouseUsecase) Create(ctx context.Context, shopID int64, req *domain.WarehouseCreateRequest) (*domain.Warehouse, error) {
	warehouse := &domain.Warehouse{
		ShopID:   shopID,
		Name:     req.Name,
		Location: req.Location,
		Active:   true,
	}

	err := u.warehouseRepo.Create(ctx, warehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseUsecase] Create", "createWarehouse", err)
		return nil, err
	}

	return warehouse, nil
}

func (u *warehouseUsecase) GetByShopID(ctx context.Context, shopID int64) ([]domain.Warehouse, error) {
	warehouses, err := u.warehouseRepo.GetByShopID(ctx, shopID)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseUsecase] GetByShopID", "getWarehouses", err)
		return nil, err
	}

	if len(warehouses) == 0 {
		return nil, domain.ErrNotFound
	}

	return warehouses, nil
}

func (u *warehouseUsecase) GetListWarehouse(ctx context.Context, shopID int64, param domain.GetListWarehouseRequest) ([]domain.Warehouse, domain.Metadata, error) {
	var metadata domain.Metadata
	warehouses, err := u.warehouseRepo.GetListWarehouse(ctx, shopID, param)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseUsecase] GetListWarehouse", "getWarehouses", err)
		return nil, metadata, err
	}

	if len(warehouses) == 0 {
		return nil, metadata, domain.ErrNotFound
	}

	count, err := u.warehouseRepo.GetListWarehouseCount(ctx, shopID, param)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseUsecase] GetListWarehouse", "getWarehousesCount", err)
		return nil, metadata, err
	}
	metadata.Page = param.Page
	metadata.Limit = param.Limit
	metadata.TotalData = count
	metadata.TotalPage = count / param.Limit
	if count%param.Limit != 0 {
		metadata.TotalPage++
	}

	return warehouses, metadata, nil
}

func (u *warehouseUsecase) UpdateStatus(ctx context.Context, id, shopID int64, req domain.WarehouseUpdateStatusRequest) error {
	warehouse, err := u.warehouseRepo.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "getWarehouse", err)
		return err
	}

	if warehouse.ShopID != shopID {
		slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "unauthorized", err)
		return domain.ErrUnauthorized
	}

	if warehouse.Active == req.Active {
		slog.InfoContext(ctx, "[warehouseUsecase] UpdateStatus", "noChange", nil)
		return nil
	}

	if err := u.warehouseRepo.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {

		stocks, err := u.stockRepo.GetByWarehouseID(ctx, id)
		if err != nil {
			slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "getStocks", err)
			return err
		}

		var stockIDs []int64
		var productIDs []int64
		for _, stock := range stocks {
			if stock.Quantity == 0 {
				continue
			}

			stockIDs = append(stockIDs, stock.ID)
			productIDs = append(productIDs, stock.ProductID)

		}

		availableStocks, err := u.stockRepo.GetAvailableStockByProductIDs(ctx, productIDs)
		if err != nil {
			slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "getAvailableStocks", err)
			return err
		}

		if !req.Active {
			reservedStocks, err := u.reservedStockRepo.GetTotalReservedStockByStockIDsAndStatus(ctx, stockIDs, domain.ReservedStockStatusActive)
			if err != nil {
				slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "getTotalReservedStockByStockIDsAndStatus", err)
				return err
			}

			for _, stock := range stocks {
				if stock.Quantity == 0 {
					continue
				}

				if reservedStocks[stock.ID] != 0 {
					slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "stockReserved", "still have reserved stock")
					return fmt.Errorf("%w: stock still have reserved stock", domain.ErrInvalidRequest)
				}
			}
		}

		for _, stock := range stocks {
			if stock.Quantity == 0 {
				continue
			}

			var availableStock int64
			if req.Active {
				availableStock = availableStocks[stock.ProductID] + stock.Quantity
			} else {
				availableStock = availableStocks[stock.ProductID] - stock.Quantity
			}

			_ = u.stockPublishBroker.PublishStockAvailable(ctx, domain.StockMessage{
				ProductID: stock.ProductID,
				Available: availableStock,
			})
		}

		err = u.warehouseRepo.UpdateStatus(ctx, id, req.Active, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "updateStatus", err)
			return err
		}

		return nil

	}); err != nil {
		slog.ErrorContext(ctx, "[warehouseUsecase] UpdateStatus", "transactionError", err)
		return err
	}

	return nil
}
