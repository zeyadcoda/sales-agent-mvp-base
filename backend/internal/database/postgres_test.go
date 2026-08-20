package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// TestOpenRejectsMissingDatabaseURL verifies that database configuration fails
// closed when DATABASE_URL is missing instead of guessing a connection target.
func TestOpenRejectsMissingDatabaseURL(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), "")

	if db != nil {
		t.Fatal("database should be nil when the database URL is missing")
	}

	if !errors.Is(err, ErrDatabaseURLRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDatabaseURLRequired)
	}
}

func TestPostgresBeginRejectsUninitializedDatabase(t *testing.T) {
	t.Parallel()

	var db *Postgres
	transaction, err := db.Begin(context.Background())

	if transaction != nil {
		t.Fatal("transaction should be nil for an uninitialized database")
	}
	if !errors.Is(err, ErrDatabaseNotInitialized) {
		t.Fatalf("error = %v, want %v", err, ErrDatabaseNotInitialized)
	}
}

// TestPostgresPing proves that the Go backend can communicate with a real
// PostgreSQL instance.
//
// The test is skipped during normal unit-test runs unless TEST_DATABASE_URL is
// explicitly supplied. This prevents ordinary tests from unexpectedly talking
// to a developer or production database.
func TestPostgresPing(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
}
