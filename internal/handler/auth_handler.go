package handler

import (
	"asocial/internal/auth"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	firebaseService *auth.FirebaseService
	logger          *slog.Logger
	appURL          string
	isDev           bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(firebaseService *auth.FirebaseService, logger *slog.Logger, appURL string, isDev bool) *AuthHandler {
	return &AuthHandler{
		firebaseService: firebaseService,
		logger:          logger,
		appURL:          appURL,
		isDev:           isDev,
	}
}

// HandleMe returns the current authenticated user
func (h *AuthHandler) HandleMe(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get full user info
	user, err := h.firebaseService.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("failed to get user", "error", err, "user_id", uid)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

// HandleLogout logs out the current user by revoking Firebase tokens
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	// Get Firebase UID from context (set by auth middleware)
	firebaseUID, exists := c.Get("firebase_uid")
	if exists {
		uid, ok := firebaseUID.(string)
		if ok {
			// Revoke all refresh tokens for this user
			if err := h.firebaseService.RevokeTokens(c.Request.Context(), uid); err != nil {
				h.logger.Error("failed to revoke tokens", "error", err, "firebase_uid", uid)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// HandleCheckUsername checks if a username is available
func (h *AuthHandler) HandleCheckUsername(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	available, err := h.firebaseService.IsUsernameAvailable(c.Request.Context(), req.Username)
	if err != nil {
		h.logger.Error("failed to check username", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
		return
	}

	if available {
		c.JSON(http.StatusOK, gin.H{"available": true})
		return
	}

	// Generate suggestions
	suggestions, err := h.firebaseService.GenerateUsernameSuggestions(c.Request.Context(), req.Username)
	if err != nil {
		h.logger.Error("failed to generate suggestions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available":   false,
		"suggestions": suggestions,
	})
}

// HandleUpdateUsername updates the user's username
func (h *AuthHandler) HandleUpdateUsername(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 3 and 20 characters"})
		return
	}

	// Check if username is available
	available, err := h.firebaseService.IsUsernameAvailable(c.Request.Context(), req.Username)
	if err != nil {
		h.logger.Error("failed to check username availability", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
		return
	}

	if !available {
		// Generate suggestions
		suggestions, err := h.firebaseService.GenerateUsernameSuggestions(c.Request.Context(), req.Username)
		if err != nil {
			h.logger.Error("failed to generate suggestions", "error", err)
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Username is already taken",
				"message": "This username is already taken",
			})
			return
		}

		c.JSON(http.StatusConflict, gin.H{
			"error":       "Username is already taken",
			"message":     "This username is already taken",
			"suggestions": suggestions,
		})
		return
	}

	// Update username
	err = h.firebaseService.UpdateUsername(c.Request.Context(), uid, req.Username)
	if err != nil {
		h.logger.Error("failed to update username", "error", err, "user_id", uid)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update username"})
		return
	}

	// Get updated user
	user, err := h.firebaseService.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("failed to get updated user", "error", err, "user_id", uid)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	h.logger.Info("username updated", "user_id", uid, "new_username", req.Username)

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

// HandleDeleteAccount deletes the user's account
func (h *AuthHandler) HandleDeleteAccount(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get Firebase UID for token revocation
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		h.logger.Warn("firebase_uid not found in context", "user_id", uid)
	}

	// Delete user from database (will cascade to room_user_settings)
	err := h.firebaseService.DeleteUser(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("failed to delete user", "error", err, "user_id", uid)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	// Revoke Firebase tokens if we have the UID
	if firebaseUID != nil {
		fuid, ok := firebaseUID.(string)
		if ok {
			if err := h.firebaseService.RevokeTokens(c.Request.Context(), fuid); err != nil {
				h.logger.Error("failed to revoke tokens", "error", err, "firebase_uid", fuid)
				// Continue anyway - account is already deleted from DB
			}
		}
	}

	h.logger.Info("account deleted", "user_id", uid)

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

