package auth

import (
	"asocial/internal/domain"
	"asocial/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUsernameConflict = errors.New("username already taken")
)

// FirebaseService handles authentication operations using Firebase
type FirebaseService struct {
	firebaseClient *auth.Client
	userRepo       *repository.UserRepository
	logger         *slog.Logger
}

// NewFirebaseService creates a new Firebase auth service
func NewFirebaseService(
	firebaseClient *auth.Client,
	userRepo *repository.UserRepository,
	logger *slog.Logger,
) *FirebaseService {
	return &FirebaseService{
		firebaseClient: firebaseClient,
		userRepo:       userRepo,
		logger:         logger,
	}
}

// VerifyToken verifies a Firebase ID token
func (s *FirebaseService) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := s.firebaseClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	return token, nil
}

// GetOrCreateUser gets a user by email, creates if not exists
// This is called by middleware on every authenticated request for auto-sync
func (s *FirebaseService) GetOrCreateUser(ctx context.Context, email string) (*domain.User, error) {
	// Try to get existing user
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// User exists - update last seen
		if err := s.userRepo.UpdateLastSeenAt(ctx, existingUser.ID); err != nil {
			s.logger.Warn("failed to update last seen", "error", err)
		}
		return existingUser, nil
	}

	// New user - create in database
	username := s.generateUsernameFromEmail(email)

	// Ensure username is unique
	available, err := s.IsUsernameAvailable(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	if !available {
		username, err = s.generateAvailableUsername(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("failed to generate username: %w", err)
		}
	}

	// Create user
	userParams := domain.CreateUserParams{
		Email:    email,
		Username: username,
	}

	user, err := s.userRepo.Create(ctx, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("new user created from Firebase auth",
		"user_id", user.ID,
		"email", user.Email,
		"username", user.Username,
	)

	return user, nil
}

// RevokeTokens revokes all refresh tokens for a user (logout)
func (s *FirebaseService) RevokeTokens(ctx context.Context, firebaseUID string) error {
	if err := s.firebaseClient.RevokeRefreshTokens(ctx, firebaseUID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	s.logger.Info("tokens revoked", "firebase_uid", firebaseUID)
	return nil
}

// GetUserByID retrieves a user by ID
func (s *FirebaseService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *FirebaseService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// IsUsernameAvailable checks if a username is available
func (s *FirebaseService) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return user == nil, nil
}

// GenerateUsernameSuggestions generates alternative username suggestions
func (s *FirebaseService) GenerateUsernameSuggestions(ctx context.Context, baseUsername string) ([]string, error) {
	suggestions := []string{}
	currentYear := time.Now().Year()

	// Try appending numbers
	for i := 2; i <= 5; i++ {
		candidate := fmt.Sprintf("%s%d", baseUsername, i)
		available, err := s.IsUsernameAvailable(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if available {
			suggestions = append(suggestions, candidate)
			if len(suggestions) >= 3 {
				break
			}
		}
	}

	// Try appending year
	if len(suggestions) < 3 {
		candidate := fmt.Sprintf("%s_%d", baseUsername, currentYear)
		available, err := s.IsUsernameAvailable(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if available {
			suggestions = append(suggestions, candidate)
		}
	}

	// Try appending random suffix
	if len(suggestions) < 3 {
		candidate := fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:8])
		suggestions = append(suggestions, candidate)
	}

	return suggestions, nil
}

// Helper functions

// generateUsernameFromEmail generates a username from an email address
func (s *FirebaseService) generateUsernameFromEmail(email string) string {
	// Extract part before @ symbol
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		username := parts[0]
		// Clean up special characters
		username = strings.ReplaceAll(username, ".", "_")
		username = strings.ReplaceAll(username, "+", "_")
		return username
	}
	return "user_" + uuid.New().String()[:8]
}

// generateAvailableUsername generates an available username automatically
func (s *FirebaseService) generateAvailableUsername(ctx context.Context, baseUsername string) (string, error) {
	// Try numbered suffix
	for i := 1; i <= 100; i++ {
		candidate := fmt.Sprintf("%s%d", baseUsername, i)
		available, err := s.IsUsernameAvailable(ctx, candidate)
		if err != nil {
			return "", err
		}
		if available {
			return candidate, nil
		}
	}

	// Fallback to UUID
	return fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:8]), nil
}
