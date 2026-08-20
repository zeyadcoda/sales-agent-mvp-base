package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"salesagent.local/backend/internal/platform/auth"
)

type fakeFixtureDatabase struct {
	query             string
	args              []any
	execErr           error
	commandTag        pgconn.CommandTag
	execHadDeadline   bool
	execTimeRemaining time.Duration
	closed            bool
}

func (db *fakeFixtureDatabase) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {
	db.query = query
	db.args = append([]any(nil), args...)
	deadline, ok := ctx.Deadline()
	db.execHadDeadline = ok
	if ok {
		db.execTimeRemaining = time.Until(deadline)
	}
	return db.commandTag, db.execErr
}

func (db *fakeFixtureDatabase) Close() {
	db.closed = true
}

type fixtureOpenCall struct {
	databaseURL   string
	hadDeadline   bool
	timeRemaining time.Duration
}

func TestRunExpiresOneChallengeWithBoundedParameterizedUpdate(t *testing.T) {
	t.Parallel()

	challengeID := integrationChallengeID(61)
	db := &fakeFixtureDatabase{commandTag: pgconn.NewCommandTag("UPDATE 1")}
	var openCall fixtureOpenCall
	opener := func(ctx context.Context, databaseURL string) (fixtureDatabase, error) {
		openCall.databaseURL = databaseURL
		deadline, ok := ctx.Deadline()
		openCall.hadDeadline = ok
		if ok {
			openCall.timeRemaining = time.Until(deadline)
		}
		return db, nil
	}
	getenv := fixtureEnvironment(map[string]string{
		"APP_ENV":           "test",
		"TEST_DATABASE_URL": "postgres://user:secret@127.0.0.1:5432/sales_agent_integration_test?sslmode=disable",
	})

	if err := run(
		context.Background(),
		[]string{"expire-challenge", challengeID},
		getenv,
		opener,
	); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !openCall.hadDeadline || openCall.timeRemaining <= 0 || openCall.timeRemaining > fixtureDatabaseTimeout {
		t.Fatalf("open context deadline = %v, remaining %s", openCall.hadDeadline, openCall.timeRemaining)
	}
	if !db.execHadDeadline || db.execTimeRemaining <= 0 || db.execTimeRemaining > fixtureDatabaseTimeout {
		t.Fatalf("exec context deadline = %v, remaining %s", db.execHadDeadline, db.execTimeRemaining)
	}
	if !db.closed {
		t.Fatal("database was not closed")
	}
	if len(db.args) != 1 || db.args[0] != challengeID {
		t.Fatalf("Exec() args = %#v, want only challenge ID", db.args)
	}
	if !strings.Contains(db.query, "WHERE id = $1") || strings.Contains(db.query, challengeID) {
		t.Fatalf("update is not safely parameterized: %q", db.query)
	}
	if !strings.Contains(db.query, "created_at + INTERVAL '1 microsecond'") {
		t.Fatalf("update does not preserve expiry constraint: %q", db.query)
	}
}

func TestRunRejectsInvalidCommandBeforeEnvironmentOrDatabaseAccess(t *testing.T) {
	t.Parallel()

	tests := [][]string{
		nil,
		{},
		{"expire-challenge"},
		{"expire-challenge", integrationChallengeID(1), "extra"},
		{"delete-challenge", integrationChallengeID(1)},
	}
	for _, args := range tests {
		args := args
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			t.Parallel()
			environmentRead := false
			opened := false
			err := run(
				context.Background(),
				args,
				func(string) string {
					environmentRead = true
					return ""
				},
				func(context.Context, string) (fixtureDatabase, error) {
					opened = true
					return nil, nil
				},
			)
			if err == nil || !strings.Contains(err.Error(), "usage:") {
				t.Fatalf("run(%v) error = %v, want usage", args, err)
			}
			if environmentRead || opened {
				t.Fatal("invalid command reached environment/database access")
			}
		})
	}
}

func TestRunRequiresExactTestEnvironment(t *testing.T) {
	t.Parallel()

	for _, environment := range []string{"", "local", "production", "TEST", " test ", "testing"} {
		environment := environment
		t.Run(environment, func(t *testing.T) {
			t.Parallel()
			opened := false
			err := run(
				context.Background(),
				[]string{"expire-challenge", integrationChallengeID(2)},
				fixtureEnvironment(map[string]string{"APP_ENV": environment}),
				func(context.Context, string) (fixtureDatabase, error) {
					opened = true
					return nil, nil
				},
			)
			if err == nil || !strings.Contains(err.Error(), "APP_ENV=test") {
				t.Fatalf("run() error = %v, want test-environment rejection", err)
			}
			if opened {
				t.Fatal("unsafe environment reached database opener")
			}
		})
	}
}

func TestValidateFixtureDatabaseTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		databaseURL     string
		wantErrContains string
	}{
		{name: "localhost", databaseURL: "postgres://user:secret@localhost:5432/sales_agent_integration_test?sslmode=disable"},
		{name: "IPv4 loopback range", databaseURL: "postgresql://user:secret@127.0.0.2:5432/sales_agent_integration_test"},
		{name: "IPv6 loopback", databaseURL: "postgres://user:secret@[::1]:5432/sales_agent_integration_test"},
		{name: "non PostgreSQL", databaseURL: "mysql://localhost/sales_agent_integration_test", wantErrContains: "postgres or postgresql"},
		{name: "remote", databaseURL: "postgres://user:secret@db.example.com/sales_agent_integration_test", wantErrContains: "hostname must be loopback"},
		{name: "missing host", databaseURL: "postgres:///sales_agent_integration_test", wantErrContains: "loopback PostgreSQL"},
		{name: "wrong suffix", databaseURL: "postgres://localhost/sales_agent_test", wantErrContains: "end exactly with _integration_test"},
		{name: "suffix not terminal", databaseURL: "postgres://localhost/sales_agent_integration_test_backup", wantErrContains: "end exactly with _integration_test"},
		{name: "nested path", databaseURL: "postgres://localhost/archive/sales_agent_integration_test", wantErrContains: "end exactly with _integration_test"},
		{name: "host override", databaseURL: "postgres://localhost/sales_agent_integration_test?host=db.example.com", wantErrContains: "must not override"},
		{name: "hostaddr override", databaseURL: "postgres://localhost/sales_agent_integration_test?hostaddr=203.0.113.1", wantErrContains: "must not override"},
		{name: "dbname override", databaseURL: "postgres://localhost/sales_agent_integration_test?dbname=production", wantErrContains: "must not override"},
		{name: "database override", databaseURL: "postgres://localhost/sales_agent_integration_test?database=production", wantErrContains: "must not override"},
		{name: "service override", databaseURL: "postgres://localhost/sales_agent_integration_test?service=production", wantErrContains: "must not override"},
		{name: "servicefile override", databaseURL: "postgres://localhost/sales_agent_integration_test?servicefile=/tmp/service", wantErrContains: "must not override"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateFixtureDatabaseTarget(test.databaseURL)
			if test.wantErrContains == "" {
				if err != nil {
					t.Fatalf("validateFixtureDatabaseTarget() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErrContains) {
				t.Fatalf("validateFixtureDatabaseTarget() error = %v, want %q", err, test.wantErrContains)
			}
			if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), test.databaseURL) {
				t.Fatalf("validation error leaked database credentials/URL: %v", err)
			}
		})
	}
}

func TestRunRequiresDatabaseURLAndCanonicalChallengeID(t *testing.T) {
	t.Parallel()

	safeURL := "postgres://localhost/sales_agent_integration_test"
	tests := []struct {
		name        string
		databaseURL string
		challengeID string
		want        string
	}{
		{name: "missing database URL", challengeID: integrationChallengeID(3), want: "TEST_DATABASE_URL is required"},
		{name: "malformed challenge", databaseURL: safeURL, challengeID: "not-a-challenge", want: "canonical 256-bit"},
		{name: "padded challenge", databaseURL: safeURL, challengeID: integrationChallengeID(3) + "=", want: "canonical 256-bit"},
		{name: "short challenge", databaseURL: safeURL, challengeID: integrationChallengeID(3)[:42], want: "canonical 256-bit"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			opened := false
			err := run(
				context.Background(),
				[]string{"expire-challenge", test.challengeID},
				fixtureEnvironment(map[string]string{
					"APP_ENV":           "test",
					"TEST_DATABASE_URL": test.databaseURL,
				}),
				func(context.Context, string) (fixtureDatabase, error) {
					opened = true
					return nil, nil
				},
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("run() error = %v, want %q", err, test.want)
			}
			if opened {
				t.Fatal("invalid input reached database opener")
			}
		})
	}
}

func TestRunCollapsesDatabaseErrorsAndRequiresExactlyOneRow(t *testing.T) {
	t.Parallel()

	challengeID := integrationChallengeID(62)
	getenv := fixtureEnvironment(map[string]string{
		"APP_ENV":           "test",
		"TEST_DATABASE_URL": "postgres://user:do-not-leak@127.0.0.1/sales_agent_integration_test",
	})
	tests := []struct {
		name      string
		opener    fixtureDatabaseOpener
		want      string
		wantClose bool
	}{
		{
			name: "open failure",
			opener: func(context.Context, string) (fixtureDatabase, error) {
				return nil, errors.New("postgres://user:do-not-leak@internal/production")
			},
			want: "could not initialize fixture PostgreSQL connection",
		},
		{
			name: "exec failure",
			opener: func(context.Context, string) (fixtureDatabase, error) {
				return &fakeFixtureDatabase{execErr: errors.New("SQL secret detail")}, nil
			},
			want:      "could not expire OTP challenge",
			wantClose: true,
		},
		{
			name: "zero rows",
			opener: func(context.Context, string) (fixtureDatabase, error) {
				return &fakeFixtureDatabase{commandTag: pgconn.NewCommandTag("UPDATE 0")}, nil
			},
			want:      "not found or is not eligible",
			wantClose: true,
		},
		{
			name: "multiple rows",
			opener: func(context.Context, string) (fixtureDatabase, error) {
				return &fakeFixtureDatabase{commandTag: pgconn.NewCommandTag("UPDATE 2")}, nil
			},
			want:      "not found or is not eligible",
			wantClose: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var openedDB *fakeFixtureDatabase
			opener := test.opener
			wrappedOpener := func(ctx context.Context, databaseURL string) (fixtureDatabase, error) {
				db, err := opener(ctx, databaseURL)
				if typed, ok := db.(*fakeFixtureDatabase); ok {
					openedDB = typed
				}
				return db, err
			}
			err := run(
				context.Background(),
				[]string{"expire-challenge", challengeID},
				getenv,
				wrappedOpener,
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("run() error = %v, want %q", err, test.want)
			}
			for _, secret := range []string{"do-not-leak", "SQL secret detail", "internal/production"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("run() leaked internal detail %q: %v", secret, err)
				}
			}
			if test.wantClose && (openedDB == nil || !openedDB.closed) {
				t.Fatal("opened database was not closed after failure")
			}
		})
	}
}

func fixtureEnvironment(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func integrationChallengeID(value byte) string {
	candidate := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
	if !auth.ValidOTPChallengeID(candidate) {
		panic("test helper generated a non-canonical challenge ID")
	}
	return candidate
}
