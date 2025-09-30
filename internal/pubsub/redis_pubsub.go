package pubsub

import (
	"asocial/internal/domain"
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub handles Redis pub/sub operations
type RedisPubSub struct {
	client  *redis.Client
	channel string
	logger  *slog.Logger
}

// NewRedisPubSub creates a new Redis pub/sub client
func NewRedisPubSub(addr, password, channel string, db int, logger *slog.Logger) (*RedisPubSub, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis", "addr", addr, "db", db)

	return &RedisPubSub{
		client:  client,
		channel: channel,
		logger:  logger,
	}, nil
}

// Publish publishes a message to the Redis channel
func (r *RedisPubSub) Publish(ctx context.Context, msg *domain.Message) error {
	data := msg.Encode()
	if err := r.client.Publish(ctx, r.channel, data).Err(); err != nil {
		r.logger.Error("Failed to publish message", "error", err, "channel", r.channel)
		return fmt.Errorf("%w: %v", domain.ErrPublishFailed, err)
	}

	r.logger.Debug("Published message", "message_id", msg.MessageID, "channel", r.channel)
	return nil
}

// Subscribe subscribes to the Redis channel and processes messages with the provided handler
func (r *RedisPubSub) Subscribe(ctx context.Context, handler func(*domain.Message) error) error {
	pubsub := r.client.Subscribe(ctx, r.channel)
	defer pubsub.Close()

	// Wait for confirmation that subscription is created
	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", r.channel, err)
	}

	r.logger.Info("Subscribed to Redis channel", "channel", r.channel)

	// Get channel for receiving messages
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Subscription cancelled", "channel", r.channel)
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				r.logger.Warn("Redis channel closed")
				return nil
			}

			message, err := domain.DecodeMessage([]byte(msg.Payload))
			if err != nil {
				r.logger.Error("Failed to decode message", "error", err, "payload", msg.Payload)
				continue
			}

			if err := handler(message); err != nil {
				r.logger.Error("Handler failed to process message", "error", err, "message_id", message.MessageID)
				// Continue processing other messages even if handler fails
				continue
			}

			r.logger.Debug("Processed message", "message_id", message.MessageID)
		}
	}
}

// HealthCheck checks if Redis connection is healthy
func (r *RedisPubSub) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis client connection
func (r *RedisPubSub) Close() error {
	return r.client.Close()
}
