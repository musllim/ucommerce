// @title           UCommerce API
// @version         1.0
// @description     This is the API documentation for UCommerce e-commerce platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"log"
	"net/http"

	"github.com/musllim/ecommerce/internal/config"
	"github.com/musllim/ecommerce/internal/handlers"
	ecomiddleware "github.com/musllim/ecommerce/internal/middleware"
	"github.com/musllim/ecommerce/internal/repository"
	"github.com/musllim/ecommerce/internal/service"
	"github.com/musllim/ecommerce/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	_ "github.com/musllim/ecommerce/docs" // swagger docs - temporarily disabled due to network issues
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewSQLiteDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(cfg.MigrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	productRepo := repository.NewProductRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	cartRepo := repository.NewCartRepository(db.DB)
	orderRepo := repository.NewOrderRepository(db.DB)

	productService := service.NewProductService(productRepo)
	userService := service.NewUserService(userRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	authMiddleware := ecomiddleware.NewAuthMiddleware(userService)
	productHandler := handlers.NewProductHandler(productService, userService, authMiddleware)
	cartHandler := handlers.NewCartHandler(cartService, authMiddleware)
	orderHandler := handlers.NewOrderHandler(orderService, authMiddleware)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/cart", cartHandler.RegisterRoutes)
		r.Route("/orders", orderHandler.RegisterRoutes)
		r.Route("/products", productHandler.RegisterRoutes)
		r.Route("/users", userHandler.RegisterRoutes)
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
