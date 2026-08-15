package handler

import (
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

// HistoryQuery defines the parameters, now using 'username' for admin lookups
type HistoryQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
	Category string `form:"category"`
	From     string `form:"from"`
	To       string `form:"to"`
	Username string `form:"username"`
}

// Helper to get token data securely
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

// ----------------- Updated Withdraw Handler -----------------
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{"message": "Withdraw successful", "balance": newBalance}
	if warning != "" {
		response["warning"] = warning
	}
	c.JSON(http.StatusOK, response)
}

// ----------------- Updated Transfer Handler -----------------
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{"message": "Transfer successful", "balance": newBalance}
	if warning != "" {
		response["warning"] = warning
	}
	c.JSON(http.StatusOK, response)
}
