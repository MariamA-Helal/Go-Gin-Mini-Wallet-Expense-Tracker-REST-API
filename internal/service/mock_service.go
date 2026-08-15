package service

import (
	"context"
	"errors"
	"wallet-api/internal/models"
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
// 2. Mock Wallet Repository
// ---------------------------------------------------------

type mockWalletRepo struct{}

func (m *mockWalletRepo) CreateWallet(ctx context.Context, wallet *models.Wallet) error { return nil }

func (m *mockWalletRepo) GetWalletByUserID(ctx context.Context, userID uint) (*models.Wallet, error) {
	return &models.Wallet{
		ID:      1,
		UserID:  userID,
		Balance: 10000,
	}, nil
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

func (m *mockWalletRepo) GetTransactionsByWalletID(ctx context.Context, walletID uint, filter models.TransactionFilter) ([]models.Transaction, error) {
	return []models.Transaction{}, nil
}

func (m *mockWalletRepo) SetBudget(ctx context.Context, budget *models.Budget) error {
	return nil
}

func (m *mockWalletRepo) GetBudgetByCategory(ctx context.Context, userID uint, category string) (*models.Budget, error) {
	return nil, errors.New("not found")
}

func (m *mockWalletRepo) GetMonthlySummary(ctx context.Context, walletID uint) ([]models.CategorySummary, error) {
	return []models.CategorySummary{}, nil
}

func (m *mockWalletRepo) GetAllBudgets(ctx context.Context, userID uint) ([]models.Budget, error) {
	return []models.Budget{}, nil
}
