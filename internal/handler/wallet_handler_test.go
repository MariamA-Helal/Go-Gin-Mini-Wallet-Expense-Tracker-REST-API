package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouterForTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Mocking the Context with a fake Token ID to bypass middleware for unit testing handlers
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("role", "user")
		c.Next()
	})

	walletHandler := NewWalletHandler(&mockWalletService{})
	r.POST("/api/wallet/transfer", walletHandler.Transfer)
	return r
}

func TestWalletHandler_Transfer_Integration(t *testing.T) {
	router := setupRouterForTest()

	tests := []struct {
		name         string
		payload      TransferInput
		expectedCode int
	}{
		{
			name:         "Happy Path",
			payload:      TransferInput{ReceiverUsername: "valid_user", Amount: 2000, Category: "Gift"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Insufficient Funds",
			payload:      TransferInput{ReceiverUsername: "valid_user", Amount: 20000, Category: "Gift"}, // > 10000 returns error in our mock
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Non-Existent User",
			payload:      TransferInput{ReceiverUsername: "ghost", Amount: 1000, Category: "Gift"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/api/wallet/transfer", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d. Response: %s", tt.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}
