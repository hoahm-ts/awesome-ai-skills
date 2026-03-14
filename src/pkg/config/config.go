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
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Temporal TemporalConfig `yaml:"temporal"`
	Datadog  DatadogConfig  `yaml:"datadog"`
}

// AppConfig holds HTTP server and general application settings.
type AppConfig struct {
	Name            string        `yaml:"name"`
	Env             string        `yaml:"env"`
	Port            int           `yaml:"port"`
	LogLevel        string        `yaml:"log_level"`
	Timeout         time.Duration `yaml:"-"`
	ReadTimeout     time.Duration `yaml:"-"`
	WriteTimeout    time.Duration `yaml:"-"`
	IdleTimeout     time.Duration `yaml:"-"`
	ShutdownTimeout time.Duration `yaml:"-"`
}

// DatabaseConfig holds PostgreSQL connection settings.
// DATABASE_DSN takes precedence when set; otherwise the DSN is built from per-field settings.
type DatabaseConfig struct {
	// DSN is the full postgres connection URL. When set it overrides per-field settings.
	DSN string `yaml:"dsn"`

	// Per-field settings — used when DSN is empty.
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	// TLS/SSL settings follow GCP Cloud SQL and AWS RDS best practices.
	// SSLMode values: disable | require | verify-ca | verify-full
	SSLMode     string `yaml:"ssl_mode"`
	SSLCert     string `yaml:"ssl_cert"`      // path to client certificate (mutual TLS)
	SSLKey      string `yaml:"ssl_key"`       // path to client private key (mutual TLS)
	SSLRootCert string `yaml:"ssl_root_cert"` // path to root CA cert (verify-ca / verify-full)

	// Connection pool settings (not loaded from YAML).
	MaxOpenConns int           `yaml:"-"`
	MaxIdleConns int           `yaml:"-"`
	MaxLifetime  time.Duration `yaml:"-"`
}

// EffectiveDSN returns the DSN to use for opening the database connection.
// It returns the explicit DSN when set, otherwise constructs one from per-field settings.
func (d DatabaseConfig) EffectiveDSN() string {
	if d.DSN != "" {
		return d.DSN
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode,
	)
	if d.SSLCert != "" {
		dsn += " sslcert=" + d.SSLCert
	}
	if d.SSLKey != "" {
		dsn += " sslkey=" + d.SSLKey
	}
	if d.SSLRootCert != "" {
		dsn += " sslrootcert=" + d.SSLRootCert
	}
	return dsn
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// KafkaConfig holds Kafka connection settings.
type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupID string   `yaml:"group_id"`
}

// TemporalConfig holds Temporal connection settings.
type TemporalConfig struct {
	HostPort  string `yaml:"host_port"`
	Namespace string `yaml:"namespace"`
	TaskQueue string `yaml:"task_queue"`
}

// DatadogConfig holds Datadog observability settings.
type DatadogConfig struct {
	ServiceName string `yaml:"service"`
	Env         string `yaml:"env"`
	AgentHost   string `yaml:"agent_host"`
}

// Load reads configuration from environment variables and returns a validated Config.
// Returns an error if any required value is missing.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "awesome-ai-skills"),
			Env:             getEnv("APP_ENV", "development"),
			Port:            getEnvInt("APP_PORT", 8080),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			Timeout:         getEnvDuration("APP_TIMEOUT_SECONDS", 30),
			ReadTimeout:     getEnvDuration("APP_READ_TIMEOUT_SECONDS", 15),
			WriteTimeout:    getEnvDuration("APP_WRITE_TIMEOUT_SECONDS", 15),
			IdleTimeout:     getEnvDuration("APP_IDLE_TIMEOUT_SECONDS", 60),
			ShutdownTimeout: getEnvDuration("APP_SHUTDOWN_TIMEOUT_SECONDS", 30),
		},
		Database: DatabaseConfig{
			DSN:          getEnv("DATABASE_DSN", ""),
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			Name:         getEnv("DB_NAME", ""),
			User:         getEnv("DB_USER", ""),
			Password:     getEnv("DB_PASSWORD", ""),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			SSLCert:      getEnv("DB_SSL_CERT", ""),
			SSLKey:       getEnv("DB_SSL_KEY", ""),
			SSLRootCert:  getEnv("DB_SSL_ROOT_CERT", ""),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvDuration("DB_CONN_MAX_LIFETIME_SECONDS", 300),
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
			ServiceName: getEnv("APP_NAME", "awesome-ai-skills"),
			Env:         getEnv("DD_ENV", "development"),
			AgentHost:   getEnv("DD_AGENT_HOST", "localhost"),
		},
	}

	if cfg.Database.DSN == "" && (cfg.Database.Host == "" || cfg.Database.Name == "" || cfg.Database.User == "") {
		return nil, fmt.Errorf("database configuration required: set DATABASE_DSN or DB_HOST + DB_NAME + DB_USER")
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

// getEnvDuration reads an env var as an integer number of seconds and returns a time.Duration.
func getEnvDuration(key string, defaultSeconds int) time.Duration {
	return time.Duration(getEnvInt(key, defaultSeconds)) * time.Second
}
