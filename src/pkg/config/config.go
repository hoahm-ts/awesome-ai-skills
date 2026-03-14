// Package config loads and validates application configuration from a YAML file.
// The YAML file is rendered at container startup by dockerize from etc/config/template.yml.
// Configuration is read once at startup in the composition root and injected via DI.
package config

import (
	"fmt"
	"os"
	"time"

	pkgyaml "github.com/hoahm-ts/awesome-ai-skills/pkg/yaml"
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
	Name     string `yaml:"name"`
	Env      string `yaml:"env"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`

	// Raw integer seconds read from YAML; the Duration fields below are derived from these.
	TimeoutSeconds         int `yaml:"timeout_seconds"`
	ReadTimeoutSeconds     int `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds    int `yaml:"write_timeout_seconds"`
	IdleTimeoutSeconds     int `yaml:"idle_timeout_seconds"`
	ShutdownTimeoutSeconds int `yaml:"shutdown_timeout_seconds"`

	// Computed from *Seconds fields after loading — not marshalled.
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

	// Connection pool settings. Zero values trigger defaults in LoadFromFile.
	MaxOpenConns       int           `yaml:"max_open_conns"`
	MaxIdleConns       int           `yaml:"max_idle_conns"`
	MaxLifetimeSeconds int           `yaml:"max_lifetime_seconds"`
	MaxLifetime        time.Duration `yaml:"-"` // derived from MaxLifetimeSeconds after loading
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

// Load reads configuration from the YAML file specified by the CONFIG_PATH environment variable,
// falling back to etc/config/app_config.yml when unset.
// The YAML file is normally rendered at container startup by dockerize from etc/config/template.yml.
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "etc/config/app_config.default.yml"
	}
	return LoadFromFile(path)
}

// Default values for database connection-pool settings when omitted from the YAML file.
const (
	defaultMaxOpenConns       = 25
	defaultMaxIdleConns       = 5
	defaultMaxLifetimeSeconds = 300
)

// LoadFromFile reads and validates application configuration from the YAML file at path.
// After parsing, it derives the time.Duration fields from the raw integer-seconds fields
// and applies default values for database connection-pool settings when they are unset.
func LoadFromFile(path string) (*Config, error) {
	cfg, err := pkgyaml.ParseFile[Config](path)
	if err != nil {
		return nil, err
	}

	// Derive time.Duration fields from the raw integer-seconds values.
	cfg.App.Timeout = time.Duration(cfg.App.TimeoutSeconds) * time.Second
	cfg.App.ReadTimeout = time.Duration(cfg.App.ReadTimeoutSeconds) * time.Second
	cfg.App.WriteTimeout = time.Duration(cfg.App.WriteTimeoutSeconds) * time.Second
	cfg.App.IdleTimeout = time.Duration(cfg.App.IdleTimeoutSeconds) * time.Second
	cfg.App.ShutdownTimeout = time.Duration(cfg.App.ShutdownTimeoutSeconds) * time.Second

	// Apply operational defaults for connection-pool settings when absent from the YAML file.
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = defaultMaxOpenConns
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = defaultMaxIdleConns
	}
	if cfg.Database.MaxLifetimeSeconds == 0 {
		cfg.Database.MaxLifetimeSeconds = defaultMaxLifetimeSeconds
	}
	cfg.Database.MaxLifetime = time.Duration(cfg.Database.MaxLifetimeSeconds) * time.Second

	if err := validate(cfg); err != nil {
		return nil, err
	}

	// Normalise: ensure DSN is always set so consumers can use cfg.Database.DSN directly.
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = cfg.Database.EffectiveDSN()
	}

	return cfg, nil
}

// validate checks that required fields are present and that numeric fields are in range.
func validate(cfg *Config) error {
	if cfg.App.TimeoutSeconds < 0 || cfg.App.ReadTimeoutSeconds < 0 ||
		cfg.App.WriteTimeoutSeconds < 0 || cfg.App.IdleTimeoutSeconds < 0 ||
		cfg.App.ShutdownTimeoutSeconds < 0 {
		return fmt.Errorf("app timeout_seconds values must be non-negative")
	}
	if cfg.Database.DSN == "" && (cfg.Database.Host == "" || cfg.Database.Name == "" || cfg.Database.User == "") {
		return fmt.Errorf("database configuration required: set dsn or host + name + user in the config file")
	}
	return nil
}
