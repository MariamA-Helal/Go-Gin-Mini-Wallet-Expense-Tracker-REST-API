package handler

import (
	"log"
	"net/http"
	"wallet-api/internal/models"
	"wallet-api/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService service.WalletService
}

type BudgetInput struct {
	Category     string `json:"category" binding:"required"`
	MonthlyLimit int64  `json:"monthly_limit" binding:"required,gt=0"`
}

// NewWalletHandler initializes a new WalletHandler with the required service dependency.
func NewWalletHandler(ws service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: ws}
}

type TransactionInput struct {
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Category string `json:"category" binding:"required"`
	Note     string `json:"note"`
}

type TransferInput struct {
	ReceiverUsername string `json:"receiver_username" binding:"required"`
	Amount           int64  `json:"amount" binding:"required,gt=0"`
	Category         string `json:"category" binding:"required"`
	Note             string `json:"note"`
}

type HistoryQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
	Category string `form:"category"`
	From     string `form:"from"`
	To       string `form:"to"`
	Username string `form:"username"`
}

// getTokenData extracts user identity and role from the JWT claims stored in the context.
func getTokenData(c *gin.Context) (uint, string, bool) {
	userIDVal, existsID := c.Get("userID")
	roleVal, existsRole := c.Get("role")

	if !existsID || !existsRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return 0, "", false
	}
	return userIDVal.(uint), roleVal.(string), true
}

// ------------------- READ Operations -------------------

// GetWallet retrieves the current wallet balance for the authenticated user or an admin-requested user.
// GetWallet retrieves the current user's wallet balance.
// @Summary Get wallet balance
// @Description Returns the balance of the authenticated user's wallet (or specific user for admins).
// @Tags Wallet
// @Produce json
// @Security ApiKeyAuth
// @Param username query string false "Username (Admin only)"
// @Success 200 {object} map[string]interface{} "balance"
// @Failure 401 {object} map[string]string "error"
// @Failure 403 {object} map[string]string "error"
// @Router /wallet [get]
func (h *WalletHandler) GetWallet(c *gin.Context) {
	tokenUserID, tokenRole, ok := getTokenData(c)
	if !ok {
		return
	}

	requestedUsername := c.Query("username")

	wallet, err := h.walletService.GetWallet(c.Request.Context(), tokenUserID, tokenRole, requestedUsername)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": wallet.Balance})
}

// GetHistory retrieves a paginated and filtered list of transactions for the authenticated user.
// GetHistory retrieves a paginated and filtered list of transactions.
// @Summary Get transaction history
// @Description Returns transaction history with options for pagination, category, date filters, and username lookup for admins.
// @Tags Wallet
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param category filter query string false "Filter by category"
// @Param from query string false "Filter from date (YYYY-MM-DD)"
// @Param to query string false "Filter to date (YYYY-MM-DD)"
// @Param username query string false "Target username (Admin only)"
// @Success 200 {object} map[string]interface{} "paginated transactions"
// @Failure 400 {object} map[string]string "error"
// @Failure 403 {object} map[string]string "error"
// @Router /wallet/transactions [get]
func (h *WalletHandler) GetHistory(c *gin.Context) {
	tokenUserID, tokenRole, ok := getTokenData(c)
	if !ok {
		return
	}

	var q HistoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	filter := models.TransactionFilter{
		Page:     q.Page,
		Limit:    q.Limit,
		Category: q.Category,
		From:     q.From,
		To:       q.To,
	}

	transactions, err := h.walletService.GetTransactionHistory(c.Request.Context(), tokenUserID, tokenRole, q.Username, filter)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":         q.Page,
		"limit":        q.Limit,
		"filters_used": filter,
		"transactions": transactions,
	})
}

// GetSummary returns a monthly aggregated breakdown of expenses by category.
// GetSummary returns category-wise monthly expenditure aggregates.
// @Summary Get monthly expense summary
// @Description Groups transactions by category for the current month.
// @Tags Wallet
// @Produce json
// @Security ApiKeyAuth
// @Param username query string false "Target username (Admin only)"
// @Success 200 {object} map[string]interface{} "monthly summary"
// @Failure 403 {object} map[string]string "error"
// @Router /wallet/transactions/summary [get]
func (h *WalletHandler) GetSummary(c *gin.Context) {
	tokenUserID, tokenRole, ok := getTokenData(c)
	if !ok {
		return
	}

	requestedUsername := c.Query("username")

	summary, err := h.walletService.GetMonthlySummary(c.Request.Context(), tokenUserID, tokenRole, requestedUsername)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"monthly_summary": summary})
}

