package config

import (
	"testing"

	"salesagent.local/backend/internal/runtimeenv"
)

// setValidEnvironment creates a complete valid configuration for each test.
//
// t.Setenv automatically restores the previous environment after the test,
// which prevents one test from changing another test's configuration.
func setValidEnvironment(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "local")
	t.Setenv("EXECUTION_ENV", "TEST")
	t.Setenv("API_HOST", "127.0.0.1")
	t.Setenv("API_PORT", "8081")

	t.Setenv(
		"DATABASE_URL",
		"postgres://sales_agent:sales_agent_local@127.0.0.1:5432/sales_agent?sslmode=disable",
	)

	t.Setenv(
		"REDIS_URL",
		"redis://127.0.0.1:6379/0",
	)
}

// TestLoadValidConfiguration verifies that valid environment variables become
// one validated Config object that the application can safely use.
func TestLoadValidConfiguration(t *testing.T) {
	setValidEnvironment(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.AppEnvironment != AppLocal {
		t.Fatalf(
			"AppEnvironment = %q, want %q",
			cfg.AppEnvironment,
			AppLocal,
		)
	}

	if cfg.ExecutionEnvironment != runtimeenv.Test {
		t.Fatalf(
			"ExecutionEnvironment = %q, want %q",
			cfg.ExecutionEnvironment,
			runtimeenv.Test,
		)
	}

	if cfg.APIAddress() != "127.0.0.1:8081" {
		t.Fatalf(
			"APIAddress() = %q, want %q",
			cfg.APIAddress(),
			"127.0.0.1:8081",
		)
	}
}

// TestLoadRejectsInvalidExecutionEnvironment proves that an unknown execution
// environment cannot silently fall back to TEST or PRODUCTION.
func TestLoadRejectsInvalidExecutionEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("EXECUTION_ENV", "something-else")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an invalid execution environment")
	}
}

// TestLoadRejectsInvalidPort proves that malformed network configuration is
// rejected during startup rather than failing later while serving requests.
func TestLoadRejectsInvalidPort(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("API_PORT", "99999")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an invalid API port")
	}
}

// TestLoadRejectsMissingDatabaseURL proves that the application cannot start
// without an explicitly configured PostgreSQL database.
func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a missing database URL")
	}
}

// TestLoadRejectsInvalidDatabaseScheme prevents accidental configuration with
// an HTTP endpoint or another non-PostgreSQL service.
func TestLoadRejectsInvalidDatabaseScheme(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_URL", "https://example.com/database")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a non-PostgreSQL database URL")
	}
}

// TestLoadRejectsMissingRedisURL proves that Redis must also be configured
// explicitly before the application considers configuration valid.
func TestLoadRejectsMissingRedisURL(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("REDIS_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a missing Redis URL")
	}
}
