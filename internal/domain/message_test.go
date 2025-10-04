package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	messageID := "msg-123"
	channelID := "default"
	userID := "user-456"
	payload := "Hello, World!"
	position := Position{X: 100, Y: 200}

	before := time.Now().UnixMilli()
	msg := NewMessage(messageID, channelID, userID, payload, position)
	after := time.Now().UnixMilli()

	if msg.MessageID == nil || *msg.MessageID != messageID {
		t.Errorf("Expected MessageID %s, got %v", messageID, msg.MessageID)
	}
	if msg.ChannelID != channelID {
		t.Errorf("Expected ChannelID %s, got %s", channelID, msg.ChannelID)
	}
	if msg.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, msg.UserID)
	}
	if msg.Payload == nil || *msg.Payload != payload {
		t.Errorf("Expected Payload %s, got %v", payload, msg.Payload)
	}
	if msg.Position == nil || msg.Position.X != position.X || msg.Position.Y != position.Y {
		t.Errorf("Expected Position %+v, got %+v", position, msg.Position)
	}
	if msg.Timestamp < before || msg.Timestamp > after {
		t.Errorf("Expected Timestamp between %d and %d, got %d", before, after, msg.Timestamp)
	}
}

func TestMessage_Encode(t *testing.T) {
	messageID := "msg-123"
	payload := "Hello"
	position := &Position{X: 50, Y: 75}

	msg := &Message{
		MessageID: &messageID,
		ChannelID: "default",
		UserID:    "user-456",
		Payload:   &payload,
		Position:  position,
		Timestamp: 1609459200000,
	}

	data := msg.Encode()

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal encoded data: %v", err)
	}

	if decoded["message_id"] != *msg.MessageID {
		t.Errorf("Expected message_id %s, got %v", *msg.MessageID, decoded["message_id"])
	}
	if decoded["channel_id"] != msg.ChannelID {
		t.Errorf("Expected channel_id %s, got %v", msg.ChannelID, decoded["channel_id"])
	}
	if decoded["user_id"] != msg.UserID {
		t.Errorf("Expected user_id %s, got %v", msg.UserID, decoded["user_id"])
	}
	if decoded["payload"] != *msg.Payload {
		t.Errorf("Expected payload %s, got %v", *msg.Payload, decoded["payload"])
	}
}

