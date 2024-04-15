package chat

import (
	"encoding/json"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Message struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Payload   string `json:"payload"`
	Position Position `json:"position"`
}

func DecodeToMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (m *Message) Encode() []byte {
	result, _ := json.Marshal(m)
	return result
}