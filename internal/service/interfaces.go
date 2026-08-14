package service

import (
	"context"
)

// UserService defines the business logic for users and authentication
type UserService interface {
	Signup(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
}
