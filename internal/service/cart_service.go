// internal/service/cart_service.go
package service

import (
	"context"
	"errors"

	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/repository"
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) GetOrCreateCart(ctx context.Context, userID int64) (*models.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err == nil {
		return cart, nil
	}

	if err.Error() == "cart not found" {
		return s.cartRepo.CreateCart(ctx, userID)
	}

	return nil, err
}

func (s *CartService) AddToCart(ctx context.Context, userID, productID int64, quantity int) error {
	// Get or create cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	// Check product exists and has enough stock
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		return errors.New("not enough stock available")
	}

	// Add or update cart item
	_, err = s.cartRepo.AddCartItem(ctx, cart.ID, productID, quantity)
	return err
}

func (s *CartService) GetCartItems(ctx context.Context, userID int64) ([]*models.CartItem, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err.Error() == "cart not found" {
			return []*models.CartItem{}, nil
		}
		return nil, err
	}

	return s.cartRepo.GetCartItems(ctx, cart.ID)
}

func (s *CartService) RemoveFromCart(ctx context.Context, userID, productID int64) error {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return err
	}

	item, err := s.cartRepo.GetCartItemByProduct(ctx, cart.ID, productID)
	if err != nil {
		return err
	}

	return s.cartRepo.RemoveCartItem(ctx, item.ID)
}

func (s *CartService) UpdateCartItemQuantity(ctx context.Context, userID, productID int64, quantity int) error {
	if quantity <= 0 {
		return s.RemoveFromCart(ctx, userID, productID)
	}

	// Check product stock
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		return errors.New("not enough stock available")
	}

	// Get cart
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Update item
	item, err := s.cartRepo.GetCartItemByProduct(ctx, cart.ID, productID)
	if err != nil {
		return err
	}

	_, err = s.cartRepo.UpdateCartItemQuantity(ctx, item.ID, quantity)
	return err
}
