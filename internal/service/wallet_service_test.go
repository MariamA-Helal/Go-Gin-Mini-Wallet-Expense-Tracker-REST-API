package service

import (
	"context"
	"testing"
	"wallet-api/internal/models"
)

// ---------------------------------------------------------
// Mock Wallet Repository
// ---------------------------------------------------------
type mockWalletRepo struct{}

func (m *mockWalletRepo) CreateWallet(ctx context.Context, wallet *models.Wallet) error { return nil }
func (m *mockWalletRepo) GetWalletByUserID(ctx context.Context, userID uint) (*models.Wallet, error) {
	return &models.Wallet{UserID: userID, Balance: 1000}, nil
}
func (m *mockWalletRepo) Deposit(ctx context.Context, userID uint, amount int64, category, note string) error {
	return nil
}
func (m *mockWalletRepo) Withdraw(ctx context.Context, userID uint, amount int64, category, note string) error {
	return nil
}
func (m *mockWalletRepo) Transfer(ctx context.Context, senderUserID, receiverUserID uint, amount int64, category, note string) error {
	return nil // Assume DB operation succeeds for testing service logic
}

// ---------------------------------------------------------
// Tests
// ---------------------------------------------------------
func TestWalletService_Transfer_TableDriven(t *testing.T) {
	mockURepo := &mockUserRepo{users: make(map[string]*models.User)}
	mockWRepo := &mockWalletRepo{}

	// Seed mock users
	mockURepo.users["receiver_user"] = &models.User{ID: 2, Username: "receiver_user"}
	mockURepo.users["sender_user"] = &models.User{ID: 1, Username: "sender_user"}

	svc := NewWalletService(mockWRepo, mockURepo)

	tests := []struct {
		name             string
		senderID         uint
		receiverUsername string
		amount           int64
		expectErr        bool
	}{
		{"Valid Transfer", 1, "receiver_user", 500, false},
		{"Zero Amount", 1, "receiver_user", 0, true},
		{"Negative Amount", 1, "receiver_user", -100, true},
		{"Receiver Not Found", 1, "ghost_user", 500, true},
		{"Transfer to Self", 1, "sender_user", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Transfer(context.Background(), tt.senderID, tt.receiverUsername, tt.amount, "Gift", "Note")
			if (err != nil) != tt.expectErr {
				t.Errorf("Transfer() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
