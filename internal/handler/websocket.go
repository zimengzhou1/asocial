package handler

import (
	"asocial/internal/domain"
	"asocial/internal/service"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// WebSocketHandler handles WebSocket connections and messages
type WebSocketHandler struct {
	melody  *melody.Melody
	service *service.MessageService
	logger  *slog.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(m *melody.Melody, svc *service.MessageService, logger *slog.Logger) *WebSocketHandler {
	handler := &WebSocketHandler{
		melody:  m,
		service: svc,
		logger:  logger,
	}

	// Register Melody event handlers
	m.HandleConnect(handler.handleConnect)
	m.HandleMessage(handler.handleMessage)
	m.HandleDisconnect(handler.handleDisconnect)

	return handler
}

// HandleUpgrade upgrades HTTP connection to WebSocket
func (h *WebSocketHandler) HandleUpgrade(c *gin.Context) {
	if err := h.melody.HandleRequest(c.Writer, c.Request); err != nil {
		h.logger.Error("Failed to upgrade WebSocket", "error", err, "remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}
}

// handleConnect is called when a new WebSocket connection is established
func (h *WebSocketHandler) handleConnect(sess *melody.Session) {
	userID := sess.Request.URL.Query().Get("uid")
	if userID == "" {
		h.logger.Warn("Connection without user ID", "remote_addr", sess.Request.RemoteAddr)
		sess.Close()
		return
	}

	// Store user ID and channel ID in session
	// For now, all users join the "default" channel
	channelID := "default"
	sess.Set("user_id", userID)
	sess.Set("channel_id", channelID)

	ctx := context.Background()

	// Add user to Redis presence set
	if err := h.service.GetPubSubClient().AddUserToChannel(ctx, channelID, userID); err != nil {
		h.logger.Error("Failed to add user to channel", "error", err, "user_id", userID)
	}

	// Get current list of users in the channel and send to the connecting user
	users, err := h.service.GetPubSubClient().GetChannelUsers(ctx, channelID)
	if err != nil {
		h.logger.Error("Failed to get channel users", "error", err, "user_id", userID)
	} else {
		// Send user list directly to this session (not via pub/sub)
		syncMsg := domain.NewUserSyncMessage(channelID, users)
		sess.Write(syncMsg.Encode())
		h.logger.Debug("Sent user sync", "user_id", userID, "user_count", len(users))
	}

	// Publish user joined event
	joinMsg := domain.NewUserJoinedMessage(channelID, userID)
	if err := h.service.PublishMessage(ctx, joinMsg); err != nil {
		h.logger.Error("Failed to publish join event", "error", err, "user_id", userID)
	}

	// Start heartbeat to keep presence alive
	go h.startHeartbeat(sess, channelID, userID)

	h.logger.Info("WebSocket connected", "user_id", userID, "channel_id", channelID, "remote_addr", sess.Request.RemoteAddr)
}

// handleMessage is called when a message is received from a WebSocket client
func (h *WebSocketHandler) handleMessage(sess *melody.Session, data []byte) {
	// Decode the message
	msg, err := domain.DecodeMessage(data)
	if err != nil {
		h.logger.Error("Failed to decode message", "error", err, "data", string(data))
		return
	}

	// Get user ID from session
	userID, exists := sess.Get("user_id")
	if !exists {
		h.logger.Warn("Message from session without user ID")
		return
	}

	// Validate that the message's user ID matches the session
	if msg.UserID != userID {
		h.logger.Warn("Message user ID mismatch", "session_user_id", userID, "message_user_id", msg.UserID)
		return
	}

	payloadLen := 0
	if msg.Payload != nil {
		payloadLen = len(*msg.Payload)
	}

	h.logger.Info("Received message",
		"message_id", msg.MessageID,
		"user_id", msg.UserID,
		"channel_id", msg.ChannelID,
		"payload_length", payloadLen,
	)

	// Publish the message via the service
	ctx := context.Background()
	if err := h.service.PublishMessage(ctx, msg); err != nil {
		h.logger.Error("Failed to publish message", "error", err, "message_id", msg.MessageID)
		return
	}
}

// handleDisconnect is called when a WebSocket connection is closed
func (h *WebSocketHandler) handleDisconnect(sess *melody.Session) {
	userIDVal, _ := sess.Get("user_id")
	channelIDVal, _ := sess.Get("channel_id")

	userID, _ := userIDVal.(string)
	channelID, _ := channelIDVal.(string)

	if userID != "" && channelID != "" {
		ctx := context.Background()

		// Remove user from Redis presence set
		if err := h.service.GetPubSubClient().RemoveUserFromChannel(ctx, channelID, userID); err != nil {
			h.logger.Error("Failed to remove user from channel", "error", err, "user_id", userID)
		}

		// Publish user left event
		leaveMsg := domain.NewUserLeftMessage(channelID, userID)
		if err := h.service.PublishMessage(ctx, leaveMsg); err != nil {
			h.logger.Error("Failed to publish leave event", "error", err, "user_id", userID)
		}
	}

	h.logger.Info("WebSocket disconnected",
		"user_id", userID,
		"channel_id", channelID,
		"remote_addr", sess.Request.RemoteAddr,
	)
}

// startHeartbeat periodically refreshes user presence in Redis
func (h *WebSocketHandler) startHeartbeat(sess *melody.Session, channelID, userID string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Check if session is still active
		if sess.IsClosed() {
			h.logger.Debug("Session closed, stopping heartbeat", "user_id", userID)
			return
		}

		// Refresh presence TTL
		ctx := context.Background()
		if err := h.service.GetPubSubClient().RefreshUserPresence(ctx, channelID, userID); err != nil {
			h.logger.Error("Failed to refresh user presence", "error", err, "user_id", userID)
			return
		}

		h.logger.Debug("Refreshed user presence", "user_id", userID, "channel_id", channelID)
	}
}
