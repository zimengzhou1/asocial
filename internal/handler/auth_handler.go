package handler

import (
	"asocial/internal/auth"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	firebaseService *auth.FirebaseService
	logger          *slog.Logger
	appURL          string
	isDev           bool
	jwtSecret       string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(firebaseService *auth.FirebaseService, logger *slog.Logger, appURL string, isDev bool, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		firebaseService: firebaseService,
		logger:          logger,
		appURL:          appURL,
		isDev:           isDev,
		jwtSecret:       jwtSecret,
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

// AnonymousSessionRequest represents a request for an anonymous session
type AnonymousSessionRequest struct {
	ClientID string `json:"client_id" binding:"required"`
}

// AnonymousSessionResponse contains the session token
type AnonymousSessionResponse struct {
	SessionToken string `json:"session_token"`
	UserID       string `json:"user_id"`
	ExpiresAt    int64  `json:"expires_at"`
}

// HandleAnonymousSession issues a session token for anonymous users
func (h *AuthHandler) HandleAnonymousSession(c *gin.Context) {
	var req AnonymousSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validate client ID format (prevent injection)
	if !isValidClientID(req.ClientID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id format"})
		return
	}

	// Generate server-controlled user ID
	userID := "anon-" + uuid.New().String()[:8]

	// Create JWT with short expiration (24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"client_id":  req.ClientID,
		"session_id": uuid.New().String(),
		"anonymous":  true,
		"exp":        expiresAt.Unix(),
		"iat":        time.Now().Unix(),
	})

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.Error("Failed to sign JWT", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.logger.Info("Created anonymous session", "user_id", userID, "client_id", req.ClientID)

	c.JSON(http.StatusOK, AnonymousSessionResponse{
		SessionToken: tokenString,
		UserID:       userID,
		ExpiresAt:    expiresAt.Unix(),
	})
}

// isValidClientID validates the client ID format (UUID)
func isValidClientID(id string) bool {
	// UUID format (8-4-4-4-12 with hyphens, lowercase hex)
	match, _ := regexp.MatchString(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`, id)
	return match && len(id) == 36
}

