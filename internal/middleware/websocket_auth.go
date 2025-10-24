package middleware

import (
	"asocial/internal/auth"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// WebSocketAuthMiddleware validates tokens for WebSocket connections
// Supports both Firebase tokens (authenticated) and anonymous session tokens (JWT)
// This middleware extracts user_id from the token and adds it to the request context
func WebSocketAuthMiddleware(firebaseService *auth.FirebaseService, jwtSecret string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from query parameter
			tokenString := r.URL.Query().Get("token")
			if tokenString == "" {
				logger.Warn("WebSocket connection without token", "ip", r.RemoteAddr)
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			// Try Firebase token first (authenticated users)
			firebaseToken, err := firebaseService.VerifyToken(r.Context(), tokenString)
			if err == nil {
				// Valid Firebase token - get or create backend user
				email := firebaseToken.Claims["email"].(string)
				backendUser, err := firebaseService.GetOrCreateUser(r.Context(), email)
				if err != nil {
					logger.Error("Failed to get/create user for WebSocket", "error", err, "email", email)
					http.Error(w, "failed to get user", http.StatusInternalServerError)
					return
				}

				// Use backend database UUID as user_id
				userID := backendUser.ID.String()
				ctx := context.WithValue(r.Context(), "user_id", userID)
				ctx = context.WithValue(ctx, "anonymous", false)
				ctx = context.WithValue(ctx, "firebase_uid", firebaseToken.UID)

				logger.Debug("Authenticated user via Firebase", "backend_user_id", userID, "firebase_uid", firebaseToken.UID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Not Firebase - try anonymous session token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err == nil && token.Valid {
				// Valid anonymous token
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					logger.Warn("Invalid token claims format", "ip", r.RemoteAddr)
					http.Error(w, "invalid token claims", http.StatusUnauthorized)
					return
				}

				userID, ok := claims["user_id"].(string)
				if !ok || userID == "" {
					logger.Warn("Missing user_id in token", "ip", r.RemoteAddr)
					http.Error(w, "invalid user_id in token", http.StatusUnauthorized)
					return
				}

				// Add to context
				ctx := context.WithValue(r.Context(), "user_id", userID)
				ctx = context.WithValue(ctx, "anonymous", true)

				logger.Debug("Anonymous user authenticated", "user_id", userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Neither Firebase nor valid anonymous token
			logger.Warn("Invalid token", "error", err, "ip", r.RemoteAddr)
			http.Error(w, "invalid token", http.StatusUnauthorized)
		})
	}
}
