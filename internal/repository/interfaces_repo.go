package repository

import (
	"context"
	"wallet-api/internal/models"
)

// UserRepository defines all database operations for a user
type UserRepository interface {
	CreateUserWithWallet(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

// WalletRepository defines all database operations for a wallet
type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
	GetWalletByUserID(ctx context.Context, userID uint) (*models.Wallet, error)
	Deposit(ctx context.Context, userID uint, amount int64, category, note string) error
	Withdraw(ctx context.Context, userID uint, amount int64, category, note string) error
	Transfer(ctx context.Context, senderUserID, receiverUserID uint, amount int64, category, note string) error

	GetTransactionsByWalletID(ctx context.Context, walletID uint, filter models.TransactionFilter) ([]models.Transaction, error)
	GetMonthlySummary(ctx context.Context, walletID uint) ([]models.CategorySummary, error)

	SetBudget(ctx context.Context, budget *models.Budget) error
	GetBudgetByCategory(ctx context.Context, userID uint, category string) (*models.Budget, error)
	GetAllBudgets(ctx context.Context, userID uint) ([]models.Budget, error)
}
