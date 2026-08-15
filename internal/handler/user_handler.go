package handler

import (
	"net/http"
	"wallet-api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// SignupInput defines the expected JSON body for signup and login
type AuthInput struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

// Signup handles user registration and wallet auto-generation
// @Summary Register a new user
// @Description Create a new account with a username and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User registration details"
// @Success 201 {object} map[string]string "message: User created"
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 409 {object} map[string]string "error: User already exists"
// @Router /signup [post]
func (h *UserHandler) Signup(c *gin.Context) {
	var input AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.Signup(c.Request.Context(), input.Username, input.Password)
	if err != nil {
		if err.Error() == "user already exists" { // Depends on DB error handling, but good for UX
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully with an empty wallet"})
}

// Login handles user authentication and returns a JWT
// Login authenticates user credentials and returns a JWT access token.
// @Summary User login
// @Description Authenticates username and password, returning a signed JWT Bearer token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "Login credentials"
// @Success 200 {object} map[string]string "token"
// @Failure 400 {object} map[string]string "error"
// @Failure 401 {object} map[string]string "error"
// @Router /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var input AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}
