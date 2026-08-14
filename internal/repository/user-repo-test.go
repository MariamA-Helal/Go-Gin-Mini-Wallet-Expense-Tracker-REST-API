package repository
package repository

import (
	"context"
	"testing"
	"your_module_name/internal/models"
)

func TestCreateUserWithWallet(t *testing.T) {
	db := setupTestDB(t) // Using the same test DB setup function from the wallet test
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "newuser",
		Password: "hashedpassword",
		Role:     "user",
	}

	// 1. Test execution
	err := repo.CreateUserWithWallet(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user with wallet: %v", err)
	}

	// Ensure the user received an ID from the database
	if user.ID == 0 {
		t.Error("Expected user ID to be set, got 0")
	}

	// 2. Ensure the wallet was automatically created in the same database
	var wallet models.Wallet
	err = db.Where("user_id = ?", user.ID).First(&wallet).Error
	if err != nil {
		t.Fatalf("Expected wallet to be created, but got error: %v", err)
	}

	if wallet.Balance != 0 {
		t.Errorf("Expected initial wallet balance to be 0, got %d", wallet.Balance)
	}
}

func TestGetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testlogin",
		Password: "password123",
	}
	db.Create(user)

	fetchedUser, err := repo.GetUserByUsername(context.Background(), "testlogin")
	if err != nil {
		t.Fatalf("Failed to fetch user: %v", err)
	}

	if fetchedUser.Username != "testlogin" {
		t.Errorf("Expected username 'testlogin', got '%s'", fetchedUser.Username)
	}
}