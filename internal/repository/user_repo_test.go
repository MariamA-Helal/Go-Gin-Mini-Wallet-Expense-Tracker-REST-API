package repository

import (
	"context"
	"testing"
	"wallet-api/internal/models"
)

func TestCreateUserWithWallet_TableDriven(t *testing.T) {
	db := setupTestDB(t) // This function is defined in wallet_repo_test.go
	repo := NewUserRepository(db)

	tests := []struct {
		name      string
		user      *models.User
		expectErr bool
	}{
		{
			name:      "Valid User Creation",
			user:      &models.User{Username: "newuser", Password: "hashedpassword", Role: "user"},
			expectErr: false,
		},
		{
			name:      "Duplicate Username",
			user:      &models.User{Username: "newuser", Password: "anotherpassword", Role: "user"},
			expectErr: true, // Should fail because username must be unique
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateUserWithWallet(context.Background(), tt.user)
			if (err != nil) != tt.expectErr {
				t.Errorf("CreateUserWithWallet() error = %v, expectErr %v", err, tt.expectErr)
			}

			// If success, verify wallet creation
			if !tt.expectErr {
				var wallet models.Wallet
				db.Where("user_id = ?", tt.user.ID).First(&wallet)
				if wallet.Balance != 0 {
					t.Errorf("Expected initial wallet balance to be 0, got %d", wallet.Balance)
				}
			}
		})
	}
}

func TestGetUserByUsername_TableDriven(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Seed a user
	db.Create(&models.User{Username: "testlogin", Password: "password123"})

	tests := []struct {
		name      string
		username  string
		expectErr bool
	}{
		{"Existing User", "testlogin", false},
		{"Non-Existing User", "ghostuser", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.GetUserByUsername(context.Background(), tt.username)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetUserByUsername() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
