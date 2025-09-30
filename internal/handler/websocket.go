package handler

import (
	"asocial/internal/domain"
	"asocial/internal/service"
	"context"
	"log/slog"
	"net/http"

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

	h.logger.Info("Received message",
		"message_id", msg.MessageID,
		"user_id", msg.UserID,
		"channel_id", msg.ChannelID,
		"payload_length", len(msg.Payload),
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
	userID, _ := sess.Get("user_id")
	channelID, _ := sess.Get("channel_id")

	h.logger.Info("WebSocket disconnected",
		"user_id", userID,
		"channel_id", channelID,
		"remote_addr", sess.Request.RemoteAddr,
	)
}
