package repository

import (
	"asocial/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RoomRepository handles room-related database operations
type RoomRepository struct {
	db *sql.DB
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create creates a new room
func (r *RoomRepository) Create(ctx context.Context, params domain.CreateRoomParams) (*domain.Room, error) {
	room := &domain.Room{
		ID:          uuid.New(),
		Name:        params.Name,
		Slug:        params.Slug,
		Description: params.Description,
		OwnerID:     params.OwnerID,
		IsPublic:    params.IsPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO rooms (id, name, slug, description, owner_id, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, slug, description, owner_id, is_public, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		room.ID,
		room.Name,
		room.Slug,
		room.Description,
		room.OwnerID,
		room.IsPublic,
		room.CreatedAt,
		room.UpdatedAt,
	).Scan(
		&room.ID,
		&room.Name,
		&room.Slug,
		&room.Description,
		&room.OwnerID,
		&room.IsPublic,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return room, nil
}

// GetByID retrieves a room by ID
func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	room := &domain.Room{}

	query := `
		SELECT id, name, slug, description, owner_id, is_public, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Slug,
		&room.Description,
		&room.OwnerID,
		&room.IsPublic,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room by id: %w", err)
	}

	return room, nil
}

// GetBySlug retrieves a room by slug
func (r *RoomRepository) GetBySlug(ctx context.Context, slug string) (*domain.Room, error) {
	room := &domain.Room{}

	query := `
		SELECT id, name, slug, description, owner_id, is_public, created_at, updated_at
		FROM rooms
		WHERE slug = $1
	`

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&room.ID,
		&room.Name,
		&room.Slug,
		&room.Description,
		&room.OwnerID,
		&room.IsPublic,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room by slug: %w", err)
	}

	return room, nil
}

// ListPublicRooms retrieves all public rooms
func (r *RoomRepository) ListPublicRooms(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	query := `
		SELECT id, name, slug, description, owner_id, is_public, created_at, updated_at
		FROM rooms
		WHERE is_public = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list public rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Slug,
			&room.Description,
			&room.OwnerID,
			&room.IsPublic,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// ListUserRooms retrieves all rooms owned by a user
func (r *RoomRepository) ListUserRooms(ctx context.Context, userID uuid.UUID) ([]*domain.Room, error) {
	query := `
		SELECT id, name, slug, description, owner_id, is_public, created_at, updated_at
		FROM rooms
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Slug,
			&room.Description,
			&room.OwnerID,
			&room.IsPublic,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// Update updates a room
func (r *RoomRepository) Update(ctx context.Context, id uuid.UUID, name, description *string, isPublic *bool) error {
	query := `
		UPDATE rooms
		SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			is_public = COALESCE($3, is_public),
			updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, name, description, isPublic, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}

	return nil
}

// Delete deletes a room
func (r *RoomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM rooms WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
}
