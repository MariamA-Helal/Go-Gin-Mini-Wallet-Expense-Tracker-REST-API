package service

import (
	"context"
	"errors"
	"wallet-api/internal/models"
	"wallet-api/internal/repository"
)

type walletService struct {
	walletRepo repository.WalletRepository
	userRepo   repository.UserRepository
}

// NewWalletService creates a new instance of WalletService
func NewWalletService(wRepo repository.WalletRepository, uRepo repository.UserRepository) WalletService {
	return &walletService{
		walletRepo: wRepo,
		userRepo:   uRepo,
	}
}

func (s *walletService) GetWallet(ctx context.Context, userID uint) (*models.Wallet, error) {
	return s.walletRepo.GetWalletByUserID(ctx, userID)
}

func (s *walletService) Deposit(ctx context.Context, userID uint, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return s.walletRepo.Deposit(ctx, userID, amount, category, note)
}

func (s *walletService) Withdraw(ctx context.Context, userID uint, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return s.walletRepo.Withdraw(ctx, userID, amount, category, note)
}

// Transfer is where the core business logic shines
func (s *walletService) Transfer(ctx context.Context, senderUserID uint, receiverUsername string, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	// 1. Find the receiver by their username
	receiverUser, err := s.userRepo.GetUserByUsername(ctx, receiverUsername)
	if err != nil {
		return errors.New("receiver not found")
	}

	// 2. Prevent sending money to oneself
	if senderUserID == receiverUser.ID {
		return errors.New("cannot transfer money to yourself")
	}

	// 3. Proceed with the database transaction via repository
	return s.walletRepo.Transfer(ctx, senderUserID, receiverUser.ID, amount, category, note)
}
