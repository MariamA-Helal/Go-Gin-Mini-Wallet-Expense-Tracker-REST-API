package repository

import (
	"context"
	"sync"
	"testing"
	"wallet-api/internal/models"
)

func TestConcurrentWithdrawal(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)

	// 1. User With Wallet Balance 1000 pounds
	user := models.User{Username: "concurrent_user", Password: "123", Role: "user"}
	db.Create(&user)
	wallet := models.Wallet{UserID: user.ID, Balance: 1000}
	db.Create(&wallet)

	var wg sync.WaitGroup
	successCount := 0
	errCount := 0
	var mu sync.Mutex

	// 2. Launching two simultaneous withdrawals (each one withdrawing 1000)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.Withdraw(context.Background(), user.ID, 1000, "Test", "Concurrent")

			mu.Lock()
			if err == nil {
				successCount++
			} else {
				errCount++
			}
			mu.Unlock()
		}()
	}

	//Wait until both transactions finish
	wg.Wait()

	// 3. Checking the result (one transaction succeeds and the other fails)
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful withdrawal, got %d", successCount)
	}
	if errCount != 1 {
		t.Errorf("Expected exactly 1 failed withdrawal due to insufficient funds, got %d", errCount)
	}

	// 4.Make sure the final balance is zero and not negative
	var finalWallet models.Wallet
	db.First(&finalWallet, wallet.ID)
	if finalWallet.Balance != 0 {
		t.Errorf("Expected final balance to be 0, got %d", finalWallet.Balance)
	}
}
