package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	Redis   RedisConfig
	Webhook WebhookConfig
	App     AppConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	Database     int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	MaxConnAge   time.Duration
	PoolTimeout  time.Duration
	IdleTimeout  time.Duration
}

type WebhookConfig struct {
	URL     string
	Timeout time.Duration
	AuthKey string
}

type AppConfig struct {
	Environment         string
	SchedulerInterval   time.Duration
	MessagesPerInterval int
	MaxRetryCount       int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "auto_message_sender"),
			Timeout:  getDurationEnv("MONGODB_TIMEOUT", 10*time.Second),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getIntEnv("REDIS_DATABASE", 0),
			DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 2),
			MaxConnAge:   getDurationEnv("REDIS_MAX_CONN_AGE", 30*time.Minute),
			PoolTimeout:  getDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second),
			IdleTimeout:  getDurationEnv("REDIS_IDLE_TIMEOUT", 5*time.Minute),
		},
		Webhook: WebhookConfig{
			URL:     getEnv("WEBHOOK_URL", "https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd"),
			Timeout: getDurationEnv("WEBHOOK_TIMEOUT", 30*time.Second),
			AuthKey: getEnv("WEBHOOK_AUTH_KEY", "INS.me1x9uMcyYGlhKKQVPoc.bO3j9aZwRTOcA2Ywo"),
		},
		App: AppConfig{
			Environment:         getEnv("ENVIRONMENT", "development"),
			SchedulerInterval:   getDurationEnv("SCHEDULER_INTERVAL", 2*time.Minute),
			MessagesPerInterval: getIntEnv("MESSAGES_PER_INTERVAL", 2),
			MaxRetryCount:       getIntEnv("MAX_RETRY_COUNT", 3),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
