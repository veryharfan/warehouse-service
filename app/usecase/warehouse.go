package usecase

import (
	"context"
	"log/slog"
	"warehouse-service/app/domain"
	"warehouse-service/config"
)

type warehouseUsecase struct {
	warehouseRepo domain.WarehouseRepository
	cfg           *config.Config
}

func NewWarehouseUsecase(warehouseRepo domain.WarehouseRepository, cfg *config.Config) domain.WarehouseService {
	return &warehouseUsecase{warehouseRepo, cfg}
}

func (u *warehouseUsecase) Create(ctx context.Context, req *domain.WarehouseCreateRequest) (*domain.Warehouse, error) {
	warehouse := &domain.Warehouse{
		ShopID:   req.ShopID,
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
	return warehouses, nil
}
