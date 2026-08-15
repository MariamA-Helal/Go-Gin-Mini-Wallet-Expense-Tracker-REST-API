package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupUserRouterForTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	userHandler := NewUserHandler(&mockUserService{})

	r.POST("/api/signup", userHandler.Signup)
	r.POST("/api/login", userHandler.Login)

	return r
}

func TestUserHandler_Signup_Integration(t *testing.T) {
	router := setupUserRouterForTest()

	tests := []struct {
		name         string
		payload      AuthInput
		expectedCode int
	}{
		{
			name:         "Valid Signup",
			payload:      AuthInput{Username: "newuser", Password: "SecurePass123*"},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Duplicate User",
			payload:      AuthInput{Username: "existinguser", Password: "SecurePass123*"},
			expectedCode: http.StatusConflict,
		},
		{
			name:         "Invalid Input (Missing Username)",
			payload:      AuthInput{Username: "", Password: "SecurePass123*"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Signup() Expected status code %d, got %d. Response: %s", tt.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestUserHandler_Login_Integration(t *testing.T) {
	router := setupUserRouterForTest()

	tests := []struct {
		name         string
		payload      AuthInput
		expectedCode int
	}{
		{
			name:         "Valid Login",
			payload:      AuthInput{Username: "validuser", Password: "SecurePass123*"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid Credentials (Wrong Password)",
			payload:      AuthInput{Username: "validuser", Password: "WrongPassword123*"},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Invalid Input (Short Password)",
			payload:      AuthInput{Username: "validuser", Password: "123"},
			expectedCode: http.StatusBadRequest, // Gin binding should fail here
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Login() Expected status code %d, got %d. Response: %s", tt.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}
