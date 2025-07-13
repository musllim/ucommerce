// internal/service/product_service.go
package service

import (
	"context"
	"errors"

	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Price <= 0 {
		return errors.New("product price must be positive")
	}
	if product.Stock < 0 {
		return errors.New("product stock cannot be negative")
	}

	return s.repo.Create(ctx, product)
}

func (s *ProductService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) CountProduct(ctx context.Context) (*int, error) {
	return s.repo.Count(ctx)
}

func (s *ProductService) ListProducts(ctx context.Context, limit, offset int, category string) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset, category)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Price <= 0 {
		return errors.New("product price must be positive")
	}
	if product.Stock < 0 {
		return errors.New("product stock cannot be negative")
	}

	return s.repo.Update(ctx, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
