package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"salesagent.local/backend/internal/platform/audit"
)

// HasActiveSuperAdminRecovery is a non-authoritative read hint used only after
// successful password verification. The consuming transaction below repeats
// the account and lifecycle checks under locks before creating any session.
func (store *PostgresStore) HasActiveSuperAdminRecovery(
	ctx context.Context,
	superAdminID string,
	_ time.Time,
) (bool, error) {
	const query = `
		WITH operation_time AS MATERIALIZED (
			SELECT clock_timestamp() AS checked_at
		)
		SELECT EXISTS (
			SELECT 1
			FROM super_admin_recovery_authorizations AS recovery
			JOIN super_admin_accounts AS account
			  ON account.id = recovery.super_admin_id
			CROSS JOIN operation_time
			WHERE recovery.super_admin_id = $1::uuid
			  AND account.is_active
			  AND recovery.consumed_at IS NULL
			  AND recovery.revoked_at IS NULL
			  AND recovery.expired_at IS NULL
			  AND recovery.expires_at > operation_time.checked_at
		)
	`

	var active bool
	if err := store.db.QueryRow(ctx, query, superAdminID).Scan(&active); err != nil {
		return false, fmt.Errorf("check active Super Admin recovery authorization: %w", err)
	}

	return active, nil
}

// ConsumeSuperAdminRecoveryAndCreateSession owns the one allowed emergency
// transition. Recovery consumption, normal session insertion, OTP challenge
// invalidation, and immutable Audit append either all commit or all roll back.
func (store *PostgresStore) ConsumeSuperAdminRecoveryAndCreateSession(
	ctx context.Context,
	superAdminID string,
	_ time.Time,
	session NewSession,
	correlationID string,
) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin Super Admin recovery consumption: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	admin, err := lockRecoveryAdminByID(ctx, tx, superAdminID)
	if err != nil {
		return err
	}
	if !admin.IsActive {
		return ErrRecoveryUnavailable
	}
	consumedAt, err := recoveryDatabaseTime(ctx, tx)
	if err != nil {
		return err
	}

	authorization, err := findActiveRecoveryForUpdate(ctx, tx, admin.ID, consumedAt)
	if err != nil {
		return err
	}

	const consumeQuery = `
		UPDATE super_admin_recovery_authorizations
		SET consumed_at = $2
		WHERE id = $1::uuid
		  AND consumed_at IS NULL
		  AND revoked_at IS NULL
		  AND expired_at IS NULL
		  AND expires_at > $2
	`
	result, err := tx.Exec(ctx, consumeQuery, authorization.ID, consumedAt)
	if err != nil {
		return fmt.Errorf("consume Super Admin recovery authorization: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrRecoveryNotActive
	}
	authorization.ConsumedAt = timePointer(consumedAt)

	// A successful password login supersedes any older pending OTP flow for the
	// same account. This update follows the global account-first lock order.
	const invalidateChallengesQuery = `
		UPDATE super_admin_auth_challenges
		SET invalidated_at = GREATEST($2, created_at)
		WHERE super_admin_id = $1::uuid
		  AND consumed_at IS NULL
		  AND invalidated_at IS NULL
	`
	if _, err := tx.Exec(ctx, invalidateChallengesQuery, admin.ID, consumedAt); err != nil {
		return fmt.Errorf("invalidate OTP challenges after recovery: %w", err)
	}

	// The account ID comes only from the locked database row. No caller can
	// redirect a consumed authorization to another privileged account.
	session.SuperAdminID = admin.ID
	if err := createSession(ctx, tx, session); err != nil {
		return fmt.Errorf("create session after Super Admin recovery: %w", err)
	}

	oldValues, err := recoveryAuditObject(map[string]any{
		"recovery_authorization_id": authorization.ID,
		"status":                    string(RecoveryAuthorizationStateActive),
	})
	if err != nil {
		return err
	}
	newValues, err := recoveryAuditObject(map[string]any{
		"authorization_correlation_id": authorization.CorrelationID,
		"operator_identifier":          authorization.OperatorIdentifier,
		"recovery_authorization_id":    authorization.ID,
		"status":                       string(RecoveryAuthorizationStateConsumed),
	})
	if err != nil {
		return err
	}
	if err := audit.Append(ctx, tx, audit.Event{
		OccurredAt:        consumedAt,
		ActorType:         audit.ActorTypeSuperAdmin,
		ActorIdentifier:   admin.Email,
		Action:            audit.ActionSuperAdminRecoveryConsumed,
		ResourceType:      audit.ResourceTypeSuperAdminAccount,
		ResourceReference: admin.Email,
		OldValues:         oldValues,
		NewValues:         newValues,
		Reason:            stringPointer(authorization.Reason),
		Result:            audit.ResultSuccess,
		CorrelationID:     correlationID,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit Super Admin recovery consumption: %w", err)
	}

	return nil
}

func (store *PostgresStore) AuthorizeSuperAdminRecovery(
	ctx context.Context,
	normalizedEmail string,
	reason string,
	operatorIdentifier string,
	correlationID string,
	createdAt time.Time,
	expiresAt time.Time,
) (RecoveryAuthorization, error) {
	validity := expiresAt.Sub(createdAt)
	if validity <= 0 || validity > DefaultRecoveryAuthorizationValidity {
		return RecoveryAuthorization{}, ErrRecoveryUnavailable
	}

	tx, err := store.db.Begin(ctx)
	if err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("begin Super Admin recovery authorization: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	admin, err := lockRecoveryAdminByEmail(ctx, tx, normalizedEmail)
	operationAt, timeErr := recoveryDatabaseTime(ctx, tx)
	if timeErr != nil {
		return RecoveryAuthorization{}, timeErr
	}
	createdAt = operationAt
	expiresAt = operationAt.Add(validity)
	if errors.Is(err, ErrSuperAdminNotFound) || (err == nil && !admin.IsActive) {
		auditErr := appendRejectedRecoveryAudit(
			ctx,
			tx,
			audit.ActionSuperAdminRecoveryAuthorizationFailed,
			normalizedEmail,
			reason,
			operatorIdentifier,
			correlationID,
			createdAt,
			"TARGET_NOT_ELIGIBLE",
		)
		if auditErr != nil {
			return RecoveryAuthorization{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return RecoveryAuthorization{}, fmt.Errorf("commit rejected Super Admin recovery authorization: %w", err)
		}
		return RecoveryAuthorization{}, ErrRecoveryTargetNotEligible
	}
	if err != nil {
		return RecoveryAuthorization{}, err
	}

	if err := expireElapsedRecoveries(ctx, tx, admin.ID, createdAt); err != nil {
		return RecoveryAuthorization{}, err
	}

	_, err = findUnresolvedRecoveryForUpdate(ctx, tx, admin.ID)
	if err == nil {
		auditErr := appendRejectedRecoveryAudit(
			ctx,
			tx,
			audit.ActionSuperAdminRecoveryAuthorizationFailed,
			admin.Email,
			reason,
			operatorIdentifier,
			correlationID,
			createdAt,
			"ACTIVE_AUTHORIZATION_EXISTS",
		)
		if auditErr != nil {
			return RecoveryAuthorization{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return RecoveryAuthorization{}, fmt.Errorf("commit duplicate Super Admin recovery authorization: %w", err)
		}
		return RecoveryAuthorization{}, ErrRecoveryAlreadyActive
	}
	if !errors.Is(err, ErrRecoveryNotActive) {
		return RecoveryAuthorization{}, err
	}

	const insertQuery = `
		INSERT INTO super_admin_recovery_authorizations (
			super_admin_id,
			created_at,
			expires_at,
			reason,
			operator_identifier,
			correlation_id
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id::text
	`
	authorization := RecoveryAuthorization{
		SuperAdminID:       admin.ID,
		SuperAdminEmail:    admin.Email,
		Reason:             reason,
		OperatorIdentifier: operatorIdentifier,
		CorrelationID:      correlationID,
		CreatedAt:          createdAt,
		ExpiresAt:          expiresAt,
	}
	if err := tx.QueryRow(
		ctx,
		insertQuery,
		admin.ID,
		createdAt,
		expiresAt,
		reason,
		operatorIdentifier,
		correlationID,
	).Scan(&authorization.ID); err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("create Super Admin recovery authorization: %w", err)
	}

	newValues, err := recoveryAuditObject(map[string]any{
		"authorization_correlation_id": correlationID,
		"expires_at":                   expiresAt,
		"operator_identifier":          operatorIdentifier,
		"recovery_authorization_id":    authorization.ID,
		"status":                       string(RecoveryAuthorizationStateActive),
	})
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	if err := audit.Append(ctx, tx, audit.Event{
		OccurredAt:        createdAt,
		ActorType:         audit.ActorTypeDeploymentOperator,
		ActorIdentifier:   operatorIdentifier,
		Action:            audit.ActionSuperAdminRecoveryAuthorized,
		ResourceType:      audit.ResourceTypeSuperAdminAccount,
		ResourceReference: admin.Email,
		NewValues:         newValues,
		Reason:            stringPointer(reason),
		Result:            audit.ResultSuccess,
		CorrelationID:     correlationID,
	}); err != nil {
		return RecoveryAuthorization{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("commit Super Admin recovery authorization: %w", err)
	}

	return authorization, nil
}

func (store *PostgresStore) FindSuperAdminRecoveryStatus(
	ctx context.Context,
	normalizedEmail string,
	_ time.Time,
) (RecoveryAuthorizationStatus, error) {
	const accountQuery = `
		SELECT id::text, email, is_active, clock_timestamp()
		FROM super_admin_accounts
		WHERE email = $1
	`
	var admin SuperAdmin
	var now time.Time
	if err := store.db.QueryRow(ctx, accountQuery, normalizedEmail).Scan(
		&admin.ID,
		&admin.Email,
		&admin.IsActive,
		&now,
	); errors.Is(err, pgx.ErrNoRows) {
		return RecoveryAuthorizationStatus{}, ErrRecoveryTargetNotEligible
	} else if err != nil {
		return RecoveryAuthorizationStatus{}, fmt.Errorf("find Super Admin recovery status target: %w", err)
	}
	if !admin.IsActive {
		return RecoveryAuthorizationStatus{}, ErrRecoveryTargetNotEligible
	}

	const query = `
		SELECT
			recovery.id::text,
			recovery.super_admin_id::text,
			account.email,
			recovery.reason,
			recovery.operator_identifier,
			recovery.correlation_id,
			recovery.created_at,
			recovery.expires_at,
			recovery.consumed_at,
			recovery.revoked_at,
			recovery.expired_at
		FROM super_admin_recovery_authorizations AS recovery
		JOIN super_admin_accounts AS account
		  ON account.id = recovery.super_admin_id
		WHERE recovery.super_admin_id = $1::uuid
		ORDER BY recovery.created_at DESC, recovery.id DESC
		LIMIT 1
	`
	authorization, err := scanRecoveryAuthorization(store.db.QueryRow(ctx, query, admin.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return RecoveryAuthorizationStatus{State: RecoveryAuthorizationStateNone}, nil
	}
	if err != nil {
		return RecoveryAuthorizationStatus{}, fmt.Errorf("find Super Admin recovery status: %w", err)
	}

	return RecoveryAuthorizationStatus{
		State:         authorization.StateAt(now),
		Authorization: authorization,
	}, nil
}

func (store *PostgresStore) RevokeSuperAdminRecovery(
	ctx context.Context,
	normalizedEmail string,
	reason string,
	operatorIdentifier string,
	correlationID string,
	_ time.Time,
) (RecoveryAuthorization, error) {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("begin Super Admin recovery revocation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	admin, err := lockRecoveryAdminByEmail(ctx, tx, normalizedEmail)
	revokedAt, timeErr := recoveryDatabaseTime(ctx, tx)
	if timeErr != nil {
		return RecoveryAuthorization{}, timeErr
	}
	if errors.Is(err, ErrSuperAdminNotFound) || (err == nil && !admin.IsActive) {
		auditErr := appendRejectedRecoveryAudit(
			ctx,
			tx,
			audit.ActionSuperAdminRecoveryRevocationFailed,
			normalizedEmail,
			reason,
			operatorIdentifier,
			correlationID,
			revokedAt,
			"TARGET_NOT_ELIGIBLE",
		)
		if auditErr != nil {
			return RecoveryAuthorization{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return RecoveryAuthorization{}, fmt.Errorf("commit rejected Super Admin recovery revocation: %w", err)
		}
		return RecoveryAuthorization{}, ErrRecoveryTargetNotEligible
	}
	if err != nil {
		return RecoveryAuthorization{}, err
	}

	if err := expireElapsedRecoveries(ctx, tx, admin.ID, revokedAt); err != nil {
		return RecoveryAuthorization{}, err
	}
	authorization, err := findActiveRecoveryForUpdate(ctx, tx, admin.ID, revokedAt)
	if errors.Is(err, ErrRecoveryNotActive) {
		auditErr := appendRejectedRecoveryAudit(
			ctx,
			tx,
			audit.ActionSuperAdminRecoveryRevocationFailed,
			admin.Email,
			reason,
			operatorIdentifier,
			correlationID,
			revokedAt,
			"NO_ACTIVE_AUTHORIZATION",
		)
		if auditErr != nil {
			return RecoveryAuthorization{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return RecoveryAuthorization{}, fmt.Errorf("commit inactive Super Admin recovery revocation: %w", err)
		}
		return RecoveryAuthorization{}, ErrRecoveryNotActive
	}
	if err != nil {
		return RecoveryAuthorization{}, err
	}

	const revokeQuery = `
		UPDATE super_admin_recovery_authorizations
		SET revoked_at = $2
		WHERE id = $1::uuid
		  AND consumed_at IS NULL
		  AND revoked_at IS NULL
		  AND expired_at IS NULL
		  AND expires_at > $2
	`
	result, err := tx.Exec(ctx, revokeQuery, authorization.ID, revokedAt)
	if err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("revoke Super Admin recovery authorization: %w", err)
	}
	if result.RowsAffected() != 1 {
		return RecoveryAuthorization{}, ErrRecoveryNotActive
	}
	authorization.RevokedAt = timePointer(revokedAt)

	oldValues, err := recoveryAuditObject(map[string]any{
		"recovery_authorization_id": authorization.ID,
		"status":                    string(RecoveryAuthorizationStateActive),
	})
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	newValues, err := recoveryAuditObject(map[string]any{
		"authorization_correlation_id": authorization.CorrelationID,
		"authorization_operator":       authorization.OperatorIdentifier,
		"recovery_authorization_id":    authorization.ID,
		"revoked_by_operator":          operatorIdentifier,
		"status":                       string(RecoveryAuthorizationStateRevoked),
	})
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	if err := audit.Append(ctx, tx, audit.Event{
		OccurredAt:        revokedAt,
		ActorType:         audit.ActorTypeDeploymentOperator,
		ActorIdentifier:   operatorIdentifier,
		Action:            audit.ActionSuperAdminRecoveryRevoked,
		ResourceType:      audit.ResourceTypeSuperAdminAccount,
		ResourceReference: admin.Email,
		OldValues:         oldValues,
		NewValues:         newValues,
		Reason:            stringPointer(reason),
		Result:            audit.ResultSuccess,
		CorrelationID:     correlationID,
	}); err != nil {
		return RecoveryAuthorization{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("commit Super Admin recovery revocation: %w", err)
	}

	return authorization, nil
}

func lockRecoveryAdminByEmail(
	ctx context.Context,
	tx pgx.Tx,
	normalizedEmail string,
) (SuperAdmin, error) {
	const query = `
		SELECT id::text, email, is_active
		FROM super_admin_accounts
		WHERE email = $1
		FOR UPDATE
	`
	var admin SuperAdmin
	if err := tx.QueryRow(ctx, query, normalizedEmail).Scan(
		&admin.ID,
		&admin.Email,
		&admin.IsActive,
	); errors.Is(err, pgx.ErrNoRows) {
		return SuperAdmin{}, ErrSuperAdminNotFound
	} else if err != nil {
		return SuperAdmin{}, fmt.Errorf("lock Super Admin recovery target: %w", err)
	}

	return admin, nil
}

func lockRecoveryAdminByID(
	ctx context.Context,
	tx pgx.Tx,
	superAdminID string,
) (SuperAdmin, error) {
	const query = `
		SELECT id::text, email, is_active
		FROM super_admin_accounts
		WHERE id = $1::uuid
		FOR UPDATE
	`
	var admin SuperAdmin
	if err := tx.QueryRow(ctx, query, superAdminID).Scan(
		&admin.ID,
		&admin.Email,
		&admin.IsActive,
	); errors.Is(err, pgx.ErrNoRows) {
		return SuperAdmin{}, ErrRecoveryUnavailable
	} else if err != nil {
		return SuperAdmin{}, fmt.Errorf("lock Super Admin for recovery consumption: %w", err)
	}

	return admin, nil
}

func expireElapsedRecoveries(
	ctx context.Context,
	tx pgx.Tx,
	superAdminID string,
	now time.Time,
) error {
	const query = `
		UPDATE super_admin_recovery_authorizations
		SET expired_at = GREATEST($2, expires_at)
		WHERE super_admin_id = $1::uuid
		  AND consumed_at IS NULL
		  AND revoked_at IS NULL
		  AND expired_at IS NULL
		  AND expires_at <= $2
	`
	if _, err := tx.Exec(ctx, query, superAdminID, now); err != nil {
		return fmt.Errorf("terminalize expired Super Admin recovery authorization: %w", err)
	}

	return nil
}

// recoveryDatabaseTime is evaluated only after the account lock is acquired.
// clock_timestamp(), unlike CURRENT_TIMESTAMP, cannot remain frozen at the
// transaction start while a request waits behind another account operation.
func recoveryDatabaseTime(ctx context.Context, tx pgx.Tx) (time.Time, error) {
	var operationAt time.Time
	if err := tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&operationAt); err != nil {
		return time.Time{}, fmt.Errorf("read PostgreSQL recovery operation time: %w", err)
	}

	return operationAt.UTC(), nil
}

func findUnresolvedRecoveryForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	superAdminID string,
) (RecoveryAuthorization, error) {
	const query = `
		SELECT
			recovery.id::text,
			recovery.super_admin_id::text,
			account.email,
			recovery.reason,
			recovery.operator_identifier,
			recovery.correlation_id,
			recovery.created_at,
			recovery.expires_at,
			recovery.consumed_at,
			recovery.revoked_at,
			recovery.expired_at
		FROM super_admin_recovery_authorizations AS recovery
		JOIN super_admin_accounts AS account
		  ON account.id = recovery.super_admin_id
		WHERE recovery.super_admin_id = $1::uuid
		  AND recovery.consumed_at IS NULL
		  AND recovery.revoked_at IS NULL
		  AND recovery.expired_at IS NULL
		FOR UPDATE OF recovery
	`
	authorization, err := scanRecoveryAuthorization(tx.QueryRow(ctx, query, superAdminID))
	if errors.Is(err, pgx.ErrNoRows) {
		return RecoveryAuthorization{}, ErrRecoveryNotActive
	}
	if err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("find unresolved Super Admin recovery authorization: %w", err)
	}

	return authorization, nil
}

func findActiveRecoveryForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	superAdminID string,
	now time.Time,
) (RecoveryAuthorization, error) {
	const query = `
		SELECT
			recovery.id::text,
			recovery.super_admin_id::text,
			account.email,
			recovery.reason,
			recovery.operator_identifier,
			recovery.correlation_id,
			recovery.created_at,
			recovery.expires_at,
			recovery.consumed_at,
			recovery.revoked_at,
			recovery.expired_at
		FROM super_admin_recovery_authorizations AS recovery
		JOIN super_admin_accounts AS account
		  ON account.id = recovery.super_admin_id
		WHERE recovery.super_admin_id = $1::uuid
		  AND recovery.consumed_at IS NULL
		  AND recovery.revoked_at IS NULL
		  AND recovery.expired_at IS NULL
		  AND recovery.expires_at > $2
		FOR UPDATE OF recovery
	`
	authorization, err := scanRecoveryAuthorization(tx.QueryRow(ctx, query, superAdminID, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return RecoveryAuthorization{}, ErrRecoveryNotActive
	}
	if err != nil {
		return RecoveryAuthorization{}, fmt.Errorf("find active Super Admin recovery authorization: %w", err)
	}

	return authorization, nil
}

func scanRecoveryAuthorization(row pgx.Row) (RecoveryAuthorization, error) {
	var authorization RecoveryAuthorization
	err := row.Scan(
		&authorization.ID,
		&authorization.SuperAdminID,
		&authorization.SuperAdminEmail,
		&authorization.Reason,
		&authorization.OperatorIdentifier,
		&authorization.CorrelationID,
		&authorization.CreatedAt,
		&authorization.ExpiresAt,
		&authorization.ConsumedAt,
		&authorization.RevokedAt,
		&authorization.ExpiredAt,
	)
	return authorization, err
}

func appendRejectedRecoveryAudit(
	ctx context.Context,
	tx pgx.Tx,
	action audit.Action,
	resourceReference string,
	reason string,
	operatorIdentifier string,
	correlationID string,
	occurredAt time.Time,
	failure string,
) error {
	newValues, err := recoveryAuditObject(map[string]any{
		"failure":             failure,
		"operator_identifier": operatorIdentifier,
	})
	if err != nil {
		return err
	}

	return audit.Append(ctx, tx, audit.Event{
		OccurredAt:        occurredAt,
		ActorType:         audit.ActorTypeDeploymentOperator,
		ActorIdentifier:   operatorIdentifier,
		Action:            action,
		ResourceType:      audit.ResourceTypeSuperAdminAccount,
		ResourceReference: resourceReference,
		NewValues:         newValues,
		Reason:            stringPointer(reason),
		Result:            audit.ResultFailure,
		CorrelationID:     correlationID,
	})
}

func recoveryAuditObject(value map[string]any) (json.RawMessage, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, errors.New("encode safe Super Admin recovery audit values")
	}

	return json.RawMessage(encoded), nil
}

func stringPointer(value string) *string {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}
