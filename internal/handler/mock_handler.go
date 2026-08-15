package handler

import (
	"context"
	"errors"
	"wallet-api/internal/models"
)

// ---------------------------------------------------------
// 1. Mock Wallet Service for Handler Integration Testing
// ---------------------------------------------------------
type mockWalletService struct{}

func (m *mockWalletService) GetWallet(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) (*models.Wallet, error) {
	return nil, nil
}
func (m *mockWalletService) GetTransactionHistory(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string, filter models.TransactionFilter) ([]models.Transaction, error) {
	return nil, nil
}
func (m *mockWalletService) GetMonthlySummary(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) ([]models.CategorySummary, error) {
	return nil, nil
}

func (m *mockWalletService) SetBudget(ctx context.Context, userID uint, category string, limit int64) error {
	return nil
}
func (m *mockWalletService) GetBudgetStatus(ctx context.Context, userID uint) ([]models.BudgetStatus, error) {
	return nil, nil
}

func (m *mockWalletService) Deposit(ctx context.Context, userID uint, amount int64, category, note string) (int64, error) {
	return 0, nil
}

func (m *mockWalletService) Withdraw(ctx context.Context, userID uint, amount int64, category, note string) (int64, string, error) {
	return 0, "", nil
}

func (m *mockWalletService) Transfer(ctx context.Context, senderUserID uint, receiverUsername string, amount int64, category, note string) (int64, string, error) {
	if receiverUsername == "ghost" {
		return 0, "", errors.New("receiver not found")
	}
	if amount > 10000 {
		return 0, "", errors.New("insufficient funds")
	}
	return 5000, "", nil
}

// ---------------------------------------------------------
// 2. Mock User Service for Handler Integration Testing
// ---------------------------------------------------------
type mockUserService struct{}

func (m *mockUserService) Signup(ctx context.Context, username, password string) error {
	if username == "existinguser" {
		return errors.New("user already exists")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func (m *mockUserService) Login(ctx context.Context, username, password string) (string, error) {
	if username == "validuser" && password == "SecurePass123*" {
		return "fake-jwt-token", nil
	}
	return "", errors.New("invalid credentials")
}
