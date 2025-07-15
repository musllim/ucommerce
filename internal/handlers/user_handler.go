// internal/handlers/user_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/service"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.createUser)
	r.Get("/", h.getAllUsers)
	r.Get("/{id}", h.getUser)
	r.Post("/login", h.login)
	r.Post("/oauth-login", h.oauthLogin)
}

// @Summary      Create user
// @Description  Register a new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserInput  true  "UserInput object"
// @Success      201   {object}  models.UserInput
// @Failure      400   {object}  string
// @Security     BearerAuth
// @Router       /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var input models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Always set role to 'user' regardless of input
	input.Role = "user"

	err := h.service.CreateUser(r.Context(), input)
	if err != nil {
		fmt.Println("Error creating user:", input)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

// @Summary      Get user by ID
// @Description  Get a specific user by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      400  {object}  string
// @Failure      404  {object}  string
// @Failure      500  {object}  string
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary      User login
// @Description  Authenticate user and return user info (in production, return JWT)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "User credentials"
// @Success      200          {object}  models.User
// @Failure      400          {object}  string
// @Failure      401          {object}  string
// @Router       /users/login [post]
func (h *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary      OAuth login
// @Description  Authenticate user via OAuth and return JWT token
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        oauth_data  body      OAuthLoginRequest  true  "OAuth user data"
// @Success      200         {object}  OAuthLoginResponse
// @Failure      400         {object}  string
// @Router       /users/oauth-login [post]
func (h *UserHandler) oauthLogin(w http.ResponseWriter, r *http.Request) {
	var req OAuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	token, err := h.service.OAuthLogin(r.Context(), req.Email, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}



	response := OAuthLoginResponse{
		Token: token,
		User:  OAuthUser(req),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type OAuthLoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type OAuthUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type OAuthLoginResponse struct {
	Token string    `json:"token"`
	User  OAuthUser `json:"user"`
}

// @Summary      Get all users
// @Description  Get all users with optional pagination
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        limit   query     int  false  "Number of users to return (default: 10)"
// @Param        offset  query     int  false  "Number of users to skip (default: 0)"
// @Success      200     {object}  UsersResponse
// @Failure      500     {object}  string
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	users, err := h.service.GetAllUsers(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	total, err := h.service.CountUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UsersResponse{
		Total: total,
		Data:  users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type UsersResponse struct {
	Total int64        `json:"total"`
	Data  []*models.User `json:"data"`
} 


