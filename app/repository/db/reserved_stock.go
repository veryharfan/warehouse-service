package db

import (
	"context"
	"database/sql"
	"log/slog"
	"warehouse-service/app/domain"
)

type reservedStockRepository struct {
	conn *sql.DB
}

func NewReservedStockRepository(db *sql.DB) domain.ReservedStockRepository {
	return &reservedStockRepository{db}
}

func (r *reservedStockRepository) CreateReservedStock(ctx context.Context, stockID, quantity, orderID int64) (int64, error) {
	query := `INSERT INTO reserved_stocks (stock_id, quantity, order_id) VALUES ($1, $2, $3) RETURNING id`
	var id int64
	err := r.conn.QueryRowContext(ctx, query, stockID, quantity, orderID).Scan(&id)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockRepository] CreateReservedStock", "queryRowContext", err)
		return 0, err
	}
	return id, nil
}

func (r *reservedStockRepository) GetReservedStocksByStockIDAndStatus(ctx context.Context, stockID int64, status domain.ReservedStockStatus) ([]domain.ReservedStock, error) {
	query := `SELECT id, stock_id, quantity, status, created_at, updated_at FROM reserved_stocks WHERE stock_id = $1 `
	args := []interface{}{stockID}

	if status != "" {
		query += `AND status = $2`
		args = append(args, status)
	}

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockRepository] GetReservedStocksByStockIDAndStatus", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var reservedStocks []domain.ReservedStock
	for rows.Next() {
		var reservedStock domain.ReservedStock
		if err := rows.Scan(&reservedStock.ID, &reservedStock.StockID, &reservedStock.Quantity, &reservedStock.Status, &reservedStock.CreatedAt, &reservedStock.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[reservedStockRepository] GetReservedStocksByStockIDAndStatus", "scan", err)
			return nil, err
		}
		reservedStocks = append(reservedStocks, reservedStock)
	}

	return reservedStocks, nil
}
func (r *reservedStockRepository) UpdateReservedStockStatus(ctx context.Context, id int64, status domain.ReservedStockStatus) error {
	query := `UPDATE reserved_stocks SET status = $1 WHERE id = $2`
	_, err := r.conn.ExecContext(ctx, query, status, id)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockRepository] UpdateReservedStockStatus", "execContext", err)
		return err
	}
	return nil
}

func (r *reservedStockRepository) GetReservedStockByOrderID(ctx context.Context, orderID int64) (domain.ReservedStock, error) {
	query := `SELECT id, stock_id, quantity, order_id, status, created_at, updated_at FROM reserved_stocks WHERE order_id = $1`
	var reservedStock domain.ReservedStock
	err := r.conn.QueryRowContext(ctx, query, orderID).Scan(&reservedStock.ID, &reservedStock.StockID, &reservedStock.Quantity, &reservedStock.OrderID, &reservedStock.Status, &reservedStock.CreatedAt, &reservedStock.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockRepository] GetReservedStockByOrderID", "queryRowContext", err)
		return domain.ReservedStock{}, err
	}
	return reservedStock, nil
}

func (r *reservedStockRepository) GetTotalReservedStockByStockIDAndStatus(ctx context.Context, stockID int64, status domain.ReservedStockStatus) (int64, error) {
	query := `SELECT COALESCE(SUM(quantity), 0) FROM reserved_stocks WHERE stock_id = $1 AND status = $2`
	var total int64
	err := r.conn.QueryRowContext(ctx, query, stockID, status).Scan(&total)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockRepository] GetTotalReservedStockByStockIDAndStatus", "queryRowContext", err)
		return 0, err
	}
	return total, nil
}

func (r *reservedStockRepository) GetTotalReservedStockByStockIDsAndStatus(ctx context.Context, stockIDs []int64, status domain.ReservedStockStatus) (map[int64]int64, error) {
	query := `SELECT stock_id, COALESCE(SUM(quantity), 0) FROM reserved_stocks WHERE stock_id = ANY($1) AND status = $2 GROUP BY stock_id`
	args := []interface{}{stockIDs, status}

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "[reservedStockRepository] GetTotalReservedStockByStockIDsAndStatus", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]int64)
	for rows.Next() {
		var stockID int64
		var total int64
		if err := rows.Scan(&stockID, &total); err != nil {
			slog.ErrorContext(ctx, "[reservedStockRepository] GetTotalReservedStockByStockIDsAndStatus", "scan", err)
			return nil, err
		}
		result[stockID] = total
	}

	return result, nil
}
