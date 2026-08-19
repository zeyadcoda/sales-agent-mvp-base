package main

import (
	"strings"
	"testing"
)

func TestRunValidatesCommandBeforeDatabaseAccess(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example.invalid/test")

	for _, args := range [][]string{nil, {"up", "extra"}, {"unknown"}} {
		err := run(args)
		if err == nil {
			t.Fatalf("run(%v) should fail", args)
		}
		if strings.Contains(err.Error(), "postgres://") {
			t.Fatalf("run(%v) leaked database URL: %v", args, err)
		}
	}
}

func TestRunRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if err := run([]string{"up"}); err == nil {
		t.Fatal("run() should require DATABASE_URL")
	}
}

func TestDownMigrationTargetMustBeLocalAndLoopback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		environment string
		databaseURL string
		wantError   bool
	}{
		{
			name:        "local IPv4 loopback",
			environment: "local",
			databaseURL: "postgres://user:secret@127.0.0.1:5432/sales_agent",
		},
		{
			name:        "local IPv6 loopback",
			environment: "local",
			databaseURL: "postgres://user:secret@[::1]:5432/sales_agent",
		},
		{
			name:        "local hostname",
			environment: "local",
			databaseURL: "postgres://user:secret@localhost:5432/sales_agent",
		},
		{
			name:        "production environment",
			environment: "production",
			databaseURL: "postgres://user:secret@127.0.0.1:5432/sales_agent",
			wantError:   true,
		},
		{
			name:        "test environment",
			environment: "test",
			databaseURL: "postgres://user:secret@127.0.0.1:5432/sales_agent_test",
			wantError:   true,
		},
		{
			name:        "remote host",
			environment: "local",
			databaseURL: "postgres://user:secret@db.example.com:5432/sales_agent",
			wantError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateLocalDownTarget(test.environment, test.databaseURL)
			if (err != nil) != test.wantError {
				t.Fatalf("validateLocalDownTarget() error = %v, wantError = %v", err, test.wantError)
			}
			if err != nil && strings.Contains(err.Error(), "secret") {
				t.Fatalf("validateLocalDownTarget() leaked database credentials: %v", err)
			}
		})
	}
}
