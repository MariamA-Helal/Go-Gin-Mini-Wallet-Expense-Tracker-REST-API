package repository

import (
	"context"
	"errors"
	"wallet-api/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

func (r *walletRepo) Deposit(ctx context.Context, userID uint, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("deposit amount must be greater than zero")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		// Row-Level Locking: SELECT ... FOR UPDATE
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return errors.New("wallet not found")
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

// Transfer moves money between two wallets securely, avoiding deadlocks
func (r *walletRepo) Transfer(ctx context.Context, senderWalletID, receiverWalletID uint, amount int64, category, note string) error {
	if amount <= 0 {
		return errors.New("transfer amount must be greater than zero")
	}
	if senderWalletID == receiverWalletID {
		return errors.New("cannot transfer money to yourself")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var firstWallet, secondWallet models.Wallet
		var firstID, secondID uint

		// Order IDs in ascending order to prevent deadlocks under concurrent execution
		if senderWalletID < receiverWalletID {
			firstID, secondID = senderWalletID, receiverWalletID
		} else {
			firstID, secondID = receiverWalletID, senderWalletID
		}

		// Lock the first wallet row (lower ID)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&firstWallet, firstID).Error; err != nil {
			return errors.New("wallet not found")
		}
		// Lock the second wallet row (higher ID)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&secondWallet, secondID).Error; err != nil {
			return errors.New("wallet not found")
		}

		// Map sender and receiver pointers based on locked wallet instances
		var sender, receiver *models.Wallet
		if senderWalletID == firstWallet.ID {
			sender = &firstWallet
			receiver = &secondWallet
		} else {
			sender = &secondWallet
			receiver = &firstWallet
		}

		if sender.Balance < amount {
			return errors.New("insufficient funds")
		}

		// Apply balance updates
		sender.Balance -= amount
		receiver.Balance += amount

		if err := tx.Save(sender).Error; err != nil {
			return err
		}
		if err := tx.Save(receiver).Error; err != nil {
			return err
		}

		// Record audit transactions (Transfer Out / Transfer In)
		outTx := models.Transaction{
			WalletID:        sender.ID,
			Type:            "transfer_out",
			Amount:          amount,
			Category:        category,
			Note:            note,
			RelatedWalletID: &receiver.ID,
		}
		inTx := models.Transaction{
			WalletID:        receiver.ID,
			Type:            "transfer_in",
			Amount:          amount,
			Category:        category,
			Note:            note,
			RelatedWalletID: &sender.ID,
		}

		if err := tx.Create(&outTx).Error; err != nil {
			return err
		}
		if err := tx.Create(&inTx).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetTransactionsByWalletID fetches history with dynamic filters
func (r *walletRepo) GetTransactionsByWalletID(ctx context.Context, walletID uint, filter models.TransactionFilter) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := r.db.WithContext(ctx).Where("wallet_id = ?", walletID)

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.From != "" {
		query = query.Where("created_at >= ?", filter.From)
	}
	if filter.To != "" {
		query = query.Where("created_at <= ?", filter.To)
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at desc").Limit(filter.Limit).Offset(offset).Find(&transactions).Error

	return transactions, err
}

// GetMonthlySummary groups transactions by category for the current month
func (r *walletRepo) GetMonthlySummary(ctx context.Context, walletID uint) ([]models.CategorySummary, error) {
	var summary []models.CategorySummary

	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("category, SUM(amount) as total").
		Where("wallet_id = ? AND created_at >= date_trunc('month', CURRENT_DATE)", walletID).
		Group("category").
		Scan(&summary).Error

	return summary, err
}

func (r *walletRepo) SetBudget(ctx context.Context, budget *models.Budget) error {
	var existing models.Budget
	err := r.db.WithContext(ctx).Where("user_id = ? AND category = ?", budget.UserID, budget.Category).First(&existing).Error
	if err == nil {
		existing.MonthlyLimit = budget.MonthlyLimit
		return r.db.Save(&existing).Error
	}
	return r.db.Create(budget).Error
}

func (r *walletRepo) GetBudgetByCategory(ctx context.Context, userID uint, category string) (*models.Budget, error) {
	var budget models.Budget
	err := r.db.WithContext(ctx).Where("user_id = ? AND category = ?", userID, category).First(&budget).Error
	return &budget, err
}

func (r *walletRepo) GetAllBudgets(ctx context.Context, userID uint) ([]models.Budget, error) {
	var budgets []models.Budget
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&budgets).Error
	return budgets, err
}
