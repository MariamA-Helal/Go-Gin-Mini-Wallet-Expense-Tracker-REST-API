package service

import (
	"context"
	"errors"
	"testing"
	"wallet-api/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------
// 1. Mock Repository Setup
// ---------------------------------------------------------

type mockUserRepo struct {
	users map[string]*models.User
}

func (m *mockUserRepo) CreateUserWithWallet(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.Username]; exists {
		return errors.New("user already exists")
	}
	user.ID = uint(len(m.users) + 1)
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if user, exists := m.users[username]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

// ---------------------------------------------------------
// 2. Table-Driven Tests
// ---------------------------------------------------------

func TestSignup_TableDriven(t *testing.T) {
	mockRepo := &mockUserRepo{users: make(map[string]*models.User)}
	svc := NewUserService(mockRepo)

	tests := []struct {
		name      string
		username  string
		password  string
		expectErr bool
	}{
		{"Valid Signup", "mariam", "securepass123", false},
		{"Short Username", "ma", "123456", true},
		{"Short Password", "ahmed", "123", true},
		{"Duplicate User", "mariam", "newpass123", true}, // Should fail because 'mariam' was created in test 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Signup(context.Background(), tt.username, tt.password)
			if (err != nil) != tt.expectErr {
				t.Errorf("Signup() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestLogin_TableDriven(t *testing.T) {
	mockRepo := &mockUserRepo{users: make(map[string]*models.User)}
	svc := NewUserService(mockRepo)

	// Pre-seed a valid user into our mock database
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	mockRepo.users["validuser"] = &models.User{
		ID:       1,
		Username: "validuser",
		Password: string(hashedPassword),
		Role:     "user",
	}

	tests := []struct {
		name      string
		username  string
		password  string
		expectErr bool
	}{
		{"Valid Login", "validuser", "correctpass", false},
		{"Wrong Password", "validuser", "wrongpass", true},
		{"Non-Existent User", "ghostuser", "anypass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.Login(context.Background(), tt.username, tt.password)
			if (err != nil) != tt.expectErr {
				t.Errorf("Login() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && token == "" {
				t.Error("Expected a JWT token, got an empty string")
			}
		})
	}
}
