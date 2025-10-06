package handler

import (
	"asocial/internal/repository"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoomHandler handles room-related HTTP requests
type RoomHandler struct {
	roomRepo     *repository.RoomRepository
	settingsRepo *repository.RoomUserSettingsRepository
	userRepo     *repository.UserRepository
	logger       *slog.Logger
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(
	roomRepo *repository.RoomRepository,
	settingsRepo *repository.RoomUserSettingsRepository,
	userRepo *repository.UserRepository,
	logger *slog.Logger,
) *RoomHandler {
	return &RoomHandler{
		roomRepo:     roomRepo,
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// JoinRoomRequest represents the request body for joining a room
type JoinRoomRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Color       *string `json:"color,omitempty"`
}

// JoinRoomResponse represents the response for joining a room
type JoinRoomResponse struct {
	Room struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Description *string `json:"description,omitempty"`
		IsPublic    bool    `json:"is_public"`
	} `json:"room"`
	Settings struct {
		DisplayName string `json:"display_name"`
		Color       string `json:"color"`
		JoinedAt    string `json:"joined_at"`
	} `json:"settings"`
}

// HandleJoinRoom handles joining a room
func (h *RoomHandler) HandleJoinRoom(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get room slug from URL
	roomSlug := c.Param("slug")
	if roomSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room slug is required"})
		return
	}

	// Get room by slug
	room, err := h.roomRepo.GetBySlug(c.Request.Context(), roomSlug)
	if err != nil {
		h.logger.Error("failed to get room", "error", err, "slug", roomSlug)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room"})
		return
	}

	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Check if user has access to this room
	if !room.IsPublic {
		// For private rooms, check if user is the owner
		if room.OwnerID == nil || *room.OwnerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this room"})
			return
		}
	}

	// Parse request body
	var req JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Request body is optional
		req = JoinRoomRequest{}
	}

	// Get user's global username as default display name
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Determine display name and color
	displayName := user.Username
	if req.DisplayName != nil && *req.DisplayName != "" {
		displayName = *req.DisplayName
	}

	color := "#ef4444" // default color
	if req.Color != nil && *req.Color != "" {
		color = *req.Color
	}

	// Check if display name is taken in this room
	exists, err = h.settingsRepo.CheckDisplayNameExists(c.Request.Context(), room.ID, displayName, &userID)
	if err != nil {
		h.logger.Error("failed to check display name", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check display name"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Display name is already taken in this room",
			"field":   "display_name",
			"message": "Please choose a different display name",
		})
		return
	}

	// Create or update room user settings
	settings, err := h.settingsRepo.Upsert(c.Request.Context(), room.ID, userID, displayName, color)
	if err != nil {
		h.logger.Error("failed to upsert room user settings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join room"})
		return
	}

	// Build response
	response := JoinRoomResponse{}
	response.Room.ID = room.ID.String()
	response.Room.Name = room.Name
	response.Room.Slug = room.Slug
	response.Room.Description = room.Description
	response.Room.IsPublic = room.IsPublic
	response.Settings.DisplayName = settings.DisplayName
	response.Settings.Color = settings.Color
	response.Settings.JoinedAt = settings.JoinedAt.Format("2006-01-02T15:04:05Z")

	h.logger.Info("user joined room",
		"user_id", userID,
		"room_id", room.ID,
		"display_name", settings.DisplayName,
	)

	c.JSON(http.StatusOK, response)
}

// HandleGetRoom retrieves room information
func (h *RoomHandler) HandleGetRoom(c *gin.Context) {
	roomSlug := c.Param("slug")
	if roomSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room slug is required"})
		return
	}

	room, err := h.roomRepo.GetBySlug(c.Request.Context(), roomSlug)
	if err != nil {
		h.logger.Error("failed to get room", "error", err, "slug", roomSlug)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room"})
		return
	}

	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// For private rooms, check if user has access
	if !room.IsPublic {
		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, ok := userIDStr.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		if room.OwnerID == nil || *room.OwnerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this room"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          room.ID,
		"name":        room.Name,
		"slug":        room.Slug,
		"description": room.Description,
		"is_public":   room.IsPublic,
		"created_at":  room.CreatedAt,
	})
}

// HandleListPublicRooms lists all public rooms
func (h *RoomHandler) HandleListPublicRooms(c *gin.Context) {
	// TODO: Add pagination parameters
	limit := 50
	offset := 0

	rooms, err := h.roomRepo.ListPublicRooms(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to list public rooms", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list rooms"})
		return
	}

	// Convert to response format
	roomList := make([]gin.H, 0, len(rooms))
	for _, room := range rooms {
		roomList = append(roomList, gin.H{
			"id":          room.ID,
			"name":        room.Name,
			"slug":        room.Slug,
			"description": room.Description,
			"is_public":   room.IsPublic,
			"created_at":  room.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": roomList,
		"count": len(roomList),
	})
}
