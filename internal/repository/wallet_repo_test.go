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

func TestTransfer_TableDriven(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)

	// Setup users and wallets
	user1 := models.User{Username: "sender", Password: "123", Role: "user"}
	user2 := models.User{Username: "receiver", Password: "123", Role: "user"}
	db.Create(&user1)
	db.Create(&user2)

	wallet1 := models.Wallet{UserID: user1.ID, Balance: 10000} // Sender has 100 EGP (10000 Cents)
	wallet2 := models.Wallet{UserID: user2.ID, Balance: 5000}  // Receiver has 50 EGP (5000 Cents)
	db.Create(&wallet1)
	db.Create(&wallet2)

	tests := []struct {
		name       string
		senderID   uint
		receiverID uint
		amount     int64
		expectErr  bool
	}{
		{"Valid Transfer", user1.ID, user2.ID, 2000, false},
		{"Zero Amount Transfer", user1.ID, user2.ID, 0, true},
		{"Negative Amount Transfer", user1.ID, user2.ID, -500, true},
		{"Insufficient Funds", user1.ID, user2.ID, 20000, true}, // Trying to send 200 EGP while having 100
		{"Transfer to Self", user1.ID, user1.ID, 1000, true},    // Cannot send to yourself
		{"Sender Not Found", 999, user2.ID, 1000, true},
		{"Receiver Not Found", user1.ID, 999, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Transfer(context.Background(), tt.senderID, tt.receiverID, tt.amount, "Gift", "Test Note")
			if (err != nil) != tt.expectErr {
				t.Errorf("Transfer() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}

	// Verify final balances after the only valid transfer (2000 cents)
	var w1, w2 models.Wallet
	db.First(&w1, wallet1.ID)
	db.First(&w2, wallet2.ID)

	if w1.Balance != 8000 {
		t.Errorf("Expected sender balance 8000, got %d", w1.Balance)
	}
	if w2.Balance != 7000 {
		t.Errorf("Expected receiver balance 7000, got %d", w2.Balance)
	}
}
