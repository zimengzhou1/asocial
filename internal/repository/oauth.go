package repository

import (
	"asocial/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OAuthAccountRepository handles OAuth account-related database operations
type OAuthAccountRepository struct {
	db *sql.DB
}

// NewOAuthAccountRepository creates a new OAuth account repository
func NewOAuthAccountRepository(db *sql.DB) *OAuthAccountRepository {
	return &OAuthAccountRepository{db: db}
}

// Create creates a new OAuth account
func (r *OAuthAccountRepository) Create(ctx context.Context, params domain.CreateOAuthAccountParams) (*domain.OAuthAccount, error) {
	account := &domain.OAuthAccount{
		ID:             uuid.New(),
		UserID:         params.UserID,
		Provider:       params.Provider,
		ProviderUserID: params.ProviderUserID,
		AccessToken:    params.AccessToken,
		RefreshToken:   params.RefreshToken,
		TokenExpiresAt: params.TokenExpiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	query := `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, access_token, refresh_token, token_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, provider, provider_user_id, access_token, refresh_token, token_expires_at, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.CreatedAt,
		account.UpdatedAt,
	).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create oauth account: %w", err)
	}

	return account, nil
}

// GetByProviderAndUserID retrieves an OAuth account by provider and provider user ID
func (r *OAuthAccountRepository) GetByProviderAndUserID(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error) {
	account := &domain.OAuthAccount{}

	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2
	`

	err := r.db.QueryRowContext(ctx, query, provider, providerUserID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth account: %w", err)
	}

	return account, nil
}

// GetByUserID retrieves all OAuth accounts for a user
func (r *OAuthAccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.OAuthAccount
	for rows.Next() {
		account := &domain.OAuthAccount{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.AccessToken,
			&account.RefreshToken,
			&account.TokenExpiresAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan oauth account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// UpdateTokens updates the access and refresh tokens for an OAuth account
func (r *OAuthAccountRepository) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken *string, expiresAt *time.Time) error {
	query := `
		UPDATE oauth_accounts
		SET access_token = $1, refresh_token = $2, token_expires_at = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, accessToken, refreshToken, expiresAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update oauth account tokens: %w", err)
	}

	return nil
}
