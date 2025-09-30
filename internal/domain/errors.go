package domain

import "errors"

var (
	// ErrInvalidMessage indicates the message format is invalid
	ErrInvalidMessage = errors.New("invalid message format")

	// ErrChannelNotFound indicates the channel does not exist
	ErrChannelNotFound = errors.New("channel not found")

	// ErrUserNotFound indicates the user does not exist
	ErrUserNotFound = errors.New("user not found")

	// ErrPublishFailed indicates message publishing failed
	ErrPublishFailed = errors.New("failed to publish message")

	// ErrRedisConnection indicates Redis connection error
	ErrRedisConnection = errors.New("redis connection error")
)
