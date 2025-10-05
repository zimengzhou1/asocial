package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
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

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	FirebaseCredentialsPath string `mapstructure:"firebase_credentials_path"`
	AppURL                  string `mapstructure:"app_url"`
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
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "asocial")
	v.SetDefault("database.password", "asocial_dev_password")
	v.SetDefault("database.dbname", "asocial")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("auth.firebase_credentials_path", "")
	v.SetDefault("auth.app_url", "http://localhost")

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
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.dbname", "DB_NAME")
	v.BindEnv("database.sslmode", "DB_SSLMODE")
	v.BindEnv("auth.firebase_credentials_path", "FIREBASE_CREDENTIALS_PATH")
	v.BindEnv("auth.app_url", "APP_URL")

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
