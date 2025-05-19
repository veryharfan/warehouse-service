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
	query := `SELECT s.id, s.product_id, s.warehouse_id, s.quantity, s.reserved, s.created_at, s.updated_at 
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
			&stock.Reserved, &stock.CreatedAt, &stock.UpdatedAt); err != nil {
			slog.ErrorContext(ctx, "[stockRepository] GetByProductID", "scan", err)
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "[stockRepository] GetByProductIDAndShopID", "rowError", err)
		return nil, err
	}

	return stocks, nil
}
