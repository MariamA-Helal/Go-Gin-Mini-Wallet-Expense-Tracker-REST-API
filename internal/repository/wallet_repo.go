package repository

import (
	"context"
	"errors"
	"wallet-api/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalletRepository defines all database operations for a wallet
type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
	GetWalletByUserID(ctx context.Context, userID uint) (*models.Wallet, error)
	Deposit(ctx context.Context, userID uint, amount int64, category, note string) error
	Withdraw(ctx context.Context, userID uint, amount int64, category, note string) error
}

type walletRepo struct {
	db *gorm.DB
}

// NewWalletRepository creates a new instance of WalletRepository
func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepo{db: db}
}

// CreateWallet inserts a new wallet into the database (Used primarily for testing or specific internal actions)
func (r *walletRepo) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

// GetWalletByUserID fetches a wallet by its user ID
func (r *walletRepo) GetWalletByUserID(ctx context.Context, userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Deposit adds money and records the transaction atomically
func (r *walletRepo) Deposit(ctx context.Context, userID uint, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("deposit amount must be greater than zero")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		// Row-Level Locking: SELECT ... FOR UPDATE
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err // Will return error if user/wallet doesn't exist
		}

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			WalletID: wallet.ID,
			Type:     "deposit",
			Amount:   amount,
			Category: category,
			Note:     note,
		}
		return tx.Create(&transaction).Error
	})
}

// Withdraw subtracts money safely avoiding negative balances
func (r *walletRepo) Withdraw(ctx context.Context, userID uint, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("withdraw amount must be greater than zero")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		// Row-Level Locking: Prevents concurrent overdraws
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		// Edge Case: Insufficient Funds
		if wallet.Balance < amount {
			return errors.New("insufficient funds")
		}

		wallet.Balance -= amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			WalletID: wallet.ID,
			Type:     "withdraw",
			Amount:   amount,
			Category: category,
			Note:     note,
		}
		return tx.Create(&transaction).Error
	})
}
