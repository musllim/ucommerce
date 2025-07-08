// internal/repository/product_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/musllim/ecommerce/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (name, description, price, stock, category, image_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.ImageURL,
		now,
		now,
	).Scan(&product.ID)

	if err != nil {
		return err
	}

	product.CreatedAt = &now
	product.UpdatedAt = &now
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
		FROM products
		WHERE id = ?
	`

	product := &models.Product{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.Category,
		&product.ImageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int, category string) ([]*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
		FROM products
		WHERE (? = '' OR category = ?)
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, category, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Category,
			&product.ImageURL,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET name = ?, description = ?, price = ?, stock = ?, category = ?, image_url = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.ImageURL,
		now,
		product.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	product.UpdatedAt = &now
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}