func TestDecodeMessage(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name: "valid message",
			data: `{
				"message_id": "msg-123",
				"channel_id": "default",
				"user_id": "user-456",
				"payload": "Hello",
				"position": {"x": 100, "y": 200},
				"timestamp": 1609459200000
			}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			data:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := DecodeMessage([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if msg == nil {
					t.Error("Expected message, got nil")
				}
				if msg.MessageID == nil || *msg.MessageID != "msg-123" {
					t.Errorf("Expected MessageID msg-123, got %v", msg.MessageID)
				}
			}
		})
	}
}

func TestMessage_EncodeDecode_RoundTrip(t *testing.T) {
	messageID := "msg-789"
	payload := "Test payload"
	position := &Position{X: 123, Y: 456}

	original := &Message{
		MessageID: &messageID,
		ChannelID: "test-channel",
		UserID:    "user-999",
		Payload:   &payload,
		Position:  position,
		Timestamp: 1609459200000,
	}

	// Encode
	data := original.Encode()

	// Decode
	decoded, err := DecodeMessage(data)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Compare
	if decoded.MessageID == nil || *decoded.MessageID != *original.MessageID {
		t.Errorf("MessageID mismatch: got %v, want %s", decoded.MessageID, *original.MessageID)
	}
	if decoded.ChannelID != original.ChannelID {
		t.Errorf("ChannelID mismatch: got %s, want %s", decoded.ChannelID, original.ChannelID)
	}
	if decoded.UserID != original.UserID {
		t.Errorf("UserID mismatch: got %s, want %s", decoded.UserID, original.UserID)
	}
	if decoded.Payload == nil || *decoded.Payload != *original.Payload {
		t.Errorf("Payload mismatch: got %v, want %s", decoded.Payload, *original.Payload)
	}
	if decoded.Position == nil || decoded.Position.X != original.Position.X || decoded.Position.Y != original.Position.Y {
		t.Errorf("Position mismatch: got %+v, want %+v", decoded.Position, original.Position)
	}
	if decoded.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, original.Timestamp)
	}
}

func TestPosition(t *testing.T) {
	pos := Position{X: 42, Y: 84}

	if pos.X != 42 {
		t.Errorf("Expected X = 42, got %v", pos.X)
	}
	if pos.Y != 84 {
		t.Errorf("Expected Y = 84, got %v", pos.Y)
	}

	// Test JSON marshaling
	data, err := json.Marshal(pos)
	if err != nil {
		t.Fatalf("Failed to marshal Position: %v", err)
	}

	var decoded Position
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Position: %v", err)
	}

	if decoded.X != pos.X || decoded.Y != pos.Y {
		t.Errorf("Position round-trip failed: got %+v, want %+v", decoded, pos)
	}
}

func TestUserInfo_JSON(t *testing.T) {
	username := "Alice"
	color := "#ef4444"

	userInfo := UserInfo{
		UserID:   "user-123",
		Username: &username,
		Color:    &color,
	}

	// Test JSON marshaling
	data, err := json.Marshal(userInfo)
	if err != nil {
		t.Fatalf("Failed to marshal UserInfo: %v", err)
	}

	// Test JSON unmarshaling
	var decoded UserInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal UserInfo: %v", err)
	}

	// Verify fields
	if decoded.UserID != userInfo.UserID {
		t.Errorf("UserID mismatch: got %s, want %s", decoded.UserID, userInfo.UserID)
	}
	if decoded.Username == nil || *decoded.Username != *userInfo.Username {
		t.Errorf("Username mismatch: got %v, want %v", decoded.Username, userInfo.Username)
	}
	if decoded.Color == nil || *decoded.Color != *userInfo.Color {
		t.Errorf("Color mismatch: got %v, want %v", decoded.Color, userInfo.Color)
	}
}

func TestNewUserJoinedMessage(t *testing.T) {
	channelID := "default"
	userID := "user-123"
	username := "Alice"
	color := "#ef4444"

	before := time.Now().UnixMilli()
	msg := NewUserJoinedMessage(channelID, userID, &username, &color)
	after := time.Now().UnixMilli()

	if msg.Type != MessageTypeUserJoined {
		t.Errorf("Expected type %s, got %s", MessageTypeUserJoined, msg.Type)
	}
	if msg.ChannelID != channelID {
		t.Errorf("Expected ChannelID %s, got %s", channelID, msg.ChannelID)
	}
	if msg.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, msg.UserID)
	}
	if msg.Username == nil || *msg.Username != username {
		t.Errorf("Expected Username %s, got %v", username, msg.Username)
	}
	if msg.Color == nil || *msg.Color != color {
		t.Errorf("Expected Color %s, got %v", color, msg.Color)
	}
	if msg.Timestamp < before || msg.Timestamp > after {
		t.Errorf("Expected Timestamp between %d and %d, got %d", before, after, msg.Timestamp)
	}
}

func TestNewUsernameChangedMessage(t *testing.T) {
	channelID := "default"
	userID := "user-123"
	username := "Alice Smith"

	before := time.Now().UnixMilli()
	msg := NewUsernameChangedMessage(channelID, userID, &username)
	after := time.Now().UnixMilli()

	if msg.Type != MessageTypeUsernameChanged {
		t.Errorf("Expected type %s, got %s", MessageTypeUsernameChanged, msg.Type)
	}
	if msg.ChannelID != channelID {
		t.Errorf("Expected ChannelID %s, got %s", channelID, msg.ChannelID)
	}
	if msg.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, msg.UserID)
	}
	if msg.Username == nil || *msg.Username != username {
		t.Errorf("Expected Username %s, got %v", username, msg.Username)
	}
	if msg.Timestamp < before || msg.Timestamp > after {
		t.Errorf("Expected Timestamp between %d and %d, got %d", before, after, msg.Timestamp)
	}
}

func TestNewColorChangedMessage(t *testing.T) {
	channelID := "default"
	userID := "user-123"
	color := "#10b981"

	before := time.Now().UnixMilli()
	msg := NewColorChangedMessage(channelID, userID, &color)
	after := time.Now().UnixMilli()

	if msg.Type != MessageTypeColorChanged {
		t.Errorf("Expected type %s, got %s", MessageTypeColorChanged, msg.Type)
	}
	if msg.ChannelID != channelID {
		t.Errorf("Expected ChannelID %s, got %s", channelID, msg.ChannelID)
	}
	if msg.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, msg.UserID)
	}
	if msg.Color == nil || *msg.Color != color {
		t.Errorf("Expected Color %s, got %v", color, msg.Color)
	}
	if msg.Timestamp < before || msg.Timestamp > after {
		t.Errorf("Expected Timestamp between %d and %d, got %d", before, after, msg.Timestamp)
	}
}

func TestNewUserSyncMessage(t *testing.T) {
	channelID := "default"
	username1 := "Alice"
	color1 := "#ef4444"
	username2 := "Bob"
	color2 := "#10b981"

	users := []UserInfo{
		{UserID: "user-1", Username: &username1, Color: &color1},
		{UserID: "user-2", Username: &username2, Color: &color2},
	}

	before := time.Now().UnixMilli()
	msg := NewUserSyncMessage(channelID, users)
	after := time.Now().UnixMilli()

	if msg.Type != MessageTypeUserSync {
		t.Errorf("Expected type %s, got %s", MessageTypeUserSync, msg.Type)
	}
	if msg.ChannelID != channelID {
		t.Errorf("Expected ChannelID %s, got %s", channelID, msg.ChannelID)
	}
	if msg.UserID != "system" {
		t.Errorf("Expected UserID 'system', got %s", msg.UserID)
	}
	if len(msg.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(msg.Users))
	}
	if msg.Timestamp < before || msg.Timestamp > after {
		t.Errorf("Expected Timestamp between %d and %d, got %d", before, after, msg.Timestamp)
	}

	// Verify user data
	if msg.Users[0].UserID != "user-1" || *msg.Users[0].Username != "Alice" || *msg.Users[0].Color != "#ef4444" {
		t.Errorf("User 1 data mismatch: got %+v", msg.Users[0])
	}
	if msg.Users[1].UserID != "user-2" || *msg.Users[1].Username != "Bob" || *msg.Users[1].Color != "#10b981" {
		t.Errorf("User 2 data mismatch: got %+v", msg.Users[1])
	}
}
