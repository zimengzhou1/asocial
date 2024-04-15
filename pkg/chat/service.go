package chat

import (
	"asocial/pkg/common"
	"asocial/pkg/config"
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/olahol/melody"
)

var (
	MessagePubTopic = "rc.msg.pub"
)

type MessageSubscriber struct {
	subscriberID string
	router       *message.Router
	sub          message.Subscriber
	m            MelodyChatConn
}

func NewMessageSubscriber(name string, router *message.Router, config *config.Config, sub message.Subscriber, m MelodyChatConn) (*MessageSubscriber, error) {
	fmt.Println("Subscriber id: ", config.Chat.Subscriber.Id)
	return &MessageSubscriber{
		subscriberID: config.Chat.Subscriber.Id,
		router:       router,
		sub:          sub,
		m:            m,
	}, nil
}

func (s *MessageSubscriber) HandleMessage(msg *message.Message) error {
	message, err := DecodeToMessage([]byte(msg.Payload))
	fmt.Println("Subscriber message handler: ", message.Payload)
	if err != nil {
		return err
	}
	return s.sendMessage(message)
}

func (s *MessageSubscriber) RegisterHandler() {
	s.router.AddNoPublisherHandler(
		"randomchat_message_handler",
		s.subscriberID,
		s.sub,
		s.HandleMessage,
	)
}

func (s *MessageSubscriber) Run() error {
	return s.router.Run(context.Background())
}

func (s *MessageSubscriber) GracefulStop() error {
	return s.router.Close()
}

func (s *MessageSubscriber) sendMessage(message *Message) error {
	return s.m.BroadcastFilter(message.Encode(), func(sess *melody.Session) bool {
		channelID, exist := sess.Get(sessCidKey)
		if !exist {
			return false
		}
		userID, exist := sess.Get("user")
		if !exist {
			return false
		}
		channelExist := channelID == message.ChannelID
		userNotSame := userID != message.UserID
		return channelExist && userNotSame
	})
}

type MessageService interface {
	BroadcastTextMessage(ctx context.Context, msg *Message) error
	PublishMessage(ctx context.Context, msg *Message) error
}

type MessageServiceImpl struct {
	sf       common.IDGenerator
	p        message.Publisher
}

func NewMessageServiceImpl(sf common.IDGenerator, p message.Publisher) *MessageServiceImpl {
	return &MessageServiceImpl{sf, p}
}

func (svc *MessageServiceImpl) BroadcastTextMessage(ctx context.Context, msg *Message) error {
	// messageID, err := svc.sf.NextID()
	// if err != nil {
	// 	return fmt.Errorf("error create snowflake ID for text message: %w", err)
	// }
	
	// msg := Message{
	// 	MessageID: messageID,
	// 	ChannelID: channelID,
	// 	UserID:    userID,
	// 	Payload:   payload,
	// 	Time:      time.Now().UnixMilli(),
	// }

	if err := svc.PublishMessage(ctx, msg); err != nil {
		return fmt.Errorf("error broadcast text message: %w", err)
	}
	return nil
}

func (svc *MessageServiceImpl) PublishMessage(ctx context.Context, msg *Message) error {
	err := svc.p.Publish(MessagePubTopic, message.NewMessage(
		watermill.NewUUID(),
		msg.Encode(),
	))

	fmt.Println("published message: ", msg.Payload)

	if err != nil {
		return fmt.Errorf("error publish message: %w", err)
	}
	return nil
}