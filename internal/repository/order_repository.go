// internal/repository/order_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/musllim/ecommerce/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order with its items
func (r *OrderRepository) Create(ctx context.Context, order *models.Order, orderItems []*models.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (user_id, status, total, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id
	`

	now := time.Now()
	err = tx.QueryRowContext(ctx, orderQuery,
		order.UserID,
		order.Status,
		order.Total,
		now,
		now,
	).Scan(&order.ID)

	if err != nil {
		return err
	}

	order.CreatedAt = &now
	order.UpdatedAt = &now

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	for _, item := range orderItems {
		_, err = tx.ExecContext(ctx, itemQuery,
			order.ID,
			item.ProductID,
			item.Quantity,
			item.Price,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves an order by its ID
func (r *OrderRepository) GetByID(ctx context.Context, orderID int64) (*models.Order, error) {
	query := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE id = ?
	`

	order := &models.Order{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Total,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	return order, nil
}

// GetByUserID retrieves all orders for a user
func (r *OrderRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Total,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetOrderItems retrieves all items for an order
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]*models.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price, created_at
		FROM order_items
		WHERE order_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderItems []*models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orderItems, nil
}

// UpdateStatus updates the status of an order
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	query := `
		UPDATE orders
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), orderID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
} 