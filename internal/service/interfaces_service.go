package service

import (
	"context"
	"wallet-api/internal/models"
)

// UserRepository defines all database operations for a user
type UserService interface {
	Signup(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
}

// WalletRepository defines all database operations for a wallet
type WalletService interface {
	GetWallet(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) (*models.Wallet, error)
	GetTransactionHistory(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string, filter models.TransactionFilter) ([]models.Transaction, error)
	GetMonthlySummary(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) ([]models.CategorySummary, error)
	Deposit(ctx context.Context, userID uint, amount int64, category, note string) (int64, error)

	// Budget functions
	SetBudget(ctx context.Context, userID uint, category string, limit int64) error
	GetBudgetStatus(ctx context.Context, userID uint) ([]models.BudgetStatus, error)

	// Updated functions to return Warning (string)
	Withdraw(ctx context.Context, userID uint, amount int64, category, note string) (int64, string, error)
	Transfer(ctx context.Context, senderUserID uint, receiverUsername string, amount int64, category, note string) (int64, string, error)
}
