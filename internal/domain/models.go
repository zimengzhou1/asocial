package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system
type User struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// Room represents a chat room
type Room struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	OwnerID     *uuid.UUID `json:"owner_id,omitempty"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RoomUserSettings represents a user's settings for a specific room
type RoomUserSettings struct {
	RoomID       uuid.UUID `json:"room_id"`
	UserID       uuid.UUID `json:"user_id"`
	DisplayName  string    `json:"display_name"`
	Color        string    `json:"color"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// CreateUserParams contains parameters for creating a new user
type CreateUserParams struct {
	Email    string
	Username string
}

// CreateRoomParams contains parameters for creating a new room
type CreateRoomParams struct {
	Name        string
	Slug        string
	Description *string
	OwnerID     *uuid.UUID
	IsPublic    bool
}

// UpdateRoomUserSettingsParams contains parameters for updating room user settings
type UpdateRoomUserSettingsParams struct {
	RoomID      uuid.UUID
	UserID      uuid.UUID
	DisplayName *string
	Color       *string
}
