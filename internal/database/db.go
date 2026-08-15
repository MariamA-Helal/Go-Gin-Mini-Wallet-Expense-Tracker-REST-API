package database

import (
	"fmt"
	"log"
	"os"
	"time"
	"wallet-api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB initializes the PostgreSQL database connection and creates the DB if it doesn't exist
func ConnectDB() (*gorm.DB, error) {
	// 1. Fetch credentials securely from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Fallback to default name if not provided
	if dbName == "" {
		dbName = "wallet_db"
	}

	// 2. Connect to the default 'postgres' database to check/create our target database
	defaultDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable TimeZone=Africa/Cairo",
		dbHost, dbUser, dbPassword, dbPort)

	var defaultDB *gorm.DB
	var err error
	maxRetries := 3

	// 3. Retry logic: Attempt to connect up to 3 times
	for i := 1; i <= maxRetries; i++ {
		defaultDB, err = gorm.Open(postgres.Open(defaultDSN), &gorm.Config{})
		if err == nil {
			break // Connection successful, exit the loop
		}
		log.Printf("Attempt %d/%d: Failed to connect to default database...", i, maxRetries)
		if i < maxRetries {
			log.Println("Retrying in 2 seconds...")
			time.Sleep(2 * time.Second)
		}
	}

	// If it still fails after 3 retries, return the error
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	// 4. Check if the target database exists
	var count int64
	defaultDB.Raw("SELECT count(*) FROM pg_database WHERE datname = ?", dbName).Scan(&count)

	if count == 0 {
		log.Printf("Database '%s' not found. Creating it automatically...", dbName)
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s;", dbName)
		if err := defaultDB.Exec(createDBQuery).Error; err != nil {
			return nil, fmt.Errorf("failed to create database: %v", err)
		}
		log.Printf("Database '%s' created successfully.", dbName)
	}

	// 5. Close the connection to the default database
	sqlDB, _ := defaultDB.DB()
	sqlDB.Close()

	// 6. Connect to our actual target database (e.g., wallet_db)
	targetDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Cairo",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(targetDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target database: %v", err)
	}

	log.Printf("Successfully connected to PostgreSQL (%s)!", dbName)

	// 7. AutoMigrate tables
	err = db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{}, &models.Budget{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return nil, err
	}

	log.Println("AutoMigration completed successfully!")
	return db, nil
}
