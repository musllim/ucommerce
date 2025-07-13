// internal/service/user_service.go
package service

import (
	"context"
	"errors"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/musllim/ecommerce/internal/models"
	"github.com/musllim/ecommerce/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key") // Use a secure value in production!

type UserService struct {
	repo *repository.UserRepository
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, user models.UserInput) error {
	if user.Name == "" {
		return errors.New("user name is required")
	}
	if user.Email == "" {
		return errors.New("user email is required")
	}
	if user.Password == "" {
		return errors.New("user password is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.repo.Create(ctx, &models.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	})
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	if tokenString == "" {
		return nil, errors.New("invalid token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}
	userID := int64(userIDFloat)

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *UserService) OAuthLogin(ctx context.Context, email, name string) (string, error) {
	// Try to find existing user by email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// User doesn't exist, create new user
		// For OAuth users, we'll set a placeholder password since DB requires it
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("oauth-user-"+email), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		
		user = &models.User{
			Name:     name,
			Email:    email,
			Password: string(hashedPassword),
		}
		
		err = s.repo.Create(ctx, user)
		if err != nil {
			return "", err
		}
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
