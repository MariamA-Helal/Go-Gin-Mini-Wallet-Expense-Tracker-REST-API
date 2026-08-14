package repository

import (
	"context"
	"your_module_name/internal/models"

	"gorm.io/gorm"
)

// UserRepository defines database operations for a user
type UserRepository interface {
	CreateUserWithWallet(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// CreateUserWithWallet creates a user and an empty wallet in a single transaction
func (r *userRepo) CreateUserWithWallet(ctx context.Context, user *models.User) error {
	// Start the Database Transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create the user first
		if err := tx.Create(user).Error; err != nil {
			return err // If it fails, GORM will automatically rollback the transaction
		}

		// 2. Prepare the wallet and link it to the newly created user's ID
		wallet := &models.Wallet{
			UserID:  user.ID,
			Balance: 0, // Initial balance is 0 cents
		}

		// 3. Create the wallet
		if err := tx.Create(wallet).Error; err != nil {
			return err // If this fails, it rolls back both the wallet and the user creation
		}

		return nil // If we reach this point, the transaction is committed successfully
	})
}

// GetUserByUsername is used for the Login process
func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
