package errors

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInvalidInput      = errors.New("invalid input provided")
	ErrInternal          = errors.New("an internal server error occurred")
)
