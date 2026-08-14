package repository

import (
	"context"
	"wallet-api/internal/models"

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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		wallet := &models.Wallet{
			UserID:  user.ID,
			Balance: 0,
		}

		if err := tx.Create(wallet).Error; err != nil {
			return err
		}

		return nil
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
