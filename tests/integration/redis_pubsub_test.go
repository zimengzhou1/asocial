package integration

import (
	"asocial/internal/domain"
	"asocial/internal/pubsub"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

// TestRedisPubSub tests Redis pub/sub integration
// Requires Redis running on localhost:6379
func TestRedisPubSub(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if Redis is available
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))

	// Create Redis pub/sub client
	redisPubSub, err := pubsub.NewRedisPubSub(redisAddr, "", "test:messages", 0, logger)
	if err != nil {
		t.Skipf("Redis not available at %s: %v", redisAddr, err)
	}
	defer redisPubSub.Close()

	// Test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Channel to receive messages
	received := make(chan *domain.Message, 1)

	// Start subscriber
	go func() {
		err := redisPubSub.Subscribe(ctx, func(msg *domain.Message) error {
			received <- msg
			return nil
		})
		if err != nil && err != context.Canceled {
			t.Errorf("Subscribe error: %v", err)
		}
	}()

	// Give subscriber time to connect
	time.Sleep(100 * time.Millisecond)

	// Publish a test message
	testMsg := domain.NewMessage(
		"test-msg-123",
		"test-channel",
		"test-user",
		"Hello from integration test",
		domain.Position{X: 10, Y: 20},
	)

	err = redisPubSub.Publish(ctx, testMsg)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Wait for message to be received
	select {
	case msg := <-received:
		if msg.MessageID != testMsg.MessageID {
			t.Errorf("Expected MessageID %s, got %s", testMsg.MessageID, msg.MessageID)
		}
		if msg.Payload != testMsg.Payload {
			t.Errorf("Expected Payload %s, got %s", testMsg.Payload, msg.Payload)
		}
		if msg.ChannelID != testMsg.ChannelID {
			t.Errorf("Expected ChannelID %s, got %s", testMsg.ChannelID, msg.ChannelID)
		}
		if msg.UserID != testMsg.UserID {
			t.Errorf("Expected UserID %s, got %s", testMsg.UserID, msg.UserID)
		}
		if msg.Position.X != testMsg.Position.X || msg.Position.Y != testMsg.Position.Y {
			t.Errorf("Expected Position %+v, got %+v", testMsg.Position, msg.Position)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for message")
	}
}

// TestRedisPubSubMultipleMessages tests publishing and receiving multiple messages
func TestRedisPubSubMultipleMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	redisPubSub, err := pubsub.NewRedisPubSub(redisAddr, "", "test:multiple", 0, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisPubSub.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messageCount := 5
	received := make(chan *domain.Message, messageCount)

	// Start subscriber
	go func() {
		err := redisPubSub.Subscribe(ctx, func(msg *domain.Message) error {
			received <- msg
			return nil
		})
		if err != nil && err != context.Canceled {
			t.Errorf("Subscribe error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Publish multiple messages
	for i := 0; i < messageCount; i++ {
		msg := domain.NewMessage(
			string(rune('A'+i)),
			"test-channel",
			"test-user",
			string(rune('A'+i))+" message",
			domain.Position{X: i * 10, Y: i * 20},
		)
		if err := redisPubSub.Publish(ctx, msg); err != nil {
			t.Fatalf("Failed to publish message %d: %v", i, err)
		}
	}

	// Collect all messages
	receivedMsgs := make([]*domain.Message, 0, messageCount)
	for i := 0; i < messageCount; i++ {
		select {
		case msg := <-received:
			receivedMsgs = append(receivedMsgs, msg)
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for message %d (received %d)", i, len(receivedMsgs))
		}
	}

	if len(receivedMsgs) != messageCount {
		t.Errorf("Expected %d messages, got %d", messageCount, len(receivedMsgs))
	}
}

// TestRedisHealthCheck tests the health check functionality
func TestRedisHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	redisPubSub, err := pubsub.NewRedisPubSub(redisAddr, "", "test:health", 0, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisPubSub.Close()

	ctx := context.Background()
	if err := redisPubSub.HealthCheck(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

// TestRedisConnection_InvalidAddress tests connection failure handling
func TestRedisConnection_InvalidAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Try to connect to invalid address
	_, err := pubsub.NewRedisPubSub("invalid:9999", "", "test:invalid", 0, logger)
	if err == nil {
		t.Error("Expected error when connecting to invalid address")
	}
}
