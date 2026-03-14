package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/config"
)

// minimalValidYAML returns a YAML string that passes validation.
func minimalValidYAML() string {
	return `app:
  name: test-app
  env: test
  port: 8080
  log_level: info
  timeout_seconds: 30
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  idle_timeout_seconds: 60
  shutdown_timeout_seconds: 30
database:
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  password: testpass
  ssl_mode: disable
  max_open_conns: 10
  max_idle_conns: 2
  max_lifetime_seconds: 120
redis:
  addr: localhost:6379
  password: ""
  db: 0
kafka:
  brokers:
    - localhost:9092
  group_id: test-group
temporal:
  host_port: localhost:7233
  namespace: default
  task_queue: default
datadog:
  service: test-app
  env: test
  agent_host: localhost
`
}

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "app_config.yml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

func TestLoadFromFile_ValidConfig(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, minimalValidYAML())

	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "test-app", cfg.App.Name)
	require.Equal(t, 8080, cfg.App.Port)
	require.Equal(t, 30*time.Second, cfg.App.Timeout)
	require.Equal(t, 15*time.Second, cfg.App.ReadTimeout)
	require.Equal(t, 15*time.Second, cfg.App.WriteTimeout)
	require.Equal(t, 60*time.Second, cfg.App.IdleTimeout)
	require.Equal(t, 30*time.Second, cfg.App.ShutdownTimeout)
	require.Equal(t, 10, cfg.Database.MaxOpenConns)
	require.Equal(t, 2, cfg.Database.MaxIdleConns)
	require.Equal(t, 120*time.Second, cfg.Database.MaxLifetime)
}

func TestLoadFromFile_DefaultPoolSettings(t *testing.T) {
	t.Parallel()

	yaml := `app:
  name: test-app
  env: test
  port: 8080
  log_level: info
  timeout_seconds: 5
  read_timeout_seconds: 5
  write_timeout_seconds: 5
  idle_timeout_seconds: 5
  shutdown_timeout_seconds: 5
database:
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  password: ""
  ssl_mode: disable
redis:
  addr: localhost:6379
  password: ""
  db: 0
kafka:
  brokers:
    - localhost:9092
  group_id: g
temporal:
  host_port: localhost:7233
  namespace: default
  task_queue: default
datadog:
  service: svc
  env: test
  agent_host: localhost
`
	path := writeTempConfig(t, yaml)

	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err)
	require.Equal(t, 25, cfg.Database.MaxOpenConns)
	require.Equal(t, 5, cfg.Database.MaxIdleConns)
	require.Equal(t, 300*time.Second, cfg.Database.MaxLifetime)
}

func TestLoadFromFile_WithExplicitDSN(t *testing.T) {
	t.Parallel()

	yaml := `app:
  name: test-app
  env: test
  port: 8080
  log_level: info
  timeout_seconds: 5
  read_timeout_seconds: 5
  write_timeout_seconds: 5
  idle_timeout_seconds: 5
  shutdown_timeout_seconds: 5
database:
  dsn: "postgres://user:pass@localhost/mydb?sslmode=disable"
  ssl_mode: disable
redis:
  addr: localhost:6379
  password: ""
  db: 0
kafka:
  brokers:
    - localhost:9092
  group_id: g
temporal:
  host_port: localhost:7233
  namespace: default
  task_queue: default
datadog:
  service: svc
  env: test
  agent_host: localhost
`
	path := writeTempConfig(t, yaml)

	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err)
	require.Equal(t, "postgres://user:pass@localhost/mydb?sslmode=disable", cfg.Database.DSN)
}

func TestLoadFromFile_NegativeTimeoutReturnsError(t *testing.T) {
	t.Parallel()

	yaml := `app:
  name: test-app
  env: test
  port: 8080
  log_level: info
  timeout_seconds: -1
  read_timeout_seconds: 5
  write_timeout_seconds: 5
  idle_timeout_seconds: 5
  shutdown_timeout_seconds: 5
database:
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  password: ""
  ssl_mode: disable
redis:
  addr: localhost:6379
  password: ""
  db: 0
kafka:
  brokers:
    - localhost:9092
  group_id: g
temporal:
  host_port: localhost:7233
  namespace: default
  task_queue: default
datadog:
  service: svc
  env: test
  agent_host: localhost
`
	path := writeTempConfig(t, yaml)

	_, err := config.LoadFromFile(path)
	require.Error(t, err)
}

func TestLoadFromFile_MissingDatabaseConfig(t *testing.T) {
	t.Parallel()

	yaml := `app:
  name: test-app
  env: test
  port: 8080
  log_level: info
  timeout_seconds: 5
  read_timeout_seconds: 5
  write_timeout_seconds: 5
  idle_timeout_seconds: 5
  shutdown_timeout_seconds: 5
database:
  ssl_mode: disable
redis:
  addr: localhost:6379
  password: ""
  db: 0
kafka:
  brokers:
    - localhost:9092
  group_id: g
temporal:
  host_port: localhost:7233
  namespace: default
  task_queue: default
datadog:
  service: svc
  env: test
  agent_host: localhost
`
	path := writeTempConfig(t, yaml)

	_, err := config.LoadFromFile(path)
	require.Error(t, err)
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := config.LoadFromFile("/nonexistent/path/config.yml")
	require.Error(t, err)
}

func TestLoad_UsesCONFIG_PATH(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel.
	path := writeTempConfig(t, minimalValidYAML())
	t.Setenv("CONFIG_PATH", path)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "test-app", cfg.App.Name)
}

func TestLoad_DefaultPath_WhenCONFIG_PATH_Unset(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel.
	// When CONFIG_PATH is empty Load falls back to etc/config/app_config.default.yml
	// relative to the working directory. That path does not exist when tests run
	// inside the pkg/config package directory, so an error is expected.
	t.Setenv("CONFIG_PATH", "")

	_, err := config.Load()
	require.Error(t, err)
}

func TestEffectiveDSN_ExplicitDSN(t *testing.T) {
	t.Parallel()

	d := config.DatabaseConfig{
		DSN: "postgres://user:pass@host/db",
	}
	require.Equal(t, "postgres://user:pass@host/db", d.EffectiveDSN())
}

func TestEffectiveDSN_BuiltFromFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give config.DatabaseConfig
		want string
	}{
		{
			name: "basic fields without TLS",
			give: config.DatabaseConfig{
				Host:    "dbhost",
				Port:    5432,
				Name:    "mydb",
				User:    "myuser",
				Password: "mypass",
				SSLMode: "disable",
			},
			want: "host=dbhost port=5432 dbname=mydb user=myuser password=mypass sslmode=disable",
		},
		{
			name: "with ssl cert and key",
			give: config.DatabaseConfig{
				Host:     "dbhost",
				Port:     5432,
				Name:     "mydb",
				User:     "myuser",
				Password: "mypass",
				SSLMode:  "verify-full",
				SSLCert:  "/path/to/cert",
				SSLKey:   "/path/to/key",
				SSLRootCert: "/path/to/root",
			},
			want: "host=dbhost port=5432 dbname=mydb user=myuser password=mypass sslmode=verify-full sslcert=/path/to/cert sslkey=/path/to/key sslrootcert=/path/to/root",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.give.EffectiveDSN())
		})
	}
}
