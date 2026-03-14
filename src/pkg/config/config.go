// Package config loads and validates application configuration from environment variables.
// Configuration is read once at startup in the composition root and injected via DI.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the complete application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Temporal TemporalConfig
	Datadog  DatadogConfig
}

// AppConfig holds HTTP server and general application settings.
type AppConfig struct {
	Name    string
	Env     string
	Port    int
	Timeout time.Duration
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// KafkaConfig holds Kafka connection settings.
type KafkaConfig struct {
	Brokers []string
	GroupID string
}

// TemporalConfig holds Temporal connection settings.
type TemporalConfig struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

// DatadogConfig holds Datadog observability settings.
type DatadogConfig struct {
	ServiceName string
	Env         string
	AgentHost   string
}

// Load reads configuration from environment variables and returns a validated Config.
// Returns an error if any required value is missing.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "awesome-ai-skills"),
			Env:     getEnv("APP_ENV", "development"),
			Port:    getEnvInt("APP_PORT", 8080),
			Timeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			DSN:          getEnv("DATABASE_DSN", ""),
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKER", "localhost:9092")},
			GroupID: getEnv("KAFKA_GROUP_ID", "awesome-ai-skills"),
		},
		Temporal: TemporalConfig{
			HostPort:  getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
			Namespace: getEnv("TEMPORAL_NAMESPACE", "default"),
			TaskQueue: getEnv("TEMPORAL_TASK_QUEUE", "default"),
		},
		Datadog: DatadogConfig{
			ServiceName: getEnv("DD_SERVICE", "awesome-ai-skills"),
			Env:         getEnv("DD_ENV", "development"),
			AgentHost:   getEnv("DD_AGENT_HOST", "localhost"),
		},
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}
