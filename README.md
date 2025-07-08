# UCommerce

UCommerce is a simple e-commerce backend API written in Go, using [chi](https://github.com/go-chi/chi) for routing and SQLite for storage. It provides endpoints for user management, product browsing, cart operations, and order processing.

## Features

- User registration and authentication (JWT-based)
- Product listing and details
- Shopping cart management
- Order creation and history
- RESTful API with Swagger (OpenAPI) documentation

## Getting Started

### Prerequisites

- Go 1.18+
- [SQLite3](https://www.sqlite.org/index.html)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/musllim/ecommerce.git
   cd ecommerce
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run database migrations:**
   ```bash
   # Ensure SQLite3 is installed
   go run cmd/server/main.go
   ```
   The server will automatically run migrations on startup.

4. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```

   The server will start on the address specified in your config (default: `localhost:8080`).

### API Documentation

- Swagger UI is available at:  
  [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Project Structure
ucommerce/
cmd/server/ # Main entrypoint
docs/ # Swagger/OpenAPI docs
internal/
config/ # Configuration
handlers/ # HTTP handlers
middleware/ # Auth and other middleware
models/ # Data models
repository/ # Data access layer
service/ # Business logic
pkg/database/ # Database connection and migration
migrations/ # SQL migration files


### Authentication

- Most endpoints require a Bearer JWT token in the `Authorization` header.
- Example:  
  ```
  Authorization: Bearer <your-jwt-token>
  ```

### Example API Endpoints

- `POST /api/v1/users/register` – Register a new user
- `POST /api/v1/users/login` – Login and receive JWT
- `GET /api/v1/products` – List products
- `GET /api/v1/cart` – View cart (auth required)
- `POST /api/v1/cart/add` – Add item to cart (auth required)
- `POST /api/v1/orders` – Create order (auth required)

### Development

- Handlers are registered in `cmd/server/main.go`.
- Authentication middleware is in `internal/middleware/auth.go`.
- Business logic is in `internal/service/`.

### License

[Apache 2.0](LICENSE)

---

**Maintainer:** [musllim](https://github.com/musllim)