package service

import (
	"context"
	"testing"
	"wallet-api/internal/models"
)

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
			_, _, err := svc.Transfer(context.Background(), tt.senderID, tt.receiverUsername, tt.amount, "Gift", "Note")
			if (err != nil) != tt.expectErr {
				t.Errorf("Transfer() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
