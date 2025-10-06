package service

import (
	"asocial/internal/domain"
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)

// MessageService handles message business logic
type MessageService struct {
	pubsub PubSubClient
	melody *melody.Melody
	logger *slog.Logger
}

// PubSubClient is an interface for pub/sub operations
type PubSubClient interface {
	Publish(ctx context.Context, msg *domain.Message) error
	Subscribe(ctx context.Context, handler func(*domain.Message) error) error
	HealthCheck(ctx context.Context) error
	// Presence operations
	AddUserToChannel(ctx context.Context, channelID, userID string, username, color *string) error
	RemoveUserFromChannel(ctx context.Context, channelID, userID string) error
	RefreshUserPresence(ctx context.Context, channelID, userID string) error
	GetChannelUsers(ctx context.Context, channelID string) ([]domain.UserInfo, error)
}

// NewMessageService creates a new message service
func NewMessageService(pubsub PubSubClient, m *melody.Melody, logger *slog.Logger) *MessageService {
	return &MessageService{
		pubsub: pubsub,
		melody: m,
		logger: logger,
	}
}

// PublishMessage publishes a message to the pub/sub system
func (s *MessageService) PublishMessage(ctx context.Context, msg *domain.Message) error {
	// Generate message ID if not provided (for chat messages only)
	if msg.Type == domain.MessageTypeChat && (msg.MessageID == nil || *msg.MessageID == "") {
		id := uuid.New().String()
		msg.MessageID = &id
	}

	// Publish to Redis
	if err := s.pubsub.Publish(ctx, msg); err != nil {
		s.logger.Error("Failed to publish message", "error", err, "message_id", msg.MessageID)
		return err
	}

	s.logger.Info("Published message", "message_id", msg.MessageID, "user_id", msg.UserID, "channel", msg.ChannelID)
	return nil
}

// GetPubSubClient returns the underlying PubSubClient for presence operations
func (s *MessageService) GetPubSubClient() PubSubClient {
	return s.pubsub
}

// StartSubscriber starts listening for messages and broadcasts them via WebSocket
func (s *MessageService) StartSubscriber(ctx context.Context) error {
	s.logger.Info("Starting message subscriber")

	return s.pubsub.Subscribe(ctx, func(msg *domain.Message) error {
		// Broadcast to all WebSocket connections except the sender
		return s.broadcastMessage(msg)
	})
}

// broadcastMessage broadcasts a message to WebSocket clients
// For chat messages: filters out the sender, sends only to users in the same channel
// For presence events: sends to all users in the channel (including sender)
func (s *MessageService) broadcastMessage(msg *domain.Message) error {
	data := msg.Encode()

	s.melody.BroadcastFilter(data, func(sess *melody.Session) bool {
		// Get session channel ID
		channelID, exists := sess.Get("channel_id")
		if !exists {
			return false
		}

		// Get session user ID
		userID, exists := sess.Get("user_id")
		if !exists {
			return false
		}

		// Only send to users in the same channel
		channelMatch := channelID == msg.ChannelID
		if !channelMatch {
			return false
		}

		// For presence events (join/leave/username_changed/color_changed), send to everyone including sender
		if msg.Type == domain.MessageTypeUserJoined || msg.Type == domain.MessageTypeUserLeft ||
			msg.Type == domain.MessageTypeUsernameChanged || msg.Type == domain.MessageTypeColorChanged {
			return true
		}

		// For chat messages, don't send to sender
		differentUser := userID != msg.UserID
		return differentUser
	})

	s.logger.Debug("Broadcast message", "type", msg.Type, "message_id", msg.MessageID, "channel", msg.ChannelID)
	return nil
}

// HealthCheck checks if the service dependencies are healthy
func (s *MessageService) HealthCheck(ctx context.Context) error {
	return s.pubsub.HealthCheck(ctx)
}
