package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDatabaseURLRequired is returned when the application starts without a
// PostgreSQL connection string. We fail closed instead of guessing credentials.
var ErrDatabaseURLRequired = errors.New("database URL is required")

// ErrDatabaseNotInitialized protects callers from accidentally using a
// database dependency that was never created.
var ErrDatabaseNotInitialized = errors.New("database is not initialized")

// Postgres owns the PostgreSQL connection pool used by application services.
//
// The pool is private so other packages cannot casually bypass our database
// boundary. Later, sqlc-generated queries will use this connection through
// controlled repository/application-service layers.
type Postgres struct {
	pool *pgxpool.Pool
}

// Open creates the PostgreSQL connection pool from the supplied database URL.
//
// Open intentionally does not claim the database is healthy. Connectivity is
// checked separately through Ping so the API can remain alive while readiness
// correctly reports a temporary database outage.
func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
	// Never silently invent a database address. Missing configuration should
	// stop initialization rather than accidentally connecting somewhere else.
	if strings.TrimSpace(databaseURL) == "" {
		return nil, ErrDatabaseURLRequired
	}

	// ParseConfig validates the connection configuration before we create the
	// pool. We do not log the database URL because it may contain credentials.
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("invalid PostgreSQL connection configuration")
	}

	// NewWithConfig creates a concurrency-safe PostgreSQL connection pool.
	// pgxpool manages reusable connections instead of opening a new connection
	// for every API request.
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("could not initialize PostgreSQL connection pool")
	}

	return &Postgres{
		pool: pool,
	}, nil
}

// Ping verifies that PostgreSQL is actually reachable.
//
// Readiness checks will use this method. A failed Ping means the application
// must report NOT READY instead of pretending the database is healthy.
func (db *Postgres) Ping(ctx context.Context) error {
	if db == nil || db.pool == nil {
		return ErrDatabaseNotInitialized
	}

	return db.pool.Ping(ctx)
}

// Exec and QueryRow expose only the narrow pgx operations needed by
// domain-owned PostgreSQL repositories. The connection pool itself remains
// private so HTTP handlers and Agents cannot acquire unrestricted database
// access.
func (db *Postgres) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {
	if db == nil || db.pool == nil {
		return pgconn.CommandTag{}, ErrDatabaseNotInitialized
	}

	return db.pool.Exec(ctx, query, args...)
}

func (db *Postgres) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if db == nil || db.pool == nil {
		return errorRow{err: ErrDatabaseNotInitialized}
	}

	return db.pool.QueryRow(ctx, query, args...)
}

type errorRow struct {
	err error
}

func (row errorRow) Scan(_ ...any) error {
	return row.err
}

// Close releases all database connections owned by this process.
//
// The API and workers must call Close during graceful shutdown so connections
// are returned cleanly instead of being abandoned.
func (db *Postgres) Close() {
	if db == nil || db.pool == nil {
		return
	}

	db.pool.Close()
}
