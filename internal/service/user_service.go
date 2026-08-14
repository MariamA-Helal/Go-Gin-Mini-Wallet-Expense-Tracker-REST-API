package service

import (
	"context"
	"errors"
	"time"
	"wallet-api/internal/models"
	"wallet-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// In a real app, this should be loaded from environment variables (.env)
var jwtSecretKey = []byte("super-secret-key-change-in-production")

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new instance of UserService
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// Signup hashes the password and creates the user (with their auto-wallet via repo)
func (s *userService) Signup(ctx context.Context, username, password string) error {
	if len(username) < 3 || len(password) < 6 {
		return errors.New("invalid input: username must be >= 3 and password >= 6 characters")
	}

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     "user", // Default role
	}

	return s.repo.CreateUserWithWallet(ctx, user)
}

// Login verifies credentials and returns a JWT token
func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	// 1. Fetch user from DB
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials") // We don't expose if the user exists or not for security
	}

	// 2. Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// 3. Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
