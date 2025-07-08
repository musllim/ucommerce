// internal/handlers/order_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	ecomiddleware "github.com/musllim/ecommerce/internal/middleware"
	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/service"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	service        *service.OrderService
	authMiddleware *ecomiddleware.AuthMiddleware
}

func NewOrderHandler(service *service.OrderService, authMiddleware *ecomiddleware.AuthMiddleware) *OrderHandler {
	return &OrderHandler{service: service, authMiddleware: authMiddleware}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Use(h.authMiddleware.Authenticate)
	r.Post("/", h.createOrder)
	r.Get("/", h.getUserOrders)
	r.Get("/{id}", h.getOrderDetails)
}

// createOrder godoc
// @Summary      Create a new order
// @Description  Creates a new order for the authenticated user
// @Tags         orders
// @Accept       json
// @Produce      json
// @Success      201 {object} models.Order
// @Failure      400 {string} string "Bad Request"
// @Security     BearerAuth
// @Failure      401 {string} string "Unauthorized"
// @Router       /orders [post]
func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := user.ID

	order, err := h.service.CreateOrder(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// getUserOrders godoc
// @Summary      Get user's orders
// @Description  Retrieves all orders for the authenticated user
// @Tags         orders
// @Produce      json
// @Success      200 {array} models.Order
// @Security     BearerAuth
// @Failure      401 {string} string "Unauthorized"
// @Router       /orders [get]
func (h *OrderHandler) getUserOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := user.ID

	orders, err := h.service.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// getOrderDetails godoc
// @Summary      Get order details
// @Description  Retrieves details for a specific order of the authenticated user
// @Tags         orders
// @Produce      json
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {string}  string "Invalid Order ID"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      404  {string}  string "Order Not Found"
// @Security     BearerAuth
// @Router       /orders/{id} [get]
func (h *OrderHandler) getOrderDetails(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := user.ID

	order, items, err := h.service.GetOrderDetails(r.Context(), orderID, userID)
	if err != nil {
		if err.Error() == "order not found" {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"order": order,
		"items": items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
