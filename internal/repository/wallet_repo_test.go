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

	// Setup users
	user1 := models.User{Username: "sender", Password: "123", Role: "user"}
	user2 := models.User{Username: "receiver", Password: "123", Role: "user"}
	db.Create(&user1)
	db.Create(&user2)

	// Setup wallets explicitly linked to users
	wallet1 := models.Wallet{UserID: user1.ID, Balance: 10000}
	wallet2 := models.Wallet{UserID: user2.ID, Balance: 5000}
	db.Create(&wallet1)
	db.Create(&wallet2)

	tests := []struct {
		name          string
		setupBalances func(w1, w2 *models.Wallet)
		senderArg     uint
		receiverArg   uint
		amount        int64
		expectErr     bool
		expectedW1Bal int64
		expectedW2Bal int64
	}{
		{
			name: "Valid Transfer",
			setupBalances: func(w1, w2 *models.Wallet) {
				w1.Balance = 10000
				w2.Balance = 5000
				db.Save(w1)
				db.Save(w2)
			},
			senderArg:     wallet1.ID,
			receiverArg:   wallet2.ID,
			amount:        2000,
			expectErr:     false,
			expectedW1Bal: 8000,
			expectedW2Bal: 7000,
		},
		{
			name: "Zero Amount Transfer",
			setupBalances: func(w1, w2 *models.Wallet) {
				w1.Balance = 10000
				w2.Balance = 5000
				db.Save(w1)
				db.Save(w2)
			},
			senderArg:     wallet1.ID,
			receiverArg:   wallet2.ID,
			amount:        0,
			expectErr:     true,
			expectedW1Bal: 10000,
			expectedW2Bal: 5000,
		},
		{
			name: "Negative Amount Transfer",
			setupBalances: func(w1, w2 *models.Wallet) {
				w1.Balance = 10000
				w2.Balance = 5000
				db.Save(w1)
				db.Save(w2)
			},
			senderArg:     wallet1.ID,
			receiverArg:   wallet2.ID,
			amount:        -500,
			expectErr:     true,
			expectedW1Bal: 10000,
			expectedW2Bal: 5000,
		},
		{
			name: "Insufficient Funds",
			setupBalances: func(w1, w2 *models.Wallet) {
				w1.Balance = 10000
				w2.Balance = 5000
				db.Save(w1)
				db.Save(w2)
			},
			senderArg:     wallet1.ID,
			receiverArg:   wallet2.ID,
			amount:        20000,
			expectErr:     true,
			expectedW1Bal: 10000,
			expectedW2Bal: 5000,
		},
		{
			name: "Transfer to Self",
			setupBalances: func(w1, w2 *models.Wallet) {
				w1.Balance = 10000
				w2.Balance = 5000
				db.Save(w1)
				db.Save(w2)
			},
			senderArg:     wallet1.ID,
			receiverArg:   wallet1.ID,
			amount:        1000,
			expectErr:     true,
			expectedW1Bal: 10000,
			expectedW2Bal: 5000,
		},
		{
			name:          "Sender Not Found",
			setupBalances: func(w1, w2 *models.Wallet) {},
			senderArg:     999,
			receiverArg:   wallet2.ID,
			amount:        1000,
			expectErr:     true,
		},
		{
			name: "Receiver Not Found",
			setupBalances: func(w1, w2 *models.Wallet) {
				w1.Balance = 10000
				db.Save(w1)
			},
			senderArg:   wallet1.ID,
			receiverArg: 999,
			amount:      1000,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupBalances != nil {
				tt.setupBalances(&wallet1, &wallet2)
			}

			err := repo.Transfer(context.Background(), tt.senderArg, tt.receiverArg, tt.amount, "Gift", "Test Note")
			if (err != nil) != tt.expectErr {
				t.Errorf("Transfer() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr {
				var w1, w2 models.Wallet
				db.First(&w1, wallet1.ID)
				db.First(&w2, wallet2.ID)

				if w1.Balance != tt.expectedW1Bal {
					t.Errorf("Expected sender balance %d, got %d", tt.expectedW1Bal, w1.Balance)
				}
				if w2.Balance != tt.expectedW2Bal {
					t.Errorf("Expected receiver balance %d, got %d", tt.expectedW2Bal, w2.Balance)
				}
			}
		})
	}
}
