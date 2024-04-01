package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Chat          *ChatConfig          `mapstructure:"chat"`
}

type ChatConfig struct {
	Http struct {
		Server struct {
			Port    string
			MaxConn int64
		}
	}
	Message struct {
		MaxNum        int64
		MaxSizeByte   int64
	}
}

func setDefault() {
	viper.SetDefault("chat.http.server.port", "5001")
	viper.SetDefault("chat.http.server.maxConn", 200)

	viper.SetDefault("chat.message.maxNum", 5000)
	viper.SetDefault("chat.message.maxSizeByte", 4096)
}

func NewConfig() (*Config, error) {
	setDefault()

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}