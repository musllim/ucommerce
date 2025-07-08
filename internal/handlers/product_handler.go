// internal/handlers/product_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	ecomiddleware "github.com/musllim/ecommerce/internal/middleware"
	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/service"
)

type ProductHandler struct {
	service *service.ProductService
	userService *service.UserService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) RegisterRoutes(r chi.Router) {

	authMiddleware := ecomiddleware.NewAuthMiddleware(h.userService)

	r.Route("/", func(r chi.Router) {
		// Public routes
		r.Get("/", h.listProducts)
		r.Get("/{id}", h.getProduct)

		// Private routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Post("/", h.createProduct)
			r.Put("/{id}", h.updateProduct)
			r.Delete("/{id}", h.deleteProduct)
		})
	})
}

// @Summary      List products
// @Description  Get all products with optional filtering and pagination
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        limit     query     int     false  "Number of products to return (default: 10)"
// @Param        offset    query     int     false  "Number of products to skip (default: 0)"
// @Param        category  query     string  false  "Filter by category"
// @Success      200       {array}   models.Product
// @Failure      500       {object}  string
// @Router       /products [get]
func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	category := r.URL.Query().Get("category")

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	products, err := h.service.ListProducts(r.Context(), limit, offset, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// @Summary      Get product by ID
// @Description  Get a specific product by its ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  models.Product
// @Failure      400  {object}  string
// @Failure      404  {object}  string
// @Failure      500  {object}  string
// @Router       /products/{id} [get]
func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// @Summary      Create product
// @Description  Create a new product
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      models.Product  true  "Product object"
// @Success      201      {object}  models.Product
// @Failure      400      {object}  string
// @Security     BearerAuth
// @Router       /products [post]
func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateProduct(r.Context(), &product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// @Summary      Update product
// @Description  Update an existing product
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int             true   "Product ID"
// @Param        product  body      models.Product  true   "Product object"
// @Success      200      {object}  models.Product
// @Failure      400      {object}  string
// @Failure      404      {object}  string
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	product.ID = id

	if err := h.service.UpdateProduct(r.Context(), &product); err != nil {
		if err.Error() == "product not found" {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// @Summary      Delete product
// @Description  Delete a product by ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      204  {object}  string
// @Failure      400  {object}  string
// @Failure      404  {object}  string
// @Failure      500  {object}  string
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		if err.Error() == "product not found" {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/", h.listProducts)
	r.Get("/{id}", h.getProduct)
}

func (h *ProductHandler) RegisterPrivateRoutes(r chi.Router) {
	r.Post("/", h.createProduct)
	r.Put("/{id}", h.updateProduct)
	r.Delete("/{id}", h.deleteProduct)
}

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
