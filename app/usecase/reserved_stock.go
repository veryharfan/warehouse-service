package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"warehouse-service/app/domain"
	"warehouse-service/config"
)

type reservedStockUsecase struct {
	stockRepo          domain.StockRepository
	reservedStockRepo  domain.ReservedStockRepository
	stockPublishBroker domain.BrokerPublisher
	cfg                *config.Config
}

func NewReservedStockUsecase(
	stockRepo domain.StockRepository,
	reservedStockRepo domain.ReservedStockRepository,
	stockPublishBroker domain.BrokerPublisher,
	cfg *config.Config) domain.ReservedStockUsecase {
	return &reservedStockUsecase{stockRepo, reservedStockRepo, stockPublishBroker, cfg}
}

func (u *reservedStockUsecase) CreateReservedStock(ctx context.Context, req domain.ReservedStockCreateRequest) error {

	stocks, err := u.stockRepo.GetByProductID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	var stockIDs []int64
	for _, s := range stocks {
		stockIDs = append(stockIDs, s.ID)
	}

	reservedStocks, err := u.reservedStockRepo.GetTotalReservedStockByStockIDsAndStatus(ctx, stockIDs, domain.ReservedStockStatusActive)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "getReservedStock", err)
		return err
	}

	var createReservedStock *domain.ReservedStock
	for _, stock := range stocks {

		reservedStock, ok := reservedStocks[stock.ID]
		if ok {
			available := stock.Quantity - reservedStock
			if available < req.Quantity {
				continue
			}
		}

		if stock.Quantity < req.Quantity {
			continue
		}

		createReservedStock = &domain.ReservedStock{
			StockID:  stock.ID,
			Quantity: req.Quantity,
			Status:   domain.ReservedStockStatusActive,
			OrderID:  req.OrderID,
		}

		break
	}

	if createReservedStock == nil {
		slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "insufficientStock", "no available stock")
		return fmt.Errorf("%w: insufficient stock available", domain.ErrValidation)
	}

	if err = u.stockRepo.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Lock the stock row for update
		stock, err := u.stockRepo.LockForUpdate(ctx, createReservedStock.StockID, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "lockStock", err)
			return err
		}

		// Get total stock available
		availableStock, err := u.stockRepo.GetAvailableStockByProductID(ctx, stock.ProductID)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "getAvailableStock", err)
			return err
		}

		err = u.reservedStockRepo.CreateReservedStock(ctx, createReservedStock, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "createReservedStock", err)
			return err
		}

		err = u.stockPublishBroker.PublishStockAvailable(ctx, domain.StockMessage{
			ProductID: stock.ProductID,
			Available: availableStock - createReservedStock.Quantity,
		})
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "publishStockAvailable", err)
			return err
		}

		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "[reservedStockUsecase] CreateReservedStock", "withTransaction", err)
		return err
	}

	return nil
}

func (u *reservedStockUsecase) UpdateReservedStockStatusByOrderID(ctx context.Context, orderID int64, req domain.ReservedStockUpdateRequest) error {
	reservedStock, err := u.reservedStockRepo.GetReservedStockByOrderID(ctx, orderID)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "getReservedStockByOrderID", err)
		return err
	}

	if reservedStock.Status == req.Status {
		slog.InfoContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "noChange", nil)
		return nil
	}

	if err = u.stockRepo.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Lock the stock row for update
		stock, err := u.stockRepo.LockForUpdate(ctx, reservedStock.StockID, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "lockStock", err)
			return err
		}

		// Get total stock available
		availableStock, err := u.stockRepo.GetAvailableStockByProductID(ctx, stock.ProductID)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "getAvailableStock", err)
			return err
		}

		if req.Status == domain.ReservedStockStatusCancelled {
			availableStock += reservedStock.Quantity
		} else if req.Status == domain.ReservedStockStatusCompleted {
			stock.Quantity -= reservedStock.Quantity
		}

		err = u.reservedStockRepo.UpdateReservedStockStatus(ctx, reservedStock.ID, req.Status)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "updateReservedStockStatus", err)
			return err
		}

		err = u.stockRepo.UpdateQuantity(ctx, stock.ID, stock.Quantity, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "updateStockQuantity", err)
			return err
		}

		err = u.stockPublishBroker.PublishStockAvailable(ctx, domain.StockMessage{
			ProductID: stock.ProductID,
			Available: availableStock,
		})
		if err != nil {
			slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "publishStockAvailable", err)
			return err
		}

		return nil
	}); err != nil {

		slog.ErrorContext(ctx, "[reservedStockUsecase] UpdateReservedStockStatusByOrderID", "withTransaction", err)
		return err
	}

	return nil
}
