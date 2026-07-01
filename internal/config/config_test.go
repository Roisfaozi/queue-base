package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYSQL_USER", "testuser")
	t.Setenv("MYSQL_PASSWORD", "testpass")
	t.Setenv("MYSQL_DBNAME", "testdb")
	t.Setenv("MYSQL_PORT", "3306")
	t.Setenv("JWT_ACCESS_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_REFRESH_SECRET", "01234567890123456789012345678901")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("JWT_ACCESS_DURATION", "15m")
	t.Setenv("JWT_REFRESH_DURATION", "24h")
}

// ============================================================================
// Table Driven Tests
// ============================================================================

func TestNewConfig_DefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_DefaultValues",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)

				cfg, err := NewConfig()
				require.NoError(t, err)

				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "info", cfg.Log.Level)
				assert.Equal(t, "localhost", cfg.Mysql.Host)
				assert.Equal(t, 3306, cfg.Mysql.Port)
				assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
				assert.Equal(t, true, cfg.Casbin.Enabled)
				assert.Equal(t, true, cfg.RateLimit.Enabled)
				assert.Equal(t, "memory", cfg.RateLimit.Store)
				assert.Equal(t, "local", cfg.Storage.Driver)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_JWTDurations(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_JWTDurations",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)

				cfg, err := NewConfig()
				require.NoError(t, err)
				assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenDuration)
				assert.Equal(t, 24*time.Hour, cfg.JWT.RefreshTokenDuration)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_SecurityDefaults(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_SecurityDefaults",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)

				cfg, err := NewConfig()
				require.NoError(t, err)

				assert.Equal(t, 5, cfg.Security.MaxLoginAttempts)
				assert.Equal(t, 30*time.Minute, cfg.Security.LockoutDuration)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_CircuitBreakerDefaults(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_CircuitBreakerDefaults",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)

				cfg, err := NewConfig()
				require.NoError(t, err)

				assert.True(t, cfg.CircuitBreaker.Enabled)
				assert.Equal(t, uint32(5), cfg.CircuitBreaker.MaxRequests)
				assert.Equal(t, 60*time.Second, cfg.CircuitBreaker.Interval)
				assert.Equal(t, 30*time.Second, cfg.CircuitBreaker.Timeout)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_MetricsAuthValidation(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_MetricsAuthValidation",
			category: "edge",
			run: func(t *testing.T) {
				setupTestEnv(t)
				t.Setenv("METRICS_AUTH_ENABLED", "true")
				t.Setenv("METRICS_USERNAME", "")
				t.Setenv("METRICS_PASSWORD", "")

				_, err := NewConfig()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "metrics auth is enabled but username or password is missing")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_TrustedProxiesParsing(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_TrustedProxiesParsing",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)
				t.Setenv("SERVER_TRUSTED_PROXIES", "10.0.0.1, 10.0.0.2, 192.168.1.1")

				cfg, err := NewConfig()
				require.NoError(t, err)

				assert.Equal(t, []string{"10.0.0.1", "10.0.0.2", "192.168.1.1"}, cfg.Server.TrustedProxies)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_CORSOriginsParsing(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_CORSOriginsParsing",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)
				t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://example.com")

				cfg, err := NewConfig()
				require.NoError(t, err)

				assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.CORS.AllowedOrigins)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewConfig_StorageDrivers(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NewConfig_StorageDrivers",
			category: "positive",
			run: func(t *testing.T) {
				setupTestEnv(t)

				cfg, err := NewConfig()
				require.NoError(t, err)

				assert.Equal(t, "local", cfg.Storage.Driver)
				assert.Equal(t, "./uploads", cfg.Storage.Local.RootPath)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
