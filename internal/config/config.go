package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port           string `mapstructure:"port"`
	MaxConnections int    `mapstructure:"max_connections"`
	MaxMessageSize int    `mapstructure:"max_message_size"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	Channel  string `mapstructure:"channel"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", "3001")
	v.SetDefault("server.max_connections", 200)
	v.SetDefault("server.max_message_size", 4096)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.channel", "chat:messages")

	// Read config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Read from environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("ASOCIAL")

	// Allow environment variable overrides
	// e.g., ASOCIAL_SERVER_PORT=8080, ASOCIAL_REDIS_ADDR=redis:6379
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("redis.addr", "REDIS_ADDR")
	v.BindEnv("redis.password", "REDIS_PASSWORD")

	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
		fmt.Fprintln(os.Stderr, "No config file found, using defaults and environment variables")
	} else {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", v.ConfigFileUsed())
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
