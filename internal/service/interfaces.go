package service

import (
	"context"
	"wallet-api/internal/models"
)

// UserService defines the business logic for users and authentication
type UserService interface {
	Signup(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
}

// WalletService defines the business logic for wallet operations
type WalletService interface {
	GetWallet(ctx context.Context, userID uint) (*models.Wallet, error)
	Deposit(ctx context.Context, userID uint, amount int64, category, note string) error
	Withdraw(ctx context.Context, userID uint, amount int64, category, note string) error
	Transfer(ctx context.Context, senderUserID uint, receiverUsername string, amount int64, category, note string) error
}
