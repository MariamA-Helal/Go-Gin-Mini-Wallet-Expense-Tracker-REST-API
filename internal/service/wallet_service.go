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

// NewWalletService initializes a new WalletService with repository dependencies.
func NewWalletService(wRepo repository.WalletRepository, uRepo repository.UserRepository) WalletService {
	return &walletService{walletRepo: wRepo, userRepo: uRepo}
}

// ----------------- Security Helper -----------------

// resolveTargetUserID implements RBAC logic and resolves optional target usernames to internal user IDs.
func (s *walletService) resolveTargetUserID(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) (uint, error) {
	if requestedUsername == "" {
		return tokenUserID, nil
	}

	targetUser, err := s.userRepo.GetUserByUsername(ctx, requestedUsername)
	if err != nil {
		return 0, errors.New("requested user not found")
	}

	if targetUser.ID != tokenUserID && tokenRole != "admin" {
		return 0, errors.New("forbidden: only admins can view other users' wallets")
	}

	return targetUser.ID, nil
}

// ----------------- Read Operations (With RBAC) -----------------

// GetWallet retrieves wallet details for a user with RBAC enforcement.
func (s *walletService) GetWallet(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) (*models.Wallet, error) {
	targetUserID, err := s.resolveTargetUserID(ctx, tokenUserID, tokenRole, requestedUsername)
	if err != nil {
		return nil, err
	}

	return s.walletRepo.GetWalletByUserID(ctx, targetUserID)
}

// GetTransactionHistory fetches paginated and filtered transactions for a target wallet.
func (s *walletService) GetTransactionHistory(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string, filter models.TransactionFilter) ([]models.Transaction, error) {
	targetUserID, err := s.resolveTargetUserID(ctx, tokenUserID, tokenRole, requestedUsername)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	return s.walletRepo.GetTransactionsByWalletID(ctx, wallet.ID, filter)
}

// GetMonthlySummary returns category-wise monthly expenditure aggregates.
func (s *walletService) GetMonthlySummary(ctx context.Context, tokenUserID uint, tokenRole, requestedUsername string) ([]models.CategorySummary, error) {
	targetUserID, err := s.resolveTargetUserID(ctx, tokenUserID, tokenRole, requestedUsername)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	return s.walletRepo.GetMonthlySummary(ctx, wallet.ID)
}

// Deposit processes a credit transaction to increase the user's wallet balance.
func (s *walletService) Deposit(ctx context.Context, userID uint, amount int64, category, note string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	if err := s.walletRepo.Deposit(ctx, userID, amount, category, note); err != nil {
		return 0, err
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	return wallet.Balance, err
}

// ----------------- Budget Helper -----------------

// checkBudgetWarning evaluates current spending against category limits and returns a warning string if exceeded.
func (s *walletService) checkBudgetWarning(ctx context.Context, userID uint, category string) string {
	budget, err := s.walletRepo.GetBudgetByCategory(ctx, userID, category)
	if err != nil {
		return ""
	}

	wallet, _ := s.walletRepo.GetWalletByUserID(ctx, userID)
	summaries, _ := s.walletRepo.GetMonthlySummary(ctx, wallet.ID)

	var spent int64
	for _, sum := range summaries {
		if sum.Category == category {
			spent = sum.Total
			break
		}
	}

	if spent > budget.MonthlyLimit {
		return "Warning: You've exceeded your " + category + " budget for this month."
	}
	return ""
}

// ----------------- Budget Operations -----------------

// SetBudget creates or updates a monthly spending limit for a specific category.
func (s *walletService) SetBudget(ctx context.Context, userID uint, category string, limit int64) error {
	budget := &models.Budget{
		UserID:       userID,
		Category:     category,
		MonthlyLimit: limit,
	}
	return s.walletRepo.SetBudget(ctx, budget)
}

// GetBudgetStatus retrieves all user budgets alongside current usage statistics.
func (s *walletService) GetBudgetStatus(ctx context.Context, userID uint) ([]models.BudgetStatus, error) {
	budgets, err := s.walletRepo.GetAllBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summaries, _ := s.walletRepo.GetMonthlySummary(ctx, wallet.ID)
	summaryMap := make(map[string]int64)
	for _, sum := range summaries {
		summaryMap[sum.Category] = sum.Total
	}

	var statuses []models.BudgetStatus
	for _, b := range budgets {
		spent := summaryMap[b.Category]
		statuses = append(statuses, models.BudgetStatus{
			Category:     b.Category,
			MonthlyLimit: b.MonthlyLimit,
			SpentSoFar:   spent,
			OverBudget:   spent > b.MonthlyLimit,
		})
	}
	return statuses, nil
}

// ----------------- Updated Withdraw & Transfer -----------------

// Withdraw processes a debit transaction from the wallet and triggers budget checks.
func (s *walletService) Withdraw(ctx context.Context, userID uint, amount int64, category, note string) (int64, string, error) {
	if amount <= 0 {
		return 0, "", errors.New("amount must be greater than zero")
	}

	if err := s.walletRepo.Withdraw(ctx, userID, amount, category, note); err != nil {
		return 0, "", err
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	warning := s.checkBudgetWarning(ctx, userID, category)
	return wallet.Balance, warning, err
}

// Transfer performs an atomic move of funds between wallets.
// We implement a deterministic locking order (sorting by wallet ID)
// to avoid circular wait scenarios which cause database deadlocks.
func (s *walletService) Transfer(ctx context.Context, senderUserID uint, receiverUsername string, amount int64, category, note string) (int64, string, error) {
	if amount <= 0 {
		return 0, "", errors.New("amount must be greater than zero")
	}

	receiverUser, err := s.userRepo.GetUserByUsername(ctx, receiverUsername)
	if err != nil {
		return 0, "", errors.New("receiver not found")
	}
	if senderUserID == receiverUser.ID {
		return 0, "", errors.New("cannot transfer money to yourself")
	}

	if err := s.walletRepo.Transfer(ctx, senderUserID, receiverUser.ID, amount, category, note); err != nil {
		return 0, "", err
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, senderUserID)
	warning := s.checkBudgetWarning(ctx, senderUserID, category)
	return wallet.Balance, warning, err
}
