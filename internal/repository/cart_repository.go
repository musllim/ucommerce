// internal/repository/cart_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/musllim/ecommerce/internal/models"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// CreateCart creates a new cart for a user
func (r *CartRepository) CreateCart(ctx context.Context, userID int64) (*models.Cart, error) {
	query := `
		INSERT INTO carts (user_id, created_at, updated_at)
		VALUES (?, ?, ?)
		RETURNING id
	`

	now := time.Now()
	cart := &models.Cart{
		UserID:    userID,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	err := r.db.QueryRowContext(ctx, query, userID, now, now).Scan(&cart.ID)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// GetCartByUserID retrieves a cart for a specific user
func (r *CartRepository) GetCartByUserID(ctx context.Context, userID int64) (*models.Cart, error) {
	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE user_id = ?
	`

	cart := &models.Cart{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}

	return cart, nil
}

// GetCartByID retrieves a cart by its ID
func (r *CartRepository) GetCartByID(ctx context.Context, cartID int64) (*models.Cart, error) {
	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE id = ?
	`

	cart := &models.Cart{}
	err := r.db.QueryRowContext(ctx, query, cartID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}

	return cart, nil
}

// DeleteCart deletes a cart and all its items
func (r *CartRepository) DeleteCart(ctx context.Context, cartID int64) error {
	query := `DELETE FROM carts WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, cartID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("cart not found")
	}

	return nil
}

// AddCartItem adds a product to the cart
func (r *CartRepository) AddCartItem(ctx context.Context, cartID, productID int64, quantity int) (*models.CartItem, error) {
	// First, check if the item already exists in the cart
	existingItem, err := r.GetCartItemByProduct(ctx, cartID, productID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	now := time.Now()

	if existingItem != nil {
		// Update existing item quantity
		newQuantity := existingItem.Quantity + quantity
		return r.UpdateCartItemQuantity(ctx, existingItem.ID, newQuantity)
	}

	// Insert new cart item
	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id
	`

	cartItem := &models.CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	err = r.db.QueryRowContext(ctx, query, cartID, productID, quantity, now, now).Scan(&cartItem.ID)
	if err != nil {
		return nil, err
	}

	return cartItem, nil
}

// GetCartItemByProduct retrieves a cart item by cart ID and product ID
func (r *CartRepository) GetCartItemByProduct(ctx context.Context, cartID, productID int64) (*models.CartItem, error) {
	query := `
		SELECT id, cart_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE cart_id = ? AND product_id = ?
	`

	cartItem := &models.CartItem{}
	err := r.db.QueryRowContext(ctx, query, cartID, productID).Scan(
		&cartItem.ID,
		&cartItem.CartID,
		&cartItem.ProductID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
		&cartItem.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return cartItem, nil
}

// GetCartItemByID retrieves a cart item by its ID
func (r *CartRepository) GetCartItemByID(ctx context.Context, itemID int64) (*models.CartItem, error) {
	query := `
		SELECT id, cart_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE id = ?
	`

	cartItem := &models.CartItem{}
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&cartItem.ID,
		&cartItem.CartID,
		&cartItem.ProductID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
		&cartItem.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("cart item not found")
		}
		return nil, err
	}

	return cartItem, nil
}

// UpdateCartItemQuantity updates the quantity of a cart item
func (r *CartRepository) UpdateCartItemQuantity(ctx context.Context, itemID int64, quantity int) (*models.CartItem, error) {
	query := `
		UPDATE cart_items
		SET quantity = ?, updated_at = ?
		WHERE id = ?
		RETURNING id, cart_id, product_id, quantity, created_at, updated_at
	`

	now := time.Now()
	cartItem := &models.CartItem{}
	err := r.db.QueryRowContext(ctx, query, quantity, now, itemID).Scan(
		&cartItem.ID,
		&cartItem.CartID,
		&cartItem.ProductID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
		&cartItem.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("cart item not found")
		}
		return nil, err
	}

	cartItem.UpdatedAt = &now
	return cartItem, nil
}

// RemoveCartItem removes a specific item from the cart
func (r *CartRepository) RemoveCartItem(ctx context.Context, itemID int64) error {
	query := `DELETE FROM cart_items WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, itemID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("cart item not found")
	}

	return nil
}

// GetCartItems retrieves all items in a cart with product details
func (r *CartRepository) GetCartItems(ctx context.Context, cartID int64) ([]*models.CartItem, error) {
	query := `
		SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at
		FROM cart_items ci
		WHERE ci.cart_id = ?
		ORDER BY ci.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartItems []*models.CartItem
	for rows.Next() {
		var item models.CartItem
		err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cartItems = append(cartItems, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cartItems, nil
}

// ClearCart removes all items from a cart
func (r *CartRepository) ClearCart(ctx context.Context, cartID int64) error {
	query := `DELETE FROM cart_items WHERE cart_id = ?`
	
	_, err := r.db.ExecContext(ctx, query, cartID)
	return err
}

// GetCartWithItems retrieves a cart with all its items and product details
func (r *CartRepository) GetCartWithItems(ctx context.Context, cartID int64) (*models.Cart, []*models.CartItem, error) {
	// Get the cart
	cart, err := r.GetCartByID(ctx, cartID)
	if err != nil {
		return nil, nil, err
	}

	// Get cart items
	cartItems, err := r.GetCartItems(ctx, cartID)
	if err != nil {
		return nil, nil, err
	}

	return cart, cartItems, nil
}

// GetCartWithItemsByUser retrieves a user's cart with all its items
func (r *CartRepository) GetCartWithItemsByUser(ctx context.Context, userID int64) (*models.Cart, []*models.CartItem, error) {
	// Get the cart
	cart, err := r.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	// Get cart items
	cartItems, err := r.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, nil, err
	}

	return cart, cartItems, nil
} 