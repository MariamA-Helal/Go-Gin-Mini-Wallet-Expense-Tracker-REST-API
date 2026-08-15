package service

import (
	"context"
	"testing"
	"wallet-api/internal/models"

	"golang.org/x/crypto/bcrypt"
)

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
		{"Valid Signup", "mariam1234", "SecurePass123*", false},
		{"Short Username", "mariam", "SecurePass123*", true},
		{"Short Password", "ahmed", "A1*bc", true},
		{"No Uppercase", "ahmed", "securepass123*", true},
		{"No Number", "ahmed", "SecurePass*", true},
		{"No Special Char", "ahmed", "SecurePass123", true},
		{"Duplicate User", "mariam", "NewSecurePass123*", true},

		{"Empty Username", "", "SecurePass123*", true},
		{"Whitespace Username", "   ", "SecurePass123*", true},
		{"Username With Spaces", "mariam helal", "SecurePass123*", true},
		{"Arabic Password", "mariam123", "باسووردقوي123*", true},
		{"Very Long Username", "thisusernameiswaytoolongtobeaccepted", "SecurePass123*", true},
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

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass123*"), bcrypt.DefaultCost)
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
		{"Valid Login", "validuser", "CorrectPass123*", false},
		{"Wrong Password", "validuser", "WrongPass123*", true},
		{"Non-Existent User", "ghostuser", "AnyPass123*", true},
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
