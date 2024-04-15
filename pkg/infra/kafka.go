package infra

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	"asocial/pkg/config"
)

var (
	logger = watermill.NewStdLogger(false, false)
)

func NewKafkaPublisher(config *config.Config) (message.Publisher, error) {
	brokers := strings.Split(config.Kafka.Addrs, ",")
	fmt.Println("brokers: ", brokers)
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}
	return kafkaPublisher, nil
}

func NewKafkaSubscriber(config *config.Config) (message.Subscriber, error) {
	saramaConfig := sarama.NewConfig()
	saramaVersion, err := sarama.ParseKafkaVersion(config.Kafka.Version)
	if err != nil {
		return nil, err
	}
	saramaConfig.Version = saramaVersion
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers: strings.Split(config.Kafka.Addrs, ","),
			Unmarshaler: kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaSubscriber, nil
}

func NewBrokerRouter(name string) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          logger,
		}.Middleware,
		middleware.Recoverer,
	)
	return router, nil
}