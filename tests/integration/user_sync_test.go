package integration

import (
	"asocial/internal/domain"
	"asocial/internal/handler"
	"asocial/internal/pubsub"
	"asocial/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

// TestUserSync_TwoUsers tests the full E2E flow of two users connecting and syncing
func TestUserSync_TwoUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Setup Redis PubSub
	redisPubSub, err := pubsub.NewRedisPubSub(redisAddr, "", "test:e2e", 0, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisPubSub.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Note: The websocket handler hardcodes channel to "default"
	channelID := "default"

	// Clean up before and after test - also clean up any stale users
	cleanup := func() {
		users, _ := redisPubSub.GetChannelUsers(ctx, channelID)
		for _, user := range users {
			redisPubSub.RemoveUserFromChannel(ctx, channelID, user.UserID)
		}
	}
	cleanup()
	defer cleanup()

	// Setup server
	m := melody.New()
	msgService := service.NewMessageService(redisPubSub, m, logger)
	wsHandler := handler.NewWebSocketHandler(m, msgService, logger)

	// Start subscriber in background
	go msgService.StartSubscriber(ctx)

	// Setup HTTP server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", wsHandler.HandleUpgrade)

	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect first user
	username1 := "Alice"
	color1 := "#ef4444"
	ws1URL := fmt.Sprintf("%s/ws?uid=user1&username=%s&color=%s",
		wsURL,
		url.QueryEscape(username1),
		url.QueryEscape(color1),
	)
	ws1, _, err := websocket.DefaultDialer.Dial(ws1URL, nil)
	if err != nil {
		t.Fatalf("Failed to connect user1: %v", err)
	}
	defer ws1.Close()

	// User1 should receive user_sync with just themselves
	msg1 := readMessage(t, ws1, 2*time.Second)
	if msg1.Type != domain.MessageTypeUserSync {
		t.Errorf("Expected user_sync, got %s", msg1.Type)
	}
	if len(msg1.Users) != 1 {
		t.Errorf("Expected 1 user in sync, got %d", len(msg1.Users))
	}
	if msg1.Users[0].UserID != "user1" {
		t.Errorf("Expected user1 in sync, got %s", msg1.Users[0].UserID)
	}
	assertStringPtr(t, msg1.Users[0].Username, "Alice", "username in sync")
	assertStringPtr(t, msg1.Users[0].Color, "#ef4444", "color in sync")

	// User1 also receives user_joined for themselves (presence events sent to all)
	msg1b := readMessage(t, ws1, 2*time.Second)
	if msg1b.Type != domain.MessageTypeUserJoined {
		t.Errorf("Expected user_joined, got %s", msg1b.Type)
	}
	if msg1b.UserID != "user1" {
		t.Errorf("Expected user1 joined, got %s", msg1b.UserID)
	}

	// Connect second user
	username2 := "Bob"
	color2 := "#10b981"
	ws2URL := fmt.Sprintf("%s/ws?uid=user2&username=%s&color=%s",
		wsURL,
		url.QueryEscape(username2),
		url.QueryEscape(color2),
	)
	ws2, _, err := websocket.DefaultDialer.Dial(ws2URL, nil)
	if err != nil {
		t.Fatalf("Failed to connect user2: %v", err)
	}
	defer ws2.Close()

	// User2 should receive user_sync with both users
	msg2 := readMessage(t, ws2, 2*time.Second)
	if msg2.Type != domain.MessageTypeUserSync {
		t.Errorf("Expected user_sync, got %s", msg2.Type)
	}
	if len(msg2.Users) != 2 {
		t.Errorf("Expected 2 users in sync, got %d", len(msg2.Users))
	}

	// User1 should receive user_joined for user2
	msg3 := readMessage(t, ws1, 2*time.Second)
	if msg3.Type != domain.MessageTypeUserJoined {
		t.Errorf("Expected user_joined, got %s", msg3.Type)
	}
	if msg3.UserID != "user2" {
		t.Errorf("Expected user2 joined, got %s", msg3.UserID)
	}
	assertStringPtr(t, msg3.Username, "Bob", "username in join")
	assertStringPtr(t, msg3.Color, "#10b981", "color in join")

	// User2 also receives user_joined for themselves
	msg3b := readMessage(t, ws2, 2*time.Second)
	if msg3b.Type != domain.MessageTypeUserJoined {
		t.Errorf("Expected user_joined for user2, got %s", msg3b.Type)
	}
	if msg3b.UserID != "user2" {
		t.Errorf("Expected user2 in join, got %s", msg3b.UserID)
	}

	// User2 changes username
	newUsername := "Bob Smith"
	usernameChangeMsg := &domain.Message{
		Type:      domain.MessageTypeUsernameChanged,
		ChannelID: channelID,
		UserID:    "user2",
		Username:  &newUsername,
		Timestamp: time.Now().UnixMilli(),
	}
	if err := ws2.WriteMessage(websocket.TextMessage, usernameChangeMsg.Encode()); err != nil {
		t.Fatalf("Failed to send username change: %v", err)
	}

	// Both users should receive username_changed
	msg4 := readMessage(t, ws1, 2*time.Second)
	if msg4.Type != domain.MessageTypeUsernameChanged {
		t.Errorf("Expected username_changed, got %s", msg4.Type)
	}
	if msg4.UserID != "user2" {
		t.Errorf("Expected user2 username change, got %s", msg4.UserID)
	}
	assertStringPtr(t, msg4.Username, "Bob Smith", "new username")

	msg5 := readMessage(t, ws2, 2*time.Second)
	if msg5.Type != domain.MessageTypeUsernameChanged {
		t.Errorf("Expected username_changed, got %s", msg5.Type)
	}
	if msg5.UserID != "user2" {
		t.Errorf("Expected user2 username change, got %s", msg5.UserID)
	}

	// User1 changes color
	newColor := "#8b5cf6"
	colorChangeMsg := &domain.Message{
		Type:      domain.MessageTypeColorChanged,
		ChannelID: channelID,
		UserID:    "user1",
		Color:     &newColor,
		Timestamp: time.Now().UnixMilli(),
	}
	if err := ws1.WriteMessage(websocket.TextMessage, colorChangeMsg.Encode()); err != nil {
		t.Fatalf("Failed to send color change: %v", err)
	}

	// Both users should receive color_changed
	msg6 := readMessage(t, ws1, 2*time.Second)
	if msg6.Type != domain.MessageTypeColorChanged {
		t.Errorf("Expected color_changed, got %s", msg6.Type)
	}
	if msg6.UserID != "user1" {
		t.Errorf("Expected user1 color change, got %s", msg6.UserID)
	}
	assertStringPtr(t, msg6.Color, "#8b5cf6", "new color")

	msg7 := readMessage(t, ws2, 2*time.Second)
	if msg7.Type != domain.MessageTypeColorChanged {
		t.Errorf("Expected color_changed, got %s", msg7.Type)
	}
	if msg7.UserID != "user1" {
		t.Errorf("Expected user1 color change, got %s", msg7.UserID)
	}

	// User2 disconnects
	ws2.Close()

	// User1 should receive user_left
	msg8 := readMessage(t, ws1, 2*time.Second)
	if msg8.Type != domain.MessageTypeUserLeft {
		t.Errorf("Expected user_left, got %s", msg8.Type)
	}
	if msg8.UserID != "user2" {
		t.Errorf("Expected user2 left, got %s", msg8.UserID)
	}

	// Verify only user1 remains in Redis
	users, err := redisPubSub.GetChannelUsers(ctx, channelID)
	if err != nil {
		t.Fatalf("Failed to get channel users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user remaining, got %d", len(users))
	}
	if users[0].UserID != "user1" {
		t.Errorf("Expected user1 to remain, got %s", users[0].UserID)
	}
}

// readMessage reads and decodes a message from WebSocket with timeout
func readMessage(t *testing.T, ws *websocket.Conn, timeout time.Duration) *domain.Message {
	t.Helper()

	ws.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var msg domain.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	return &msg
}

// assertStringPtr asserts that a string pointer has the expected value
func assertStringPtr(t *testing.T, ptr *string, expected string, fieldName string) {
	t.Helper()

	if ptr == nil {
		t.Errorf("Expected %s '%s', got nil", fieldName, expected)
		return
	}
	if *ptr != expected {
		t.Errorf("Expected %s '%s', got '%s'", fieldName, expected, *ptr)
	}
}
