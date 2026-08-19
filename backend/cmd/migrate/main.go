package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsDirectory = "migrations"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "migrations: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: go run ./cmd/migrate <up|down|status>")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if args[0] == "down" {
		// A rollback is a destructive local-development aid. Forward-only
		// corrective migrations remain the production policy, so refuse down
		// migrations unless both the application environment and database host
		// prove this is a loopback-local target.
		if err := validateLocalDownTarget(os.Getenv("APP_ENV"), databaseURL); err != nil {
			return err
		}
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return errors.New("could not initialize PostgreSQL migration connection")
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return errors.New("could not configure PostgreSQL migrations")
	}

	switch args[0] {
	case "up":
		err = goose.Up(db, migrationsDirectory)
	case "down":
		err = goose.Down(db, migrationsDirectory)
	case "status":
		err = goose.Status(db, migrationsDirectory)
	default:
		return errors.New("migration command must be up, down, or status")
	}
	if err != nil {
		// This is an infrastructure-authorized local/CI command rather than a
		// browser response. Preserve the migration failure while never printing
		// the configured database URL itself.
		return fmt.Errorf("%s failed: %w", args[0], err)
	}

	return nil
}

func validateLocalDownTarget(appEnvironment string, databaseURL string) error {
	if strings.TrimSpace(appEnvironment) != "local" {
		return errors.New("down migrations require APP_ENV=local")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Hostname() == "" {
		return errors.New("down migrations require a loopback PostgreSQL URL")
	}

	hostname := strings.TrimSpace(parsed.Hostname())
	if strings.EqualFold(hostname, "localhost") {
		return nil
	}
	address := net.ParseIP(hostname)
	if address == nil || !address.IsLoopback() {
		return errors.New("down migrations require a loopback PostgreSQL URL")
	}

	return nil
}
