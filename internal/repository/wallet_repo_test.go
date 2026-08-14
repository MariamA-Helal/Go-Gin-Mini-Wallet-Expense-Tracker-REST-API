package repository

import (
	"context"
	"testing"
	"wallet-api/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB prepares a fast in-memory SQLite database for testing (Shared across the repository package)
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{})
	return db
}

func TestDepositAndWithdraw_TableDriven(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)

	// Setup a valid user and wallet
	user := models.User{Username: "fintech_user", Password: "pass", Role: "user"}
	db.Create(&user)
	wallet := models.Wallet{UserID: user.ID, Balance: 10000}
	db.Create(&wallet)

	depositTests := []struct {
		name      string
		userID    uint
		amount    int64
		category  string
		expectErr bool
	}{
		{"Valid Deposit", user.ID, 5000, "Salary", false},
		{"Zero Deposit", user.ID, 0, "None", true},
		{"Negative Deposit", user.ID, -200, "None", true},
		{"Non-Existent User Deposit", 999, 1000, "Salary", true}, // Edge Case
	}

	for _, tt := range depositTests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Deposit(context.Background(), tt.userID, tt.amount, tt.category, "note")
			if (err != nil) != tt.expectErr {
				t.Errorf("Deposit() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}

	withdrawTests := []struct {
		name      string
		userID    uint
		amount    int64
		category  string
		expectErr bool
	}{
		{"Valid Withdraw", user.ID, 2000, "Food", false},
		{"Zero Withdraw", user.ID, 0, "None", true},
		{"Negative Withdraw", user.ID, -100, "None", true},
		{"Insufficient Funds", user.ID, 20000, "Shopping", true},
		{"Non-Existent User Withdraw", 999, 1000, "Cash", true}, // Edge Case
	}

	for _, tt := range withdrawTests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Withdraw(context.Background(), tt.userID, tt.amount, tt.category, "note")
			if (err != nil) != tt.expectErr {
				t.Errorf("Withdraw() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
