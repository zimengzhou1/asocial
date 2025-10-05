package middleware

import (
	"asocial/internal/auth"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a middleware that requires authentication
func AuthMiddleware(firebaseService *auth.FirebaseService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Debug("no authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Debug("invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		idToken := parts[1]

		// Verify Firebase ID token
		token, err := firebaseService.VerifyToken(c.Request.Context(), idToken)
		if err != nil {
			logger.Debug("invalid Firebase token", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get user from database by email (auto-creates if not exists)
		email, ok := token.Claims["email"].(string)
		if !ok || email == "" {
			logger.Debug("no email in Firebase token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Get or create user in PostgreSQL
		user, err := firebaseService.GetOrCreateUser(c.Request.Context(), email)
		if err != nil {
			logger.Error("failed to get/create user", "email", email, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("email", user.Email)
		c.Set("username", user.Username)
		c.Set("firebase_uid", token.UID)

		c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that optionally authenticates
// If a valid token is present, it sets user info in context
// If no token or invalid token, it continues without setting user info
func OptionalAuthMiddleware(firebaseService *auth.FirebaseService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token, continue without auth
			c.Next()
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without auth
			c.Next()
			return
		}

		idToken := parts[1]

		// Verify Firebase ID token
		token, err := firebaseService.VerifyToken(c.Request.Context(), idToken)
		if err != nil {
			// Invalid token, continue without auth
			logger.Debug("invalid Firebase token in optional auth", "error", err)
			c.Next()
			return
		}

		// Get user from database by email (auto-creates if not exists)
		email, ok := token.Claims["email"].(string)
		if !ok || email == "" {
			// No email, continue without auth
			c.Next()
			return
		}

		// Get or create user in PostgreSQL
		user, err := firebaseService.GetOrCreateUser(c.Request.Context(), email)
		if err != nil {
			// Failed to get/create user, continue without auth
			logger.Debug("failed to get/create user in optional auth", "email", email, "error", err)
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("email", user.Email)
		c.Set("username", user.Username)
		c.Set("firebase_uid", token.UID)

		c.Next()
	}
}