// ------------------- WRITE Operations (Locked to token owner) -------------------

// Deposit adds funds to the user's wallet and records the transaction.
// Deposit adds funds to the user's wallet.
// @Summary Deposit money
// @Description Increases the wallet balance and records a deposit transaction.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body TransactionInput true "Deposit details"
// @Success 200 {object} map[string]interface{} "message, balance"
// @Failure 400 {object} map[string]string "error"
// @Router /wallet/deposit [post]
func (h *WalletHandler) Deposit(c *gin.Context) {
	userID, _, ok := getTokenData(c)
	if !ok {
		return
	}

	var input TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBalance, err := h.walletService.Deposit(c.Request.Context(), userID, input.Amount, input.Category, input.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "balance": newBalance})
}

// SetBudget creates or updates a spending limit for a specific transaction category.
// SetBudget creates or updates a monthly spending limit for a category.
// @Summary Set or update category budget
// @Description Sets a monthly budget cap (in cents) for a specific spending category.
// @Tags Budgets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body BudgetInput true "Budget details"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "error"
// @Router /wallet/budgets [post]
// @Router /wallet/budgets/{category} [put]
func (h *WalletHandler) SetBudget(c *gin.Context) {
	userID, _, ok := getTokenData(c)
	if !ok {
		return
	}

	var input BudgetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.walletService.SetBudget(c.Request.Context(), userID, input.Category, input.MonthlyLimit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set budget"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Budget set successfully"})
}

// GetBudgetStatus returns the current spending usage against defined budget limits.
// GetBudgetStatus returns spending usage against defined budget limits.
// @Summary Get budget statuses
// @Description Compares current month's spending against category limits and flags over-budget status.
// @Tags Budgets
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "budgets status list"
// @Failure 500 {object} map[string]string "error"
// @Router /wallet/budgets/status [get]
func (h *WalletHandler) GetBudgetStatus(c *gin.Context) {
	userID, _, ok := getTokenData(c)
	if !ok {
		return
	}

	statuses, err := h.walletService.GetBudgetStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budget status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"budgets": statuses})
}

// Withdraw handles user withdrawal requests with balance validation and budget alerting.
// @Summary Withdraw funds from wallet
// @Description Withdraws a specified amount from the user wallet and checks budget limits.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body TransactionInput true "Withdrawal details"
// @Success 200 {object} map[string]interface{} "message, balance, warning"
// @Failure 400 {object} map[string]string "error"
// @Router /wallet/withdraw [post]
func (h *WalletHandler) Withdraw(c *gin.Context) {
	userID, _, ok := getTokenData(c)
	if !ok {
		return
	}

	var input TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBalance, warning, err := h.walletService.Withdraw(c.Request.Context(), userID, input.Amount, input.Category, input.Note)
	if err != nil {
		log.Printf("Error during withdrawal for user %d: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{"message": "Withdrawal successful", "balance": newBalance}
	if warning != "" {
		response["warning"] = warning
	}
	c.JSON(http.StatusOK, response)
}

// Transfer processes an atomic money transfer between two wallets.
// @Summary Transfer money to another wallet
// @Description Moves money securely between two user wallets using an atomic transaction with deadlock protection and budget checks.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body TransferInput true "Transfer details"
// @Success 200 {object} map[string]interface{} "message, balance, warning"
// @Failure 400 {object} map[string]string "error"
// @Failure 401 {object} map[string]string "error"
// @Router /wallet/transfer [post]
func (h *WalletHandler) Transfer(c *gin.Context) {
	userID, _, ok := getTokenData(c)
	if !ok {
		return
	}

	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBalance, warning, err := h.walletService.Transfer(c.Request.Context(), userID, input.ReceiverUsername, input.Amount, input.Category, input.Note)
	if err != nil {
		log.Printf("Error during transfer from user %d to %s: %v", userID, input.ReceiverUsername, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{"message": "Transfer successful", "balance": newBalance}
	if warning != "" {
		response["warning"] = warning
	}
	c.JSON(http.StatusOK, response)
}
