package db

import (
	"context"
	"database/sql"
	"log/slog"
	"warehouse-service/app/domain"
)

type warehouseRepository struct {
	conn *sql.DB
}

func NewWarehouseRepository(db *sql.DB) domain.WarehouseRepository {
	return &warehouseRepository{db}
}

func (r *warehouseRepository) Create(ctx context.Context, warehouse *domain.Warehouse) error {
	query := `INSERT INTO warehouses (shop_id, name, location, active) 
	VALUES ($1, $2, $3, $4)
	Returning id, created_at, updated_at
	`

	err := r.conn.QueryRowContext(ctx, query, warehouse.ShopID, warehouse.Name, warehouse.Location, warehouse.Active).
		Scan(
			&warehouse.ID,
			&warehouse.CreatedAt,
			&warehouse.UpdatedAt,
		)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] Create", "queryRowContext", err)
		return err
	}
	return nil
}

func (r *warehouseRepository) GetByShopID(ctx context.Context, shopID int64) ([]domain.Warehouse, error) {
	query := `SELECT id, shop_id, name, active, created_at, updated_at 
	FROM warehouses WHERE shop_id = $1`
	rows, err := r.conn.QueryContext(ctx, query, shopID)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] GetByShopID", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var warehouse domain.Warehouse
		if err := rows.Scan(&warehouse.ID, &warehouse.ShopID, &warehouse.Name, &warehouse.Active,
			&warehouse.CreatedAt, &warehouse.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[warehouseRepository] GetByShopID", "scan", err)
			return nil, err
		}
		warehouses = append(warehouses, warehouse)
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] GetByShopID", "rowError", err)
		return nil, err
	}

	return warehouses, nil
}
