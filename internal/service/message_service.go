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
	// Generate message ID if not provided
	if msg.MessageID == "" {
		msg.MessageID = uuid.New().String()
	}

	// Publish to Redis
	if err := s.pubsub.Publish(ctx, msg); err != nil {
		s.logger.Error("Failed to publish message", "error", err, "message_id", msg.MessageID)
		return err
	}

	s.logger.Info("Published message", "message_id", msg.MessageID, "user_id", msg.UserID, "channel", msg.ChannelID)
	return nil
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
// It filters out the sender and only sends to users in the same channel
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

		// Only send to users in the same channel, but not to the sender
		channelMatch := channelID == msg.ChannelID
		differentUser := userID != msg.UserID

		return channelMatch && differentUser
	})

	s.logger.Debug("Broadcast message", "message_id", msg.MessageID, "channel", msg.ChannelID)
	return nil
}

// HealthCheck checks if the service dependencies are healthy
func (s *MessageService) HealthCheck(ctx context.Context) error {
	return s.pubsub.HealthCheck(ctx)
}
