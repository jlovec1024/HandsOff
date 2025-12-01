package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/jwt"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	db     *gorm.DB
	cfg    *config.Config
	log    *logger.Logger
	jwtGen *jwt.Generator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		cfg:    cfg,
		log:    log,
		jwtGen: jwt.NewGenerator(cfg.Security.JWTSecret, cfg.Security.JWTExpiry),
	}
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find user by username
	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		h.log.Error("Failed to query user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is disabled"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token, err := h.jwtGen.Generate(user.ID, user.Username)
	if err != nil {
		h.log.Error("Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.log.Info("User logged in", "user_id", user.ID, "username", user.Username)

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  &user,
	})
}

// Logout handles user logout (stateless, just for logging)
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	h.log.Info("User logged out", "user_id", userID, "username", username)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentUser returns current authenticated user info
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.log.Error("Failed to query user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}
