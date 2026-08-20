package main

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/platform/auth"
)

const fixtureDatabaseTimeout = 5 * time.Second

type fixtureDatabase interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Close()
}

type fixtureDatabaseOpener func(
	ctx context.Context,
	databaseURL string,
) (fixtureDatabase, error)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, openFixtureDatabase); err != nil {
		fmt.Fprintf(os.Stderr, "E2E authentication fixture: %v\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	openDatabase fixtureDatabaseOpener,
) error {
	if len(args) != 2 || args[0] != "expire-challenge" {
		return errors.New("usage: go run ./cmd/e2e-auth-fixture expire-challenge <challenge-id>")
	}
	if getenv == nil || getenv("APP_ENV") != "test" {
		return errors.New("E2E authentication fixtures require APP_ENV=test")
	}

	databaseURL := ""
	if getenv != nil {
		databaseURL = getenv("TEST_DATABASE_URL")
	}
	if strings.TrimSpace(databaseURL) == "" {
		return errors.New("TEST_DATABASE_URL is required")
	}
	if err := validateFixtureDatabaseTarget(databaseURL); err != nil {
		return err
	}

	challengeID := args[1]
	if !auth.ValidOTPChallengeID(challengeID) {
		return errors.New("challenge ID must be a canonical 256-bit base64url value")
	}
	if openDatabase == nil {
		return errors.New("could not initialize fixture PostgreSQL connection")
	}

	openCtx, cancelOpen := context.WithTimeout(ctx, fixtureDatabaseTimeout)
	db, err := openDatabase(openCtx, databaseURL)
	cancelOpen()
	if err != nil || db == nil {
		// Database errors can contain credentials or topology. This command is
		// intentionally no more verbose than the browser-facing auth boundary.
		return errors.New("could not initialize fixture PostgreSQL connection")
	}
	defer db.Close()

	updateCtx, cancelUpdate := context.WithTimeout(ctx, fixtureDatabaseTimeout)
	defer cancelUpdate()
	return expireChallenge(updateCtx, db, challengeID)
}

func openFixtureDatabase(
	ctx context.Context,
	databaseURL string,
) (fixtureDatabase, error) {
	return database.Open(ctx, databaseURL)
}

func expireChallenge(
	ctx context.Context,
	db fixtureDatabase,
	challengeID string,
) error {
	if db == nil || !auth.ValidOTPChallengeID(challengeID) {
		return errors.New("could not expire OTP challenge")
	}

	// Preserve the expires_at > created_at constraint while moving an
	// API-created, unconsumed challenge into the past. The opaque challenge ID
	// remains a query parameter and never becomes SQL text.
	const query = `
		UPDATE super_admin_auth_challenges
		SET expires_at = created_at + INTERVAL '1 microsecond'
		WHERE id = $1
		  AND consumed_at IS NULL
		  AND created_at + INTERVAL '1 microsecond' < NOW()
	`
	result, err := db.Exec(ctx, query, challengeID)
	if err != nil {
		return errors.New("could not expire OTP challenge")
	}
	if result.RowsAffected() != 1 {
		return errors.New("OTP challenge was not found or is not eligible for expiry")
	}

	return nil
}

func validateFixtureDatabaseTarget(databaseURL string) error {
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Hostname() == "" {
		return errors.New("TEST_DATABASE_URL must identify a loopback PostgreSQL integration-test database")
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errors.New("TEST_DATABASE_URL must use the postgres or postgresql scheme")
	}
	for _, key := range []string{
		"host",
		"hostaddr",
		"dbname",
		"database",
		"service",
		"servicefile",
	} {
		if parsed.Query().Has(key) {
			return errors.New("TEST_DATABASE_URL must not override its target in query parameters")
		}
	}

	hostname := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	loopback := hostname == "localhost"
	if address, parseErr := netip.ParseAddr(hostname); parseErr == nil {
		loopback = address.Unmap().IsLoopback()
	}
	if !loopback {
		return errors.New("TEST_DATABASE_URL hostname must be loopback")
	}

	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" ||
		strings.Contains(databaseName, "/") ||
		!strings.HasSuffix(databaseName, "_integration_test") {
		return errors.New("TEST_DATABASE_URL database name must end exactly with _integration_test")
	}

	return nil
}
