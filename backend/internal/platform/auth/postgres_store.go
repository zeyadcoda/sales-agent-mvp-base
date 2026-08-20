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
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

type queryExecutor interface {
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
	if err := createSession(ctx, store.db, session); err != nil {
		return fmt.Errorf("create super admin session: %w", err)
	}

	return nil
}

func createSession(ctx context.Context, db queryExecutor, session NewSession) error {
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

	if _, err := db.Exec(
		ctx,
		query,
		session.SuperAdminID,
		session.TokenHash,
		session.CSRFToken,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastSeenAt,
	); err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) CreateOTPChallenge(
	ctx context.Context,
	challenge NewOTPChallenge,
) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin OTP challenge creation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock the account even when it has no existing challenge rows. This
	// serializes simultaneous password successes so exactly one challenge flow
	// remains usable for a Super Admin.
	var lockedAdminID string
	if err := tx.QueryRow(
		ctx,
		`SELECT id::text FROM super_admin_accounts WHERE id = $1::uuid FOR UPDATE`,
		challenge.SuperAdminID,
	).Scan(&lockedAdminID); err != nil {
		return fmt.Errorf("lock Super Admin for OTP challenge creation: %w", err)
	}

	const invalidatePriorQuery = `
		UPDATE super_admin_auth_challenges
		SET invalidated_at = GREATEST($2, created_at)
		WHERE super_admin_id = $1::uuid
		  AND consumed_at IS NULL
		  AND invalidated_at IS NULL
	`
	if _, err := tx.Exec(
		ctx,
		invalidatePriorQuery,
		challenge.SuperAdminID,
		challenge.CreatedAt,
	); err != nil {
		return fmt.Errorf("invalidate prior OTP challenges: %w", err)
	}

	const query = `
		INSERT INTO super_admin_auth_challenges (
			id,
			super_admin_id,
			otp_hash,
			created_at,
			expires_at,
			failed_attempts,
			resend_available_at,
			delivery_version,
			active_version,
			delivery_started_at,
			activated_at
		)
		VALUES ($1, $2::uuid, $3, $4, $5, 0, $6, $7, NULL, $8, NULL)
	`

	if _, err := tx.Exec(
		ctx,
		query,
		challenge.ID,
		challenge.SuperAdminID,
		challenge.OTPHash,
		challenge.CreatedAt,
		challenge.ExpiresAt,
		challenge.ResendAvailableAt,
		challenge.DeliveryVersion,
		challenge.DeliveryStartedAt,
	); err != nil {
		return fmt.Errorf("create OTP challenge: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit OTP challenge creation: %w", err)
	}

	return nil
}

