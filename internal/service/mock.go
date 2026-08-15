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
