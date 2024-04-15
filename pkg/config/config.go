package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Chat          *ChatConfig          `mapstructure:"chat"`
	Kafka 			 *KafkaConfig         `mapstructure:"kafka"`
}

type ChatConfig struct {
	Http struct {
		Server struct {
			Port    string
			MaxConn int64
		}
	}
	Subscriber struct {
		Id string
	}
	Message struct {
		MaxNum        int64
		MaxSizeByte   int64
	}
}

type KafkaConfig struct {
	Addrs   string
	Version string
}

func setDefault() {
	viper.SetDefault("chat.http.server.port", "5001")
	viper.SetDefault("chat.http.server.maxConn", 200)

	viper.SetDefault("chat.message.maxNum", 5000)
	viper.SetDefault("chat.message.maxSizeByte", 4096)

	viper.SetDefault("kafka.addrs", "kafka:9092")
	viper.SetDefault("kafka.version", "1.0.0")
	viper.SetDefault("chat.subscriber.id", "rc.msg.pub")
	//viper.SetDefault("chat.subscriber.id", "rc.msg."+os.Getenv("HOSTNAME"))
}

func NewConfig() (*Config, error) {
	setDefault()

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}