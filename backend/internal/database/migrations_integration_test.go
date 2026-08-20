package database

import (
	"database/sql"
	"errors"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestMigrationsUpAndDown proves the complete Goose chain can build an empty
// PostgreSQL database and roll back locally. It requires a dedicated database
// with an explicitly test-only name because rollback is intentionally destructive.
func TestMigrationsUpAndDown(t *testing.T) {
	databaseURL := os.Getenv("TEST_MIGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_MIGRATION_DATABASE_URL is not set; skipping migration integration test")
	}
	assertDedicatedTestDatabase(t, os.Getenv("APP_ENV"), databaseURL)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close migration database: %v", err)
		}
	})

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set Goose dialect: %v", err)
	}
	migrationsDirectory := migrationDirectory(t)

	// A prior interrupted run may have left the dedicated test database at a
	// migration version. Rolling it down first gives this test a true clean
	// starting point without ever touching a non-test database.
	if err := goose.DownTo(db, migrationsDirectory, 0); err != nil {
		t.Fatalf("reset migration test database: %v", err)
	}

	if err := goose.UpTo(db, migrationsDirectory, 1); err != nil {
		t.Fatalf("apply migration 00001: %v", err)
	}
	assertMigrationTableExists(t, db, "schema_marker", true)

	if err := goose.UpTo(db, migrationsDirectory, 2); err != nil {
		t.Fatalf("apply migration 00002: %v", err)
	}
	assertMigrationTableExists(t, db, "super_admin_accounts", true)
	assertMigrationTableExists(t, db, "super_admin_sessions", true)

	if err := goose.UpTo(db, migrationsDirectory, 3); err != nil {
		t.Fatalf("apply migration 00003: %v", err)
	}
	assertMigrationTableExists(t, db, "super_admin_auth_challenges", true)

	// Prove the new migration can roll back independently without disturbing
	// the historical Milestone 01 schema, then apply cleanly again.
	if err := goose.Down(db, migrationsDirectory); err != nil {
		t.Fatalf("roll migration 00003 down: %v", err)
	}
	assertMigrationTableExists(t, db, "super_admin_auth_challenges", false)
	assertMigrationTableExists(t, db, "super_admin_accounts", true)
	assertMigrationTableExists(t, db, "super_admin_sessions", true)
	if err := goose.Up(db, migrationsDirectory); err != nil {
		t.Fatalf("re-apply migration 00003: %v", err)
	}
	assertMigrationTableExists(t, db, "super_admin_auth_challenges", true)

	if err := goose.DownTo(db, migrationsDirectory, 0); err != nil {
		t.Fatalf("roll migrations down: %v", err)
	}
	for _, table := range []string{
		"schema_marker",
		"super_admin_accounts",
		"super_admin_sessions",
		"super_admin_auth_challenges",
	} {
		assertMigrationTableExists(t, db, table, false)
	}
}

func assertMigrationTableExists(t *testing.T, db *sql.DB, table string, want bool) {
	t.Helper()

	var exists bool
	if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, "public."+table).Scan(&exists); err != nil {
		t.Fatalf("check table %s: %v", table, err)
	}
	if exists != want {
		t.Fatalf("table %s existence = %v, want %v", table, exists, want)
	}
}

func assertDedicatedTestDatabase(t *testing.T, appEnvironment, databaseURL string) {
	t.Helper()
	if err := validateMigrationTestTarget(appEnvironment, databaseURL); err != nil {
		t.Fatalf("unsafe migration integration test target: %v", err)
	}
}

