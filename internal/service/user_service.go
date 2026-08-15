package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"
	"unicode"
	"wallet-api/internal/models"
	"wallet-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("super-secret-key-change-in-production-fallback")
	}
	return []byte(secret)
}

var jwtSecretKey = getJWTSecret()

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// ---------------- Validation Helpers ----------------

func validateUsername(username string) error {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return errors.New("username cannot be empty or whitespace")
	}
	if strings.Contains(username, " ") {
		return errors.New("username cannot contain spaces")
	}
	if len(username) < 10 {
		return errors.New("username must be at least 10 characters long")
	}
	if len(username) > 30 {
		return errors.New("username is too long (max 30 characters)")
	}
	return nil
}

func validatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		if char > unicode.MaxASCII {
			return errors.New("password contains invalid characters (only English letters and standard symbols allowed)")
		}

		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char) || unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	return nil
}

// ---------------- Service Methods ----------------

func (s *userService) Signup(ctx context.Context, username, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePasswordComplexity(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	return s.repo.CreateUserWithWallet(ctx, user)
}

// ... كملي باقي الكود بتاع الـ Login زي ما هو

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString(jwtSecretKey)
}
