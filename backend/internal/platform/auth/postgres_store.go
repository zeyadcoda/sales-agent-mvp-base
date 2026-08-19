package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the minimum parameterized-query capability required by this
// repository. Keeping it narrow prevents the auth application service from
// obtaining a connection pool or constructing SQL itself.
type DBTX interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

type PostgresStore struct {
	db DBTX
}

func NewPostgresStore(db DBTX) (*PostgresStore, error) {
	if db == nil {
		return nil, errors.New("PostgreSQL auth repository requires a database")
	}

	return &PostgresStore{db: db}, nil
}

func (store *PostgresStore) FindSuperAdminByEmail(
	ctx context.Context,
	normalizedEmail string,
) (SuperAdmin, error) {
	const query = `
		SELECT
			id::text,
			email,
			password_hash,
			display_name,
			is_active,
			created_at,
			updated_at
		FROM super_admin_accounts
		WHERE email = $1
	`

	var admin SuperAdmin
	err := store.db.QueryRow(ctx, query, normalizedEmail).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.DisplayName,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SuperAdmin{}, ErrSuperAdminNotFound
	}
	if err != nil {
		return SuperAdmin{}, fmt.Errorf("find super admin: %w", err)
	}

	return admin, nil
}

func (store *PostgresStore) CreateSession(ctx context.Context, session NewSession) error {
	const query = `
		INSERT INTO super_admin_sessions (
			super_admin_id,
			token_hash,
			csrf_token,
			created_at,
			expires_at,
			last_seen_at
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
	`

	if _, err := store.db.Exec(
		ctx,
		query,
		session.SuperAdminID,
		session.TokenHash,
		session.CSRFToken,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastSeenAt,
	); err != nil {
		return fmt.Errorf("create super admin session: %w", err)
	}

	return nil
}

func (store *PostgresStore) FindSessionByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) (Session, error) {
	const query = `
		SELECT
			s.id::text,
			s.csrf_token,
			s.created_at,
			s.expires_at,
			s.revoked_at,
			s.last_seen_at,
			a.id::text,
			a.email,
			a.display_name,
			a.is_active,
			a.created_at,
			a.updated_at
		FROM super_admin_sessions AS s
		JOIN super_admin_accounts AS a ON a.id = s.super_admin_id
		WHERE s.token_hash = $1
	`

	var session Session
	err := store.db.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.CSRFToken,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.LastSeenAt,
		&session.SuperAdmin.ID,
		&session.SuperAdmin.Email,
		&session.SuperAdmin.DisplayName,
		&session.SuperAdmin.IsActive,
		&session.SuperAdmin.CreatedAt,
		&session.SuperAdmin.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("find super admin session: %w", err)
	}

	return session, nil
}

func (store *PostgresStore) TouchSession(
	ctx context.Context,
	sessionID string,
	seenAt time.Time,
) error {
	const query = `
		UPDATE super_admin_sessions
		SET last_seen_at = $2
		WHERE id = $1::uuid
	`

	if _, err := store.db.Exec(ctx, query, sessionID, seenAt); err != nil {
		return fmt.Errorf("touch super admin session: %w", err)
	}

	return nil
}

func (store *PostgresStore) RevokeSession(
	ctx context.Context,
	sessionID string,
	revokedAt time.Time,
) error {
	const query = `
		UPDATE super_admin_sessions
		SET revoked_at = $2
		WHERE id = $1::uuid
		  AND revoked_at IS NULL
	`

	if _, err := store.db.Exec(ctx, query, sessionID, revokedAt); err != nil {
		return fmt.Errorf("revoke super admin session: %w", err)
	}

	return nil
}

func (store *PostgresStore) CreateSuperAdmin(
	ctx context.Context,
	account NewSuperAdmin,
) (SuperAdmin, error) {
	const query = `
		INSERT INTO super_admin_accounts (
			email,
			password_hash,
			display_name,
			is_active
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id::text,
			email,
			display_name,
			is_active,
			created_at,
			updated_at
	`

	var created SuperAdmin
	err := store.db.QueryRow(
		ctx,
		query,
		account.Email,
		account.PasswordHash,
		account.DisplayName,
		account.IsActive,
	).Scan(
		&created.ID,
		&created.Email,
		&created.DisplayName,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return SuperAdmin{}, ErrSuperAdminExists
	}
	if err != nil {
		return SuperAdmin{}, fmt.Errorf("create super admin: %w", err)
	}

	return created, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