func validateMigrationTestTarget(appEnvironment, databaseURL string) error {
	if appEnvironment != "test" {
		return errors.New("APP_ENV must be exactly test")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return errors.New("TEST_MIGRATION_DATABASE_URL is invalid")
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errors.New("TEST_MIGRATION_DATABASE_URL must use the postgres or postgresql scheme")
	}
	for _, key := range []string{"host", "hostaddr", "dbname", "database", "service", "servicefile"} {
		if parsed.Query().Has(key) {
			// PostgreSQL URI query parameters can override the authority/path that
			// this guard validates. Reject target-changing parameters so Goose and
			// the safety check cannot resolve different databases.
			return errors.New("TEST_MIGRATION_DATABASE_URL must not override its target in query parameters")
		}
	}

	hostname := strings.ToLower(parsed.Hostname())
	loopback := hostname == "localhost"
	if address, parseErr := netip.ParseAddr(hostname); parseErr == nil {
		loopback = address.Unmap().IsLoopback()
	}
	if !loopback {
		return errors.New("TEST_MIGRATION_DATABASE_URL hostname must be loopback")
	}

	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" || strings.Contains(databaseName, "/") || !strings.HasSuffix(databaseName, "_migration_test") {
		return errors.New("TEST_MIGRATION_DATABASE_URL database name must end exactly with _migration_test")
	}

	return nil
}

func TestValidateMigrationTestTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		appEnvironment  string
		databaseURL     string
		wantErrContains string
	}{
		{
			name:           "localhost is allowed",
			appEnvironment: "test",
			databaseURL:    "postgres://user:password@localhost:5432/sales_agent_migration_test?sslmode=disable",
		},
		{
			name:           "IPv4 loopback range is allowed",
			appEnvironment: "test",
			databaseURL:    "postgresql://user:password@127.0.0.2:5432/sales_agent_migration_test",
		},
		{
			name:           "IPv6 loopback is allowed",
			appEnvironment: "test",
			databaseURL:    "postgres://user:password@[::1]:5432/sales_agent_migration_test",
		},
		{
			name:            "missing test environment is rejected",
			appEnvironment:  "",
			databaseURL:     "postgres://localhost/sales_agent_migration_test",
			wantErrContains: "APP_ENV must be exactly test",
		},
		{
			name:            "local environment is rejected",
			appEnvironment:  "local",
			databaseURL:     "postgres://localhost/sales_agent_migration_test",
			wantErrContains: "APP_ENV must be exactly test",
		},
		{
			name:            "non-PostgreSQL scheme is rejected",
			appEnvironment:  "test",
			databaseURL:     "mysql://localhost/sales_agent_migration_test",
			wantErrContains: "postgres or postgresql scheme",
		},
		{
			name:            "remote hostname is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://db.example.com/sales_agent_migration_test",
			wantErrContains: "hostname must be loopback",
		},
		{
			name:            "query host override is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://localhost/sales_agent_migration_test?host=db.example.com",
			wantErrContains: "must not override its target",
		},
		{
			name:            "query database override is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://localhost/sales_agent_migration_test?dbname=sales_agent",
			wantErrContains: "must not override its target",
		},
		{
			name:            "private network address is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://10.0.0.8/sales_agent_migration_test",
			wantErrContains: "hostname must be loopback",
		},
		{
			name:            "generic test suffix is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://localhost/sales_agent_test",
			wantErrContains: "must end exactly with _migration_test",
		},
		{
			name:            "migration test text not at end is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://localhost/sales_agent_migration_test_backup",
			wantErrContains: "must end exactly with _migration_test",
		},
		{
			name:            "nested database path is rejected",
			appEnvironment:  "test",
			databaseURL:     "postgres://localhost/archive/sales_agent_migration_test",
			wantErrContains: "must end exactly with _migration_test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateMigrationTestTarget(test.appEnvironment, test.databaseURL)
			if test.wantErrContains == "" {
				if err != nil {
					t.Fatalf("validateMigrationTestTarget() error = %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("validateMigrationTestTarget() error = nil, want error containing %q", test.wantErrContains)
			}
			if !strings.Contains(err.Error(), test.wantErrContains) {
				t.Fatalf("validateMigrationTestTarget() error = %q, want error containing %q", err, test.wantErrContains)
			}
		})
	}
}

func migrationDirectory(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve migration test path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "migrations"))
}
