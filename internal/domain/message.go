package domain

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeChat            MessageType = "chat"
	MessageTypeUserJoined      MessageType = "user_joined"
	MessageTypeUserLeft        MessageType = "user_left"
	MessageTypeUserSync        MessageType = "user_sync"
	MessageTypeUsernameChanged MessageType = "username_changed"
	MessageTypeColorChanged    MessageType = "color_changed"
)

// UserInfo represents a user with ID and optional username and color
type UserInfo struct {
	UserID   string  `json:"user_id"`
	Username *string `json:"username,omitempty"`
	Color    *string `json:"color,omitempty"`
}

// Message represents a chat message or presence event
type Message struct {
	Type      MessageType `json:"type"`
	MessageID *string     `json:"message_id,omitempty"`
	ChannelID string      `json:"channel_id"`
	UserID    string      `json:"user_id"`
	Username  *string     `json:"username,omitempty"` // For user_joined and username_changed
	Color     *string     `json:"color,omitempty"`    // For user_joined and color_changed
	Payload   *string     `json:"payload,omitempty"`
	Position  *Position   `json:"position,omitempty"`
	Users     []UserInfo  `json:"users,omitempty"` // For user_sync messages
	Timestamp int64       `json:"timestamp"`
}

// Position represents the x,y coordinates on the canvas
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Encode serializes the message to JSON bytes
func (m *Message) Encode() []byte {
	data, _ := json.Marshal(m)
	return data
}

// DecodeMessage deserializes JSON bytes to a Message
func DecodeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewMessage creates a new chat message with timestamp
func NewMessage(messageID, channelID, userID, payload string, position Position) *Message {
	return &Message{
		Type:      MessageTypeChat,
		MessageID: &messageID,
		ChannelID: channelID,
		UserID:    userID,
		Payload:   &payload,
		Position:  &position,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewUserJoinedMessage creates a user joined presence event
func NewUserJoinedMessage(channelID, userID string, username, color *string) *Message {
	return &Message{
		Type:      MessageTypeUserJoined,
		ChannelID: channelID,
		UserID:    userID,
		Username:  username,
		Color:     color,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewUserLeftMessage creates a user left presence event
func NewUserLeftMessage(channelID, userID string) *Message {
	return &Message{
		Type:      MessageTypeUserLeft,
		ChannelID: channelID,
		UserID:    userID,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewUserSyncMessage creates a user sync message with the list of all users in channel
func NewUserSyncMessage(channelID string, users []UserInfo) *Message {
	return &Message{
		Type:      MessageTypeUserSync,
		ChannelID: channelID,
		UserID:    "system",
		Users:     users,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewUsernameChangedMessage creates a username changed event
func NewUsernameChangedMessage(channelID, userID string, username *string) *Message {
	return &Message{
		Type:      MessageTypeUsernameChanged,
		ChannelID: channelID,
		UserID:    userID,
		Username:  username,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewColorChangedMessage creates a color changed event
func NewColorChangedMessage(channelID, userID string, color *string) *Message {
	return &Message{
		Type:      MessageTypeColorChanged,
		ChannelID: channelID,
		UserID:    userID,
		Color:     color,
		Timestamp: time.Now().UnixMilli(),
	}
}
