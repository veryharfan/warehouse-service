package usecase

import (
	"context"
	"database/sql"
	"log/slog"
	"warehouse-service/app/domain"
)

type stockTransferUsecase struct {
	stockTransferRepo  domain.StockTransferRepository
	warehouseRepo      domain.WarehouseRepository
	stockRepo          domain.StockRepository
	stockPublishBroker domain.BrokerPublisher
}

func NewStockTransferUsecase(stockTransferRepo domain.StockTransferRepository,
	warehouseRepo domain.WarehouseRepository,
	stockRepo domain.StockRepository,
	stockPublishBroker domain.BrokerPublisher) domain.StockTransferUsecase {
	return &stockTransferUsecase{stockTransferRepo, warehouseRepo, stockRepo, stockPublishBroker}
}

func (u *stockTransferUsecase) CreateTransfer(ctx context.Context, shopID int64, req domain.StockTransferCreateRequest) (*domain.StockTransfer, error) {
	fromWarehouse, err := u.warehouseRepo.GetByID(ctx, req.FromWarehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] CreateTransfer", "getFromWarehouse", err)
		return nil, err
	}

	toWarehouse, err := u.warehouseRepo.GetByID(ctx, req.ToWarehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] CreateTransfer", "getToWarehouse", err)
		return nil, err
	}

	if fromWarehouse.ShopID != toWarehouse.ShopID || fromWarehouse.ShopID != shopID {
		slog.ErrorContext(ctx, "[stockTransferUsecase] CreateTransfer", "invalidShopID", "shopID")
		return nil, domain.ErrInvalidRequest
	}

	fromWarehouseStock, err := u.stockRepo.GetByProductIDAndWarehouseID(ctx, req.ProductID, req.FromWarehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] CreateTransfer", "getFromWarehouseStock", err)
		return nil, err
	}

	if (fromWarehouseStock.Quantity - fromWarehouseStock.Reserved) < req.Quantity {
		slog.ErrorContext(ctx, "[stockTransferUsecase] CreateTransfer", "insufficientStock", "fromWarehouseStock")
		return nil, domain.ErrInvalidRequest
	}

	// Create the stock transfer
	stockTransfer := &domain.StockTransfer{
		ProductID:     req.ProductID,
		FromWarehouse: req.FromWarehouse,
		ToWarehouse:   req.ToWarehouse,
		Quantity:      req.Quantity,
		Status:        domain.TransferStatusNotStarted,
		Description:   req.Description,
	}
	// Create the stock transfer in the database
	err = u.stockTransferRepo.Create(ctx, stockTransfer)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] CreateTransfer", "createTransfer", err)
		return nil, err
	}

	slog.InfoContext(ctx, "[stockTransferUsecase] CreateTransfer", "transfer", stockTransfer)
	return stockTransfer, nil
}

func (u *stockTransferUsecase) GetTransferByID(ctx context.Context, id int64, shopID *int64) (domain.StockTransfer, error) {
	var st domain.StockTransfer
	var err error

	st, err = u.stockTransferRepo.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] GetTransferByID", "getTransfer", err)
		return st, err
	}

	warehouse, err := u.warehouseRepo.GetByID(ctx, st.FromWarehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] GetTransferByID", "getFromWarehouse", err)
		return st, err
	}

	if shopID != nil && *shopID != warehouse.ShopID {
		slog.ErrorContext(ctx, "[stockTransferUsecase] GetTransferByID", "invalidShopID", err)
		return st, domain.ErrInvalidRequest
	}

	return st, nil
}

