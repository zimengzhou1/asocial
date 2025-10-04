package repository

import (
	"asocial/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RoomUserSettingsRepository handles room user settings-related database operations
type RoomUserSettingsRepository struct {
	db *sql.DB
}

// NewRoomUserSettingsRepository creates a new room user settings repository
func NewRoomUserSettingsRepository(db *sql.DB) *RoomUserSettingsRepository {
	return &RoomUserSettingsRepository{db: db}
}

// Upsert creates or updates room user settings
func (r *RoomUserSettingsRepository) Upsert(ctx context.Context, roomID, userID uuid.UUID, displayName, color string) (*domain.RoomUserSettings, error) {
	settings := &domain.RoomUserSettings{
		RoomID:       roomID,
		UserID:       userID,
		DisplayName:  displayName,
		Color:        color,
		JoinedAt:     time.Now(),
		LastActiveAt: time.Now(),
	}

	query := `
		INSERT INTO room_user_settings (room_id, user_id, display_name, color, joined_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (room_id, user_id)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			color = EXCLUDED.color,
			last_active_at = EXCLUDED.last_active_at
		RETURNING room_id, user_id, display_name, color, joined_at, last_active_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		settings.RoomID,
		settings.UserID,
		settings.DisplayName,
		settings.Color,
		settings.JoinedAt,
		settings.LastActiveAt,
	).Scan(
		&settings.RoomID,
		&settings.UserID,
		&settings.DisplayName,
		&settings.Color,
		&settings.JoinedAt,
		&settings.LastActiveAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert room user settings: %w", err)
	}

	return settings, nil
}

// Get retrieves room user settings
func (r *RoomUserSettingsRepository) Get(ctx context.Context, roomID, userID uuid.UUID) (*domain.RoomUserSettings, error) {
	settings := &domain.RoomUserSettings{}

	query := `
		SELECT room_id, user_id, display_name, color, joined_at, last_active_at
		FROM room_user_settings
		WHERE room_id = $1 AND user_id = $2
	`

	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(
		&settings.RoomID,
		&settings.UserID,
		&settings.DisplayName,
		&settings.Color,
		&settings.JoinedAt,
		&settings.LastActiveAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room user settings: %w", err)
	}

	return settings, nil
}

// Update updates specific fields of room user settings
func (r *RoomUserSettingsRepository) Update(ctx context.Context, params domain.UpdateRoomUserSettingsParams) error {
	query := `
		UPDATE room_user_settings
		SET
			display_name = COALESCE($1, display_name),
			color = COALESCE($2, color),
			last_active_at = $3
		WHERE room_id = $4 AND user_id = $5
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		params.DisplayName,
		params.Color,
		time.Now(),
		params.RoomID,
		params.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update room user settings: %w", err)
	}

	return nil
}

// UpdateLastActive updates the last active timestamp
func (r *RoomUserSettingsRepository) UpdateLastActive(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `
		UPDATE room_user_settings
		SET last_active_at = $1
		WHERE room_id = $2 AND user_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	return nil
}

// ListByRoom retrieves all user settings for a room
func (r *RoomUserSettingsRepository) ListByRoom(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomUserSettings, error) {
	query := `
		SELECT room_id, user_id, display_name, color, joined_at, last_active_at
		FROM room_user_settings
		WHERE room_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to list room user settings: %w", err)
	}
	defer rows.Close()

	var settingsList []*domain.RoomUserSettings
	for rows.Next() {
		settings := &domain.RoomUserSettings{}
		err := rows.Scan(
			&settings.RoomID,
			&settings.UserID,
			&settings.DisplayName,
			&settings.Color,
			&settings.JoinedAt,
			&settings.LastActiveAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room user settings: %w", err)
		}
		settingsList = append(settingsList, settings)
	}

	return settingsList, nil
}

// ListByUser retrieves all room settings for a user
func (r *RoomUserSettingsRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.RoomUserSettings, error) {
	query := `
		SELECT room_id, user_id, display_name, color, joined_at, last_active_at
		FROM room_user_settings
		WHERE user_id = $1
		ORDER BY last_active_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user room settings: %w", err)
	}
	defer rows.Close()

	var settingsList []*domain.RoomUserSettings
	for rows.Next() {
		settings := &domain.RoomUserSettings{}
		err := rows.Scan(
			&settings.RoomID,
			&settings.UserID,
			&settings.DisplayName,
			&settings.Color,
			&settings.JoinedAt,
			&settings.LastActiveAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room user settings: %w", err)
		}
		settingsList = append(settingsList, settings)
	}

	return settingsList, nil
}

// CheckDisplayNameExists checks if a display name is already taken in a room
func (r *RoomUserSettingsRepository) CheckDisplayNameExists(ctx context.Context, roomID uuid.UUID, displayName string, excludeUserID *uuid.UUID) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM room_user_settings
			WHERE room_id = $1 AND display_name = $2 AND ($3::uuid IS NULL OR user_id != $3)
		)
	`

	err := r.db.QueryRowContext(ctx, query, roomID, displayName, excludeUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check display name exists: %w", err)
	}

	return exists, nil
}

// Delete removes room user settings
func (r *RoomUserSettingsRepository) Delete(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `DELETE FROM room_user_settings WHERE room_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete room user settings: %w", err)
	}

	return nil
}
