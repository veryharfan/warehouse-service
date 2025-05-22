package db

import (
	"context"
	"database/sql"
	"fmt"
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
	query := `SELECT id, shop_id, name, location, active, created_at, updated_at 
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
		if err := rows.Scan(&warehouse.ID, &warehouse.ShopID, &warehouse.Name, &warehouse.Location, &warehouse.Active,
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

func (r *warehouseRepository) GetByID(ctx context.Context, id int64) (domain.Warehouse, error) {
	query := `SELECT id, shop_id, name, location, active, created_at, updated_at 
	FROM warehouses WHERE id = $1`

	var warehouse domain.Warehouse
	err := r.conn.QueryRowContext(ctx, query, id).Scan(&warehouse.ID, &warehouse.ShopID,
		&warehouse.Name, &warehouse.Location, &warehouse.Active, &warehouse.CreatedAt, &warehouse.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] GetByID", "queryRowContext", err)
		if err == sql.ErrNoRows {
			return warehouse, domain.ErrNotFound
		}
		return warehouse, err
	}

	return warehouse, nil
}

func (r *warehouseRepository) GetListWarehouse(ctx context.Context, shopID int64, param domain.GetListWarehouseRequest) ([]domain.Warehouse, error) {
	query := `SELECT id, shop_id, name, location, active, created_at, updated_at 
	FROM warehouses WHERE shop_id = $1`
	args := []interface{}{shopID}
	placeholder := 2

	if param.Active {
		query += ` AND active = $1`
		args = append(args, param.Active)
		placeholder++
	}

	if param.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", param.SortBy)
		if param.SortOrder != "" {
			query += fmt.Sprintf(" %s", param.SortOrder)
		}
	} else {
		query += ` ORDER BY created_at DESC`
	}

	offset := (param.Page - 1) * param.Limit
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", placeholder, placeholder+1)
	args = append(args, param.Limit, offset)

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] GetListWarehouse", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var warehouse domain.Warehouse
		if err := rows.Scan(&warehouse.ID, &warehouse.ShopID,
			&warehouse.Name, &warehouse.Location,
			&warehouse.Active, &warehouse.CreatedAt,
			&warehouse.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[warehouseRepository] GetListWarehouse", "scan", err)
			return nil, err
		}
		warehouses = append(warehouses, warehouse)
	}
	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] GetListWarehouse", "rowError", err)
		return nil, err
	}
	return warehouses, nil
}
func (r *warehouseRepository) GetListWarehouseCount(ctx context.Context, shopID int64, param domain.GetListWarehouseRequest) (int64, error) {
	query := `SELECT COUNT(*) FROM warehouses WHERE shop_id = $1`
	args := []interface{}{shopID}
	placeholder := 2

	if param.Active {
		query += ` AND active = $1`
		args = append(args, param.Active)
		placeholder++
	}

	var count int64
	err := r.conn.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] GetListWarehouseCount", "queryRowContext", err)
		return 0, err
	}
	return count, nil
}
func (r *warehouseRepository) UpdateStatus(ctx context.Context, id int64, active bool, tx *sql.Tx) error {
	query := `UPDATE warehouses SET active = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, active, id)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] UpdateStatus", "execContext", err)
		return err
	}
	return nil
}

func (r *warehouseRepository) WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] WithTransaction", "beginTx", err)
		return err
	}

	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			slog.ErrorContext(ctx, "[warehouseRepository] WithTransaction", "rollback", rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "[warehouseRepository] WithTransaction", "commit", err)
		return err
	}
	return nil
}