func (u *stockTransferUsecase) UpdateTransferStatus(ctx context.Context, id int64, req domain.StockTransferUpdateRequest) error {
	st, err := u.stockTransferRepo.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "getTransfer", err)
		return err
	}

	fromWarehouseStock, err := u.stockRepo.GetByProductIDAndWarehouseID(ctx, st.ProductID, st.FromWarehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "getFromWarehouseStock", err)
		return err
	}
	toWarehouseStock, err := u.stockRepo.GetByProductIDAndWarehouseID(ctx, st.ProductID, st.ToWarehouse)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "getToWarehouseStock", err)
		return err
	}

	availableStock, err := u.stockRepo.GetAvailableStockByProductID(ctx, st.ProductID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "getAvailableStock", err)
		return err
	}

	var updateFromWarehouseStock, updateToWarehouseStock bool

	switch req.Status {
	case domain.TransferStatusInProgress:
		if st.Status != domain.TransferStatusNotStarted {
			return domain.ErrInvalidRequest
		}

		if fromWarehouseStock.Quantity-fromWarehouseStock.Reserved < st.Quantity {
			return domain.ErrInvalidRequest
		}

		fromWarehouseStock.Quantity -= st.Quantity
		availableStock -= st.Quantity
		updateFromWarehouseStock = true

	case domain.TransferStatusCompleted:
		if st.Status != domain.TransferStatusInProgress {
			return domain.ErrInvalidRequest
		}

		toWarehouseStock.Quantity += st.Quantity
		availableStock += st.Quantity
		updateToWarehouseStock = true

	case domain.TransferStatusReverted:
		if st.Status != domain.TransferStatusInProgress {
			return domain.ErrInvalidRequest
		}

		fromWarehouseStock.Quantity += st.Quantity
		availableStock += st.Quantity
		updateFromWarehouseStock = true

	case domain.TransferStatusFailed:
		if st.Status != domain.TransferStatusInProgress {
			return domain.ErrInvalidRequest
		}
	default:
		return domain.ErrInvalidRequest
	}

	st.Status = req.Status
	st.Description = req.Description

	if err = u.stockTransferRepo.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err = u.stockTransferRepo.UpdateStatus(ctx, st, tx)
		if err != nil {
			slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "updateStatus", err)
			return err
		}

		if updateFromWarehouseStock {
			err = u.stockRepo.UpdateQuantity(ctx, fromWarehouseStock.ID, fromWarehouseStock.Quantity, fromWarehouseStock.Version, tx)
			if err != nil {
				slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "updateFromWarehouseStock", err)
				return err
			}
		}

		if updateToWarehouseStock {
			err = u.stockRepo.UpdateQuantity(ctx, toWarehouseStock.ID, toWarehouseStock.Quantity, toWarehouseStock.Version, tx)
			if err != nil {
				slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "updateToWarehouseStock", err)
				return err
			}
		}

		if err = u.stockPublishBroker.PublishStockAvailable(ctx, domain.StockMessage{
			ProductID: fromWarehouseStock.ProductID,
			Available: availableStock,
		}); err != nil {
			slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "publishStockAvailable", err)
			return err
		}

		return nil

	}); err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] UpdateTransferStatus", "transactionError", err)
		return err
	}

	return nil
}

func (u *stockTransferUsecase) GetListStockTransfer(ctx context.Context, shopID int64, param domain.GetListStockTransferRequest) ([]domain.StockTransfer, domain.Metadata, error) {
	var stockTransfers []domain.StockTransfer
	var err error

	stockTransfers, err = u.stockTransferRepo.GetListStockTransfer(ctx, shopID, param)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] GetListStockTransfer", "getListStockTransfer", err)
		return nil, domain.Metadata{}, err
	}

	count, err := u.stockTransferRepo.GetListStockTransferCount(ctx, shopID, param)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferUsecase] GetListStockTransfer", "getListStockTransferCount", err)
		return nil, domain.Metadata{}, err
	}

	metadata := domain.Metadata{
		TotalData: count,
		TotalPage: (count + param.Limit - 1) / param.Limit,
		Page:      param.Page,
		Limit:     param.Limit,
		SortBy:    param.SortBy,
		SortOrder: param.SortOrder,
	}

	return stockTransfers, metadata, nil
}
