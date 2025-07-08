// internal/handlers/cart_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	ecomiddleware "github.com/musllim/ecommerce/internal/middleware"
	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/service"
)

type CartHandler struct {
	service        *service.CartService
	authMiddleware *ecomiddleware.AuthMiddleware
}

func (h *CartHandler) RegisterRoutes(r chi.Router) {
	r.Use(h.authMiddleware.Authenticate)
	r.Get("/", h.getCartItems)
	r.Post("/items", h.addToCart)
	r.Put("/items/{productID}", h.updateCartItem)
	r.Delete("/items/{productID}", h.removeFromCart)
}

func NewCartHandler(service *service.CartService, authMiddleware *ecomiddleware.AuthMiddleware) *CartHandler {
	return &CartHandler{service: service, authMiddleware: authMiddleware}
}

func getUserIDFromContext(r *http.Request) (int64, bool) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok || user == nil {
		return 0, false
	}
	return user.ID, true
}

// @Summary      Get cart items
// @Description  Get all items in the user's cart
// @Tags         cart
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   models.CartItem
// @Failure      401  {object}  string
// @Failure      500  {object}  string
// @Router       /cart [get]
func (h *CartHandler) getCartItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.service.GetCartItems(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// @Summary      Add item to cart
// @Description  Add a product to the user's cart
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        item  body      object  true  "Product ID and quantity"
// @Success      200   {object}  string
// @Failure      400   {object}  string
// @Failure      401   {object}  string
// @Router       /cart/items [post]
func (h *CartHandler) addToCart(w http.ResponseWriter, r *http.Request) {
	fmt.Println("addToCart")
	var req struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.AddToCart(r.Context(), userID, req.ProductID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Update cart item
// @Description  Update the quantity of a product in the user's cart
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        productID  path      int     true  "Product ID"
// @Param        quantity   body      object  true  "Quantity"
// @Success      200        {object}  string
// @Failure      400        {object}  string
// @Failure      401        {object}  string
// @Router       /cart/items/{productID} [put]
func (h *CartHandler) updateCartItem(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "productID")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.UpdateCartItemQuantity(r.Context(), userID, productID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Remove item from cart
// @Description  Remove a product from the user's cart
// @Tags         cart
// @Security     BearerAuth
// @Produce      json
// @Param        productID  path      int  true  "Product ID"
// @Success      204        {object}  string
// @Failure      400        {object}  string
// @Failure      401        {object}  string
// @Router       /cart/items/{productID} [delete]
func (h *CartHandler) removeFromCart(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "productID")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.RemoveFromCart(r.Context(), userID, productID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
