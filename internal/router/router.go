package router

import (
	"wallet-api/internal/handler"
	"wallet-api/internal/middleware"

	"github.com/gin-gonic/gin"

	_ "wallet-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all the API routes
func SetupRouter(userHandler *handler.UserHandler, walletHandler *handler.WalletHandler) *gin.Engine {
	r := gin.Default()

	// Public Routes
	api := r.Group("/api")
	{
		api.POST("/signup", userHandler.Signup)
		api.POST("/login", userHandler.Login)
	}

	// Protected Routes
	wallet := api.Group("/wallet")
	wallet.Use(middleware.RequireAuth())
	{
		wallet.GET("/", walletHandler.GetWallet)
		wallet.GET("/transactions", walletHandler.GetHistory)
		wallet.GET("/transactions/summary", walletHandler.GetSummary)
		wallet.POST("/deposit", walletHandler.Deposit)
		wallet.POST("/withdraw", walletHandler.Withdraw)
		wallet.POST("/transfer", walletHandler.Transfer)

		wallet.POST("/budgets", walletHandler.SetBudget)
		wallet.PUT("/budgets/:category", walletHandler.SetBudget)
		wallet.GET("/budgets/status", walletHandler.GetBudgetStatus)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
