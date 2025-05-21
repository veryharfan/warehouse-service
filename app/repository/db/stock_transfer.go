package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"warehouse-service/app/domain"
)

type stockTransferRepository struct {
	conn *sql.DB
}

func NewStockTransferRepository(db *sql.DB) domain.StockTransferRepository {
	return &stockTransferRepository{db}
}

func (r *stockTransferRepository) Create(ctx context.Context, data *domain.StockTransfer) error {
	query := `INSERT INTO stock_transfers (product_id, from_warehouse, to_warehouse, quantity, description)
	VALUES ($1, $2, $3, $4, $5) Returning id, created_at, updated_at`
	err := r.conn.QueryRowContext(ctx, query, data.ProductID, data.FromWarehouse, data.ToWarehouse, data.Quantity, data.Description).
		Scan(&data.ID, &data.CreatedAt, &data.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] Create", "queryRowContext", err)
		return err
	}

	return nil
}

func (r *stockTransferRepository) GetByID(ctx context.Context, id int64) (domain.StockTransfer, error) {
	query := `SELECT id, product_id, from_warehouse, to_warehouse, quantity, status, description, created_at, updated_at
	FROM stock_transfers WHERE id = $1`

	var stockTransfer domain.StockTransfer
	err := r.conn.QueryRowContext(ctx, query, id).Scan(&stockTransfer.ID, &stockTransfer.ProductID,
		&stockTransfer.FromWarehouse, &stockTransfer.ToWarehouse, &stockTransfer.Quantity,
		&stockTransfer.Status, &stockTransfer.Description, &stockTransfer.CreatedAt, &stockTransfer.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] GetByID", "queryRowContext", err)
		if err == sql.ErrNoRows {
			return stockTransfer, domain.ErrNotFound
		}
		return stockTransfer, err
	}

	return stockTransfer, nil
}

func (r *stockTransferRepository) UpdateStatus(ctx context.Context, st domain.StockTransfer, tx *sql.Tx) error {
	query := `UPDATE stock_transfers SET status = $1, description = $2, updated_at = now() WHERE id = $3`
	_, err := tx.ExecContext(ctx, query, st.Status, st.Description, st.ID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] UpdateStatus", "execContext", err)
		return err
	}
	return nil
}

func (r *stockTransferRepository) WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] WithTransaction", "beginTx", err)
		return err
	}

	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			slog.ErrorContext(ctx, "[stockTransferRepository] WithTransaction", "rollback", rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] WithTransaction", "commit", err)
		return err
	}

	return nil
}

func (r *stockTransferRepository) GetListStockTransfer(ctx context.Context, shopID int64, param domain.GetListStockTransferRequest) ([]domain.StockTransfer, error) {
	query := `SELECT id, product_id, from_warehouse, to_warehouse, quantity, status, description, created_at, updated_at
	FROM stock_transfers WHERE from_warehouse IN (SELECT id FROM warehouses WHERE shop_id = $1)`
	args := []interface{}{shopID}
	placeholder := 2

	if param.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", placeholder)
		args = append(args, param.Status)
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
		slog.ErrorContext(ctx, "[stockTransferRepository] GetListStockTransfer", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var stockTransfers []domain.StockTransfer
	for rows.Next() {
		var stockTransfer domain.StockTransfer
		if err := rows.Scan(&stockTransfer.ID, &stockTransfer.ProductID,
			&stockTransfer.FromWarehouse, &stockTransfer.ToWarehouse,
			&stockTransfer.Quantity, &stockTransfer.Status,
			&stockTransfer.Description, &stockTransfer.CreatedAt,
			&stockTransfer.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[stockTransferRepository] GetListStockTransfer", "scan", err)
			return nil, err
		}
		stockTransfers = append(stockTransfers, stockTransfer)
	}
	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] GetListStockTransfer", "rowError", err)
		return nil, err
	}
	return stockTransfers, nil
}

func (r *stockTransferRepository) GetListStockTransferCount(ctx context.Context, shopID int64, param domain.GetListStockTransferRequest) (int64, error) {
	query := `SELECT COUNT(*) FROM stock_transfers WHERE from_warehouse IN (SELECT id FROM warehouses WHERE shop_id = $1)`
	args := []interface{}{shopID}
	placeholder := 2

	if param.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", placeholder)
		args = append(args, param.Status)
		placeholder++
	}

	var count int64
	err := r.conn.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] GetListStockTransferCount", "queryRowContext", err)
		return 0, err
	}
	return count, nil
}
