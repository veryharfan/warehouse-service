package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"warehouse-service/app/domain"
)

type stockRepository struct {
	conn *sql.DB
}

func NewStockRepository(db *sql.DB) domain.StockRepository {
	return &stockRepository{db}
}

func (r *stockRepository) Create(ctx context.Context, stocks []domain.Stock) error {

	valuePlaceholders := []string{}
	valueArgs := []interface{}{}
	for i, stock := range stocks {
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, stock.ProductID, stock.WarehouseID)
	}

	query := fmt.Sprintf(`INSERT INTO stocks (product_id, warehouse_id) VALUES %s`, strings.Join(valuePlaceholders, ", "))

	res, err := r.conn.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] Create", "execContext", err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] Create", "rowsAffected", err)
		return err
	}

	if rowsAffected == 0 {
		slog.ErrorContext(ctx, "[stockRepository] Create", "noRowsAffected", "No rows were inserted")
		return fmt.Errorf("no rows were inserted")
	}

	slog.InfoContext(ctx, "[stockRepository] Create", "rowsAffected", rowsAffected)
	return nil
}

func (r *stockRepository) GetByProductID(ctx context.Context, productID int64) ([]domain.Stock, error) {
	query := `SELECT s.id, s.product_id, s.warehouse_id, s.quantity, s.created_at, s.updated_at 
	FROM stocks s
	WHERE s.product_id = $1`

	rows, err := r.conn.QueryContext(ctx, query, productID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByProductID", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var stocks []domain.Stock
	for rows.Next() {
		var stock domain.Stock
		if err := rows.Scan(&stock.ID, &stock.ProductID, &stock.WarehouseID, &stock.Quantity,
			&stock.CreatedAt, &stock.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[stockRepository] GetByProductID", "scan", err)
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByProductID", "rowError", err)
		return nil, err
	}

	return stocks, nil
}

func (r *stockRepository) GetByID(ctx context.Context, id int64) (domain.Stock, error) {
	query := `SELECT id, product_id, warehouse_id, quantity, created_at, updated_at 
	FROM stocks WHERE id = $1`

	var stock domain.Stock
	err := r.conn.QueryRowContext(ctx, query, id).Scan(&stock.ID, &stock.ProductID,
		&stock.WarehouseID, &stock.Quantity, &stock.CreatedAt, &stock.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByID", "queryRowContext", err)
		if err == sql.ErrNoRows {
			return stock, domain.ErrNotFound
		}
		return stock, err
	}

	return stock, nil
}

func (r *stockRepository) UpdateQuantity(ctx context.Context, id, quantity int64, tx *sql.Tx) error {
	query := `UPDATE stocks SET quantity = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, quantity, id)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] UpdateQuantity", "execContext", err)
		return err
	}

	return nil
}

func (r *stockRepository) GetAvailableStockByProductID(ctx context.Context, productID int64) (int64, error) {
	query := `SELECT COALESCE(SUM(s.quantity), 0) - COALESCE(SUM(rs.quantity), 0) AS available_stock
	FROM stocks s
	JOIN warehouses w ON s.warehouse_id = w.id
	LEFT JOIN reserved_stocks rs ON rs.stock_id = s.id AND rs.status = 'active'
	WHERE s.product_id = $1
  	AND w.active = TRUE`

	var availableStock int64
	err := r.conn.QueryRowContext(ctx, query, productID).Scan(&availableStock)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetAvailableStockByProductID", "queryRowContext", err)
		return 0, err
	}

	return availableStock, nil
}

func (r *stockRepository) WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] WithTransaction", "beginTx", err)
		return err
	}

	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			slog.ErrorContext(ctx, "[stockRepository] WithTransaction", "rollback", rollbackErr)
			return err
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "[stockTransferRepository] WithTransaction", "commit", err)
		return err
	}

	return nil
}

func (r *stockRepository) GetByProductIDAndWarehouseID(ctx context.Context, productID, warehouseID int64) (domain.Stock, error) {
	query := `SELECT id, product_id, warehouse_id, quantity, created_at, updated_at 
	FROM stocks WHERE product_id = $1 AND warehouse_id = $2`

	var stock domain.Stock
	err := r.conn.QueryRowContext(ctx, query, productID, warehouseID).Scan(&stock.ID, &stock.ProductID,
		&stock.WarehouseID, &stock.Quantity, &stock.CreatedAt, &stock.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByProductIDAndWarehouseID", "queryRowContext", err)
		if err == sql.ErrNoRows {
			return stock, domain.ErrNotFound
		}
		return stock, err
	}

	return stock, nil
}

func (r *stockRepository) GetListStock(ctx context.Context, shopID int64, param domain.GetListStockRequest) ([]domain.Stock, error) {
	query := `SELECT s.id, s.product_id, s.warehouse_id, s.quantity, s.created_at, s.updated_at
	FROM stocks s
	JOIN warehouses w ON s.warehouse_id = w.id 
	WHERE w.shop_id = $1 AND w.active = true`

	args := []any{shopID}
	placeholder := 2

	if param.ProductID != 0 {
		query += fmt.Sprintf(" AND s.product_id = $%d", placeholder)
		args = append(args, param.ProductID)
		placeholder++
	}
	if param.WarehouseID != 0 {
		query += fmt.Sprintf(" AND s.warehouse_id = $%d", placeholder)
		args = append(args, param.WarehouseID)
		placeholder++
	}

	if param.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", param.SortBy)
		if param.SortOrder != "" {
			query += fmt.Sprintf(" %s", param.SortOrder)
		}
	}

	if param.Page > 0 && param.Limit > 0 {
		offset := (param.Page - 1) * param.Limit
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", param.Limit, offset)
	}

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetListStock", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var stocks []domain.Stock
	for rows.Next() {
		var stock domain.Stock
		if err := rows.Scan(&stock.ID, &stock.ProductID, &stock.WarehouseID,
			&stock.Quantity, &stock.CreatedAt, &stock.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[stockRepository] GetListStock", "scan", err)
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetListStock", "rowError", err)
		return nil, err
	}

	return stocks, nil
}

func (r *stockRepository) GetListStockCount(ctx context.Context, shopID int64, param domain.GetListStockRequest) (int64, error) {
	query := `SELECT COUNT(*) 
	FROM stocks s
	JOIN warehouses w ON s.warehouse_id = w.id 
	WHERE w.shop_id = $1 AND w.active = true`

	args := []any{shopID}
	placeholder := 2

	if param.ProductID != 0 {
		query += fmt.Sprintf(" AND s.product_id = $%d", placeholder)
		args = append(args, param.ProductID)
		placeholder++
	}
	if param.WarehouseID != 0 {
		query += fmt.Sprintf(" AND s.warehouse_id = $%d", placeholder)
		args = append(args, param.WarehouseID)
	}

	var count int64
	err := r.conn.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetListStockCount", "queryRowContext", err)
		return 0, err
	}

	return count, nil
}

func (r *stockRepository) GetByWarehouseID(ctx context.Context, warehouseID int64) ([]domain.Stock, error) {
	query := `SELECT id, product_id, warehouse_id, quantity, created_at, updated_at 
	FROM stocks WHERE warehouse_id = $1`

	rows, err := r.conn.QueryContext(ctx, query, warehouseID)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByWarehouseID", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	var stocks []domain.Stock
	for rows.Next() {
		var stock domain.Stock
		if err := rows.Scan(&stock.ID, &stock.ProductID,
			&stock.WarehouseID, &stock.Quantity,
			&stock.CreatedAt, &stock.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[stockRepository] GetByWarehouseID", "scan", err)
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByWarehouseID", "rowError", err)
		return nil, err
	}

	return stocks, nil
}

func (r *stockRepository) LockForUpdate(ctx context.Context, id int64, tx *sql.Tx) (domain.Stock, error) {
	query := `SELECT id, product_id, warehouse_id, quantity, created_at, updated_at 
	FROM stocks WHERE id = $1 FOR UPDATE`

	var stock domain.Stock
	err := tx.QueryRowContext(ctx, query, id).Scan(&stock.ID, &stock.ProductID,
		&stock.WarehouseID, &stock.Quantity, &stock.CreatedAt, &stock.UpdatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] LockForUpdate", "queryRowContext", err)
		if err == sql.ErrNoRows {
			return stock, domain.ErrNotFound
		}
		return stock, err
	}

	return stock, nil
}

func (r *stockRepository) GetAvailableStockByProductIDs(ctx context.Context, productIDs []int64) (map[int64]int64, error) {
	query := `SELECT s.product_id, COALESCE(SUM(s.quantity), 0) - COALESCE(SUM(rs.quantity), 0) AS available_stock
	FROM stocks s
	JOIN warehouses w ON s.warehouse_id = w.id
	LEFT JOIN reserved_stocks rs ON rs.stock_id = s.id AND rs.status = 'active'
	WHERE s.product_id = ANY($1)
  	AND w.active = TRUE
	GROUP BY s.product_id`

	rows, err := r.conn.QueryContext(ctx, query, productIDs)
	if err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetAvailableStockByProductIDs", "queryContext", err)
		return nil, err
	}
	defer rows.Close()

	availableStocks := make(map[int64]int64)
	for rows.Next() {
		var productID int64
		var availableStock int64
		if err := rows.Scan(&productID, &availableStock); err != nil {
			slog.ErrorContext(ctx, "[stockRepository] GetAvailableStockByProductIDs", "scan", err)
			return nil, err
		}
		availableStocks[productID] = availableStock
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetAvailableStockByProductIDs", "rowError", err)
		return nil, err
	}

	return availableStocks, nil
}
