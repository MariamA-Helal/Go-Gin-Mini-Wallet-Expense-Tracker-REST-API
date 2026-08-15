package main

import (
	"log"
	"wallet-api/internal/database"
	"wallet-api/internal/handler"
	"wallet-api/internal/repository"
	"wallet-api/internal/router"
	"wallet-api/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables")
	}

	// 2. Initialize Database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	// 4. Initialize Services
	userService := service.NewUserService(userRepo)
	walletService := service.NewWalletService(walletRepo, userRepo)

	// 5. Initialize Handlers
	userHandler := handler.NewUserHandler(userService)
	walletHandler := handler.NewWalletHandler(walletService)

	// 6. Setup Router and Start Server
	r := router.SetupRouter(userHandler, walletHandler)

	log.Println("Starting server on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