func (store *PostgresStore) ActivateOTPChallenge(
	ctx context.Context,
	challengeID string,
	deliveryVersion int,
	activatedAt time.Time,
) error {
	const query = `
		UPDATE super_admin_auth_challenges
		SET
			active_version = delivery_version,
			activated_at = $3
		WHERE id = $1
		  AND delivery_version = $2
		  AND active_version IS NULL
		  AND consumed_at IS NULL
		  AND invalidated_at IS NULL
		  AND failed_attempts < 5
		  AND expires_at > $3
	`

	result, err := store.db.Exec(ctx, query, challengeID, deliveryVersion, activatedAt)
	if err != nil {
		return fmt.Errorf("activate OTP challenge: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrOTPChallengeNotFound
	}

	return nil
}

func (store *PostgresStore) InvalidateOTPChallengeDelivery(
	ctx context.Context,
	challengeID string,
	deliveryVersion int,
	invalidatedAt time.Time,
) error {
	const query = `
		UPDATE super_admin_auth_challenges
		SET invalidated_at = $3
		WHERE id = $1
		  AND delivery_version = $2
		  AND consumed_at IS NULL
		  AND invalidated_at IS NULL
	`

	result, err := store.db.Exec(ctx, query, challengeID, deliveryVersion, invalidatedAt)
	if err != nil {
		return fmt.Errorf("invalidate OTP challenge delivery: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrOTPChallengeNotFound
	}

	return nil
}

func (store *PostgresStore) BeginOTPChallengeResend(
	ctx context.Context,
	challengeID string,
	otpHash []byte,
	startedAt time.Time,
	expiresAt time.Time,
	resendAvailableAt time.Time,
) (OTPDelivery, error) {
	if len(otpHash) != 32 {
		return OTPDelivery{}, errors.New("invalid OTP hash length")
	}

	tx, err := store.db.Begin(ctx)
	if err != nil {
		return OTPDelivery{}, fmt.Errorf("begin OTP resend: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockOTPChallengeAdmin(ctx, tx, challengeID); err != nil {
		return OTPDelivery{}, err
	}
	challenge, err := findOTPChallenge(ctx, tx, challengeID, true)
	if err != nil {
		return OTPDelivery{}, err
	}
	if err := usableOTPChallenge(challenge, startedAt); err != nil {
		return OTPDelivery{}, err
	}
	if startedAt.Before(challenge.ResendAvailableAt) {
		return OTPDelivery{}, &OTPResendTooEarlyError{
			RetryAfter: retryAfterDuration(challenge.ResendAvailableAt.Sub(startedAt)),
		}
	}
	if otpHashesEqual(challenge.OTPHash, otpHash) {
		return OTPDelivery{}, ErrOTPRotationCollision
	}

	const updateQuery = `
		UPDATE super_admin_auth_challenges
		SET
			otp_hash = $2,
			expires_at = $3,
			resend_available_at = $4,
			delivery_version = delivery_version + 1,
			active_version = NULL,
			delivery_started_at = $5,
			activated_at = NULL
		WHERE id = $1
	`
	if _, err := tx.Exec(
		ctx,
		updateQuery,
		challengeID,
		otpHash,
		expiresAt,
		resendAvailableAt,
		startedAt,
	); err != nil {
		return OTPDelivery{}, fmt.Errorf("rotate OTP challenge: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return OTPDelivery{}, fmt.Errorf("commit OTP resend: %w", err)
	}

	return OTPDelivery{
		ChallengeID:       challenge.ID,
		SuperAdmin:        challenge.SuperAdmin,
		DeliveryVersion:   challenge.DeliveryVersion + 1,
		ExpiresAt:         expiresAt,
		ResendAvailableAt: resendAvailableAt,
	}, nil
}

func (store *PostgresStore) VerifyOTPChallengeAndCreateSession(
	ctx context.Context,
	challengeID string,
	candidateHash []byte,
	verifiedAt time.Time,
	session NewSession,
) (SuperAdmin, error) {
	if len(candidateHash) != 32 {
		return SuperAdmin{}, errors.New("invalid candidate OTP hash length")
	}

	tx, err := store.db.Begin(ctx)
	if err != nil {
		return SuperAdmin{}, fmt.Errorf("begin OTP verification: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Account-then-challenge matches challenge creation's lock order. It makes
	// active-account verification authoritative in this transaction without
	// introducing an avoidable lock-order inversion.
	if err := lockOTPChallengeAdmin(ctx, tx, challengeID); err != nil {
		return SuperAdmin{}, err
	}
	challenge, err := findOTPChallenge(ctx, tx, challengeID, true)
	if err != nil {
		return SuperAdmin{}, err
	}
	if err := usableOTPChallenge(challenge, verifiedAt); err != nil {
		return SuperAdmin{}, err
	}
	if len(challenge.OTPHash) != 32 {
		return SuperAdmin{}, errors.New("malformed persisted OTP hash")
	}

	if !otpHashesEqual(challenge.OTPHash, candidateHash) {
		failedAttempts := challenge.FailedAttempts + 1
		const failureQuery = `
			UPDATE super_admin_auth_challenges
			SET
				failed_attempts = $2::smallint,
				invalidated_at = CASE WHEN $2::smallint >= 5 THEN $3 ELSE invalidated_at END
			WHERE id = $1
		`
		if _, err := tx.Exec(ctx, failureQuery, challengeID, failedAttempts, verifiedAt); err != nil {
			return SuperAdmin{}, fmt.Errorf("record failed OTP attempt: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return SuperAdmin{}, fmt.Errorf("commit failed OTP attempt: %w", err)
		}
		if failedAttempts >= maxOTPFailedAttempts {
			return SuperAdmin{}, ErrOTPAttemptsExceeded
		}

		return SuperAdmin{}, ErrOTPInvalid
	}

	const consumeQuery = `
		UPDATE super_admin_auth_challenges
		SET consumed_at = $2
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, consumeQuery, challengeID, verifiedAt); err != nil {
		return SuperAdmin{}, fmt.Errorf("consume OTP challenge: %w", err)
	}
	// The account ID comes only from the locked challenge row. A browser can
	// never choose which privileged account receives the authenticated session.
	session.SuperAdminID = challenge.SuperAdmin.ID
	if err := createSession(ctx, tx, session); err != nil {
		return SuperAdmin{}, fmt.Errorf("create session after OTP verification: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return SuperAdmin{}, fmt.Errorf("commit OTP verification: %w", err)
	}

	return publicSuperAdmin(challenge.SuperAdmin), nil
}

func (store *PostgresStore) FindOTPChallenge(
	ctx context.Context,
	challengeID string,
) (OTPChallenge, error) {
	return findOTPChallenge(ctx, store.db, challengeID, false)
}

func lockOTPChallengeAdmin(
	ctx context.Context,
	tx pgx.Tx,
	challengeID string,
) error {
	const query = `
		SELECT a.id::text
		FROM super_admin_auth_challenges AS c
		JOIN super_admin_accounts AS a ON a.id = c.super_admin_id
		WHERE c.id = $1
		FOR UPDATE OF a
	`

	var adminID string
	if err := tx.QueryRow(ctx, query, challengeID).Scan(&adminID); errors.Is(err, pgx.ErrNoRows) {
		return ErrOTPChallengeNotFound
	} else if err != nil {
		return fmt.Errorf("lock OTP challenge Super Admin: %w", err)
	}

	return nil
}

func findOTPChallenge(
	ctx context.Context,
	db queryExecutor,
	challengeID string,
	forUpdate bool,
) (OTPChallenge, error) {
	query := `
		SELECT
			c.id,
			c.otp_hash,
			c.created_at,
			c.expires_at,
			c.failed_attempts,
			c.resend_available_at,
			c.delivery_version,
			c.active_version,
			c.delivery_started_at,
			c.activated_at,
			c.consumed_at,
			c.invalidated_at,
			a.id::text,
			a.email,
			a.display_name,
			a.is_active,
			a.created_at,
			a.updated_at
		FROM super_admin_auth_challenges AS c
		JOIN super_admin_accounts AS a ON a.id = c.super_admin_id
		WHERE c.id = $1
	`
	if forUpdate {
		query += " FOR UPDATE OF c"
	}

	var challenge OTPChallenge
	err := db.QueryRow(ctx, query, challengeID).Scan(
		&challenge.ID,
		&challenge.OTPHash,
		&challenge.CreatedAt,
		&challenge.ExpiresAt,
		&challenge.FailedAttempts,
		&challenge.ResendAvailableAt,
		&challenge.DeliveryVersion,
		&challenge.ActiveVersion,
		&challenge.DeliveryStartedAt,
		&challenge.ActivatedAt,
		&challenge.ConsumedAt,
		&challenge.InvalidatedAt,
		&challenge.SuperAdmin.ID,
		&challenge.SuperAdmin.Email,
		&challenge.SuperAdmin.DisplayName,
		&challenge.SuperAdmin.IsActive,
		&challenge.SuperAdmin.CreatedAt,
		&challenge.SuperAdmin.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OTPChallenge{}, ErrOTPChallengeNotFound
	}
	if err != nil {
		return OTPChallenge{}, fmt.Errorf("find OTP challenge: %w", err)
	}

	return challenge, nil
}

func usableOTPChallenge(challenge OTPChallenge, now time.Time) error {
	switch {
	case challenge.FailedAttempts >= maxOTPFailedAttempts:
		return ErrOTPAttemptsExceeded
	case challenge.ConsumedAt != nil:
		return ErrOTPConsumed
	case challenge.InvalidatedAt != nil || !challenge.SuperAdmin.IsActive:
		return ErrOTPInvalidated
	case !challenge.ExpiresAt.After(now):
		return ErrOTPExpired
	case challenge.ActiveVersion == nil || *challenge.ActiveVersion != challenge.DeliveryVersion:
		return ErrOTPDeliveryPending
	default:
		return nil
	}
}

func retryAfterDuration(wait time.Duration) time.Duration {
	if wait <= 0 {
		return time.Second
	}

	return ((wait + time.Second - 1) / time.Second) * time.Second
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
