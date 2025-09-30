package domain

import (
	"encoding/json"
	"time"
)

// Message represents a chat message
type Message struct {
	MessageID string   `json:"message_id"`
	ChannelID string   `json:"channel_id"`
	UserID    string   `json:"user_id"`
	Payload   string   `json:"payload"`
	Position  Position `json:"position"`
	Timestamp int64    `json:"timestamp"`
}

// Position represents the x,y coordinates on the canvas
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
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

// NewMessage creates a new message with timestamp
func NewMessage(messageID, channelID, userID, payload string, position Position) *Message {
	return &Message{
		MessageID: messageID,
		ChannelID: channelID,
		UserID:    userID,
		Payload:   payload,
		Position:  position,
		Timestamp: time.Now().UnixMilli(),
	}
}
