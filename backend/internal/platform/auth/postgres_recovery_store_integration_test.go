package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/platform/audit"
	"salesagent.local/backend/internal/requestmeta"
)

func TestPostgresRecoveryAuthorizationLifecycleAndAudit(t *testing.T) {
	ctx, db, store := openRecoveryIntegrationPostgres(t)
	admin, rawPassword, passwordHash := createRecoveryIntegrationAdmin(t, ctx, db, store, "recovery-lifecycle")

	// The domain clock supplies a duration, but PostgreSQL supplies the absolute
	// operation time after locking the target. A deliberately stale caller time
	// proves the persisted window cannot be backdated by an application host.
	callerTime := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	reason := "notification provider outage"
	operator := "recovery-integration-operator"
	authorizeCorrelation := "recovery-integration-authorize"
	service := newRecoveryIntegrationService(t, store, callerTime)
	authorizeStarted := recoveryIntegrationDatabaseTime(t, ctx, db)
	authorization, err := service.Authorize(
		requestmeta.WithCorrelationID(ctx, authorizeCorrelation),
		strings.ToUpper(admin.Email),
		reason,
		operator,
	)
	if err != nil {
		t.Fatalf("authorize recovery: %v", err)
	}
	authorizeFinished := recoveryIntegrationDatabaseTime(t, ctx, db)
	if authorization.ID == "" ||
		authorization.SuperAdminID != admin.ID ||
		authorization.SuperAdminEmail != admin.Email ||
		authorization.Reason != reason ||
		authorization.OperatorIdentifier != operator ||
		authorization.CorrelationID != authorizeCorrelation ||
		authorization.CreatedAt.Before(authorizeStarted) ||
		authorization.CreatedAt.After(authorizeFinished) ||
		authorization.ExpiresAt.Sub(authorization.CreatedAt) != DefaultRecoveryAuthorizationValidity {
		t.Fatalf("authorized recovery = %#v", authorization)
	}

	var persistedCreatedAt time.Time
	var persistedExpiresAt time.Time
	var persistedReason string
	var persistedOperator string
	var persistedCorrelation string
	if err := db.QueryRow(
		ctx,
		`SELECT created_at, expires_at, reason, operator_identifier, correlation_id
		 FROM super_admin_recovery_authorizations
		 WHERE id = $1::uuid`,
		authorization.ID,
	).Scan(
		&persistedCreatedAt,
		&persistedExpiresAt,
		&persistedReason,
		&persistedOperator,
		&persistedCorrelation,
	); err != nil {
		t.Fatalf("read authorized recovery: %v", err)
	}
	if !persistedCreatedAt.Equal(authorization.CreatedAt) ||
		!persistedExpiresAt.Equal(authorization.ExpiresAt) ||
		persistedExpiresAt.Sub(persistedCreatedAt) != 10*time.Minute ||
		persistedReason != reason ||
		persistedOperator != operator ||
		persistedCorrelation != authorizeCorrelation {
		t.Fatalf(
			"persisted recovery = created:%s expires:%s reason:%q operator:%q correlation:%q",
			persistedCreatedAt,
			persistedExpiresAt,
			persistedReason,
			persistedOperator,
			persistedCorrelation,
		)
	}

	events := recoveryIntegrationAuditEvents(t, ctx, db, admin.Email)
	authorizedEvent := recoveryIntegrationAuditEvent(
		t,
		events,
		audit.ActionSuperAdminRecoveryAuthorized,
		authorizeCorrelation,
	)
	assertRecoveryIntegrationAuditAttribution(
		t,
		authorizedEvent,
		audit.ActorTypeDeploymentOperator,
		operator,
		reason,
		audit.ResultSuccess,
	)
	authorizedValues := recoveryIntegrationAuditValues(t, authorizedEvent.NewValues)
	if authorizedValues["operator_identifier"] != operator ||
		authorizedValues["recovery_authorization_id"] != authorization.ID ||
		authorizedValues["status"] != string(RecoveryAuthorizationStateActive) {
		t.Fatalf("authorization audit values = %#v", authorizedValues)
	}

	duplicateCorrelation := "recovery-integration-duplicate"
	if _, err := service.Authorize(
		requestmeta.WithCorrelationID(ctx, duplicateCorrelation),
		admin.Email,
		"duplicate outage authorization",
		operator,
	); !errors.Is(err, ErrRecoveryAlreadyActive) {
		t.Fatalf("duplicate authorization error = %v, want %v", err, ErrRecoveryAlreadyActive)
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_recovery_authorizations WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 1 {
		t.Fatalf("authorization rows after duplicate = %d, want 1", got)
	}
	events = recoveryIntegrationAuditEvents(t, ctx, db, admin.Email)
	duplicateEvent := recoveryIntegrationAuditEvent(
		t,
		events,
		audit.ActionSuperAdminRecoveryAuthorizationFailed,
		duplicateCorrelation,
	)
	assertRecoveryIntegrationAuditAttribution(
		t,
		duplicateEvent,
		audit.ActorTypeDeploymentOperator,
		operator,
		"duplicate outage authorization",
		audit.ResultFailure,
	)
	duplicateValues := recoveryIntegrationAuditValues(t, duplicateEvent.NewValues)
	if duplicateValues["failure"] != "ACTIVE_AUTHORIZATION_EXISTS" {
		t.Fatalf("duplicate failure audit values = %#v", duplicateValues)
	}

	active, err := store.HasActiveSuperAdminRecovery(ctx, admin.ID, callerTime)
	if err != nil {
		t.Fatalf("check active recovery: %v", err)
	}
	if !active {
		t.Fatal("fresh authorization is not active")
	}

	consumedAtArgument := callerTime.Add(time.Second)
	session, rawSessionToken := recoveryIntegrationSession(admin.ID, time.Now().UTC(), "successful")
	consumeCorrelation := "recovery-integration-consume"
	if err := store.ConsumeSuperAdminRecoveryAndCreateSession(
		ctx,
		admin.ID,
		consumedAtArgument,
		session,
		consumeCorrelation,
	); err != nil {
		t.Fatalf("consume recovery: %v", err)
	}

	var persistedTokenHash []byte
	var persistedCSRFToken string
	if err := db.QueryRow(
		ctx,
		`SELECT token_hash, csrf_token
		 FROM super_admin_sessions
		 WHERE super_admin_id = $1::uuid`,
		admin.ID,
	).Scan(&persistedTokenHash, &persistedCSRFToken); err != nil {
		t.Fatalf("read recovery-created session: %v", err)
	}
	if len(persistedTokenHash) != sha256.Size ||
		!bytes.Equal(persistedTokenHash, session.TokenHash) ||
		bytes.Equal(persistedTokenHash, []byte(rawSessionToken)) ||
		persistedCSRFToken != session.CSRFToken {
		t.Fatalf("persisted recovery session material = hash length:%d csrf match:%v", len(persistedTokenHash), persistedCSRFToken == session.CSRFToken)
	}
	loadedSession, err := store.FindSessionByTokenHash(ctx, session.TokenHash)
	if err != nil {
		t.Fatalf("load recovery session through normal session mechanism: %v", err)
	}
	if loadedSession.SuperAdmin.ID != admin.ID || loadedSession.SuperAdmin.Email != admin.Email {
		t.Fatalf("recovery session identity = %#v", loadedSession.SuperAdmin)
	}
	status, err := store.FindSuperAdminRecoveryStatus(ctx, admin.Email, callerTime)
	if err != nil {
		t.Fatalf("find consumed recovery status: %v", err)
	}
	if status.State != RecoveryAuthorizationStateConsumed || status.Authorization.ID != authorization.ID {
		t.Fatalf("consumed recovery status = %#v", status)
	}

	events = recoveryIntegrationAuditEvents(t, ctx, db, admin.Email)
	consumedEvent := recoveryIntegrationAuditEvent(
		t,
		events,
		audit.ActionSuperAdminRecoveryConsumed,
		consumeCorrelation,
	)
	assertRecoveryIntegrationAuditAttribution(
		t,
		consumedEvent,
		audit.ActorTypeSuperAdmin,
		admin.Email,
		reason,
		audit.ResultSuccess,
	)
	consumedValues := recoveryIntegrationAuditValues(t, consumedEvent.NewValues)
	if consumedValues["operator_identifier"] != operator ||
		consumedValues["recovery_authorization_id"] != authorization.ID ||
		consumedValues["status"] != string(RecoveryAuthorizationStateConsumed) {
		t.Fatalf("consumption audit values = %#v", consumedValues)
	}

	secondSession, secondRawSessionToken := recoveryIntegrationSession(
		admin.ID,
		time.Now().UTC(),
		"second",
	)
	if err := store.ConsumeSuperAdminRecoveryAndCreateSession(
		ctx,
		admin.ID,
		callerTime,
		secondSession,
		"recovery-integration-second-consume",
	); !errors.Is(err, ErrRecoveryNotActive) {
		t.Fatalf("second recovery consumption error = %v, want %v", err, ErrRecoveryNotActive)
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 1 {
		t.Fatalf("sessions after second consumption = %d, want 1", got)
	}
	if got := recoveryIntegrationAuditActionCount(
		t,
		ctx,
		db,
		admin.Email,
		audit.ActionSuperAdminRecoveryConsumed,
	); got != 1 {
		t.Fatalf("consumed audit events after second consumption = %d, want 1", got)
	}

	// A genuinely elapsed short grant is unusable. Creating the next grant
	// terminalizes it and proves expired authorizations do not permanently block
	// a fresh deployment operation.
	shortCallerTime := time.Date(2002, 3, 4, 5, 6, 7, 0, time.UTC)
	shortValidity := 250 * time.Millisecond
	shortAuthorization, err := store.AuthorizeSuperAdminRecovery(
		ctx,
		admin.Email,
		"short elapsed outage authorization",
		"recovery-integration-short-operator",
		"recovery-integration-short",
		shortCallerTime,
		shortCallerTime.Add(shortValidity),
	)
	if err != nil {
		t.Fatalf("authorize short recovery: %v", err)
	}
	if shortAuthorization.ExpiresAt.Sub(shortAuthorization.CreatedAt) != shortValidity {
		t.Fatalf("short recovery validity = %s, want %s", shortAuthorization.ExpiresAt.Sub(shortAuthorization.CreatedAt), shortValidity)
	}
	waitForRecoveryIntegrationDatabaseTime(t, ctx, db, shortAuthorization.ExpiresAt)
	active, err = store.HasActiveSuperAdminRecovery(ctx, admin.ID, callerTime)
	if err != nil {
		t.Fatalf("check elapsed short recovery: %v", err)
	}
	if active {
		t.Fatal("elapsed short recovery remained active")
	}
	elapsedSession, elapsedRawToken := recoveryIntegrationSession(admin.ID, time.Now().UTC(), "elapsed")
	if err := store.ConsumeSuperAdminRecoveryAndCreateSession(
		ctx,
		admin.ID,
		shortCallerTime,
		elapsedSession,
		"recovery-integration-elapsed-consume",
	); !errors.Is(err, ErrRecoveryNotActive) {
		t.Fatalf("elapsed recovery consumption error = %v, want %v", err, ErrRecoveryNotActive)
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 1 {
		t.Fatalf("sessions after elapsed consumption = %d, want 1", got)
	}

	replacementReason := "provider outage continues"
	replacementOperator := "recovery-integration-operator-2"
	replacementCorrelation := "recovery-integration-replacement"
	replacementService := newRecoveryIntegrationService(t, store, callerTime)
	replacementStarted := recoveryIntegrationDatabaseTime(t, ctx, db)
	replacement, err := replacementService.Authorize(
		requestmeta.WithCorrelationID(ctx, replacementCorrelation),
		admin.Email,
		replacementReason,
		replacementOperator,
	)
	if err != nil {
		t.Fatalf("authorize replacement recovery: %v", err)
	}
	replacementFinished := recoveryIntegrationDatabaseTime(t, ctx, db)
	if replacement.ID == shortAuthorization.ID ||
		replacement.CreatedAt.Before(replacementStarted) ||
		replacement.CreatedAt.After(replacementFinished) ||
		replacement.ExpiresAt.Sub(replacement.CreatedAt) != DefaultRecoveryAuthorizationValidity {
		t.Fatalf("replacement recovery = %#v", replacement)
	}
	var expiredAt *time.Time
	if err := db.QueryRow(
		ctx,
		`SELECT expired_at FROM super_admin_recovery_authorizations WHERE id = $1::uuid`,
		shortAuthorization.ID,
	).Scan(&expiredAt); err != nil {
		t.Fatalf("read terminalized elapsed recovery: %v", err)
	}
	if expiredAt == nil || expiredAt.Before(shortAuthorization.ExpiresAt) {
		t.Fatalf("elapsed recovery expired_at = %v, expiry was %s", expiredAt, shortAuthorization.ExpiresAt)
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_recovery_authorizations WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 3 {
		t.Fatalf("authorization rows after replacement = %d, want 3", got)
	}
	status, err = store.FindSuperAdminRecoveryStatus(ctx, admin.Email, callerTime)
	if err != nil {
		t.Fatalf("find replacement status: %v", err)
	}
	if status.State != RecoveryAuthorizationStateActive || status.Authorization.ID != replacement.ID {
		t.Fatalf("replacement status = %#v", status)
	}

	revokeReason := "incident resolved before recovery was used"
	revoker := "recovery-integration-revoker"
	revokeCorrelation := "recovery-integration-revoke"
	revokeService := newRecoveryIntegrationService(t, store, callerTime)
	revokeStarted := recoveryIntegrationDatabaseTime(t, ctx, db)
	revoked, err := revokeService.Revoke(
		requestmeta.WithCorrelationID(ctx, revokeCorrelation),
		admin.Email,
		revokeReason,
		revoker,
	)
	if err != nil {
		t.Fatalf("revoke recovery: %v", err)
	}
	revokeFinished := recoveryIntegrationDatabaseTime(t, ctx, db)
	if revoked.ID != replacement.ID ||
		revoked.RevokedAt == nil ||
		revoked.RevokedAt.Before(revokeStarted) ||
		revoked.RevokedAt.After(revokeFinished) {
		t.Fatalf("revoked recovery = %#v", revoked)
	}
	active, err = store.HasActiveSuperAdminRecovery(ctx, admin.ID, callerTime)
	if err != nil {
		t.Fatalf("check revoked recovery: %v", err)
	}
	if active {
		t.Fatal("revoked recovery remained active")
	}
	status, err = store.FindSuperAdminRecoveryStatus(ctx, admin.Email, callerTime)
	if err != nil {
		t.Fatalf("find revoked recovery status: %v", err)
	}
	if status.State != RecoveryAuthorizationStateRevoked || status.Authorization.ID != replacement.ID {
		t.Fatalf("revoked recovery status = %#v", status)
	}
	events = recoveryIntegrationAuditEvents(t, ctx, db, admin.Email)
	revokedEvent := recoveryIntegrationAuditEvent(
		t,
		events,
		audit.ActionSuperAdminRecoveryRevoked,
		revokeCorrelation,
	)
	assertRecoveryIntegrationAuditAttribution(
		t,
		revokedEvent,
		audit.ActorTypeDeploymentOperator,
		revoker,
		revokeReason,
		audit.ResultSuccess,
	)
	revokedValues := recoveryIntegrationAuditValues(t, revokedEvent.NewValues)
	if revokedValues["recovery_authorization_id"] != replacement.ID ||
		revokedValues["status"] != string(RecoveryAuthorizationStateRevoked) ||
		revokedValues["revoked_by_operator"] != revoker {
		t.Fatalf("revocation audit values = %#v", revokedValues)
	}

	assertRecoveryIntegrationAuditExcludes(
		t,
		events,
		rawPassword,
		passwordHash,
		elapsedRawToken,
		elapsedSession.CSRFToken,
		hex.EncodeToString(elapsedSession.TokenHash),
		rawSessionToken,
		session.CSRFToken,
		hex.EncodeToString(session.TokenHash),
		secondRawSessionToken,
		secondSession.CSRFToken,
		hex.EncodeToString(secondSession.TokenHash),
	)
}

func TestPostgresRecoveryUnknownTargetFailureIsAudited(t *testing.T) {
	ctx, db, store := openRecoveryIntegrationPostgres(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	email := fmt.Sprintf("recovery-unknown-%d@example.com", time.Now().UnixNano())
	reason := "notification provider outage"
	operator := "recovery-integration-unknown-operator"
	correlation := "recovery-integration-unknown"
	service := newRecoveryIntegrationService(t, store, now)
	if _, err := service.Authorize(
		requestmeta.WithCorrelationID(ctx, correlation),
		email,
		reason,
		operator,
	); !errors.Is(err, ErrRecoveryTargetNotEligible) {
		t.Fatalf("unknown target authorization error = %v, want %v", err, ErrRecoveryTargetNotEligible)
	}

	events := recoveryIntegrationAuditEvents(t, ctx, db, email)
	if len(events) != 1 {
		t.Fatalf("unknown-target audit event count = %d, want 1", len(events))
	}
	failedEvent := recoveryIntegrationAuditEvent(
		t,
		events,
		audit.ActionSuperAdminRecoveryAuthorizationFailed,
		correlation,
	)
	assertRecoveryIntegrationAuditAttribution(
		t,
		failedEvent,
		audit.ActorTypeDeploymentOperator,
		operator,
		reason,
		audit.ResultFailure,
	)
	failedValues := recoveryIntegrationAuditValues(t, failedEvent.NewValues)
	if failedValues["failure"] != "TARGET_NOT_ELIGIBLE" || failedValues["operator_identifier"] != operator {
		t.Fatalf("unknown-target audit values = %#v", failedValues)
	}
}

func TestPostgresRecoveryConsumptionRollsBackWhenSessionInsertFails(t *testing.T) {
	ctx, db, store := openRecoveryIntegrationPostgres(t)
	admin, _, _ := createRecoveryIntegrationAdmin(t, ctx, db, store, "recovery-rollback")

	service := newRecoveryIntegrationService(t, store, time.Date(2003, 4, 5, 6, 7, 8, 0, time.UTC))
	authorization, err := service.Authorize(
		requestmeta.WithCorrelationID(ctx, "recovery-integration-rollback-authorize"),
		admin.Email,
		"notification provider rollback test",
		"recovery-integration-rollback-operator",
	)
	if err != nil {
		t.Fatalf("authorize rollback recovery: %v", err)
	}

	sessionTime := time.Now().UTC()
	invalidSession := NewSession{
		SuperAdminID: admin.ID,
		TokenHash:    []byte("not-a-sha256-digest"),
		CSRFToken:    "too-short",
		CreatedAt:    sessionTime,
		ExpiresAt:    sessionTime.Add(time.Hour),
		LastSeenAt:   sessionTime,
	}
	consumeErr := store.ConsumeSuperAdminRecoveryAndCreateSession(
		ctx,
		admin.ID,
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		invalidSession,
		"recovery-integration-rollback-consume",
	)
	if consumeErr == nil || errors.Is(consumeErr, ErrRecoveryNotActive) {
		t.Fatalf("invalid session consumption error = %v, want session insert failure", consumeErr)
	}

	var consumedAt *time.Time
	var revokedAt *time.Time
	var expiredAt *time.Time
	if err := db.QueryRow(
		ctx,
		`SELECT consumed_at, revoked_at, expired_at
		 FROM super_admin_recovery_authorizations
		 WHERE id = $1::uuid`,
		authorization.ID,
	).Scan(&consumedAt, &revokedAt, &expiredAt); err != nil {
		t.Fatalf("read recovery after session rollback: %v", err)
	}
	if consumedAt != nil || revokedAt != nil || expiredAt != nil {
		t.Fatalf("failed session insert changed recovery state: consumed=%v revoked=%v expired=%v", consumedAt, revokedAt, expiredAt)
	}
	active, err := store.HasActiveSuperAdminRecovery(ctx, admin.ID, time.Time{})
	if err != nil {
		t.Fatalf("check recovery after session rollback: %v", err)
	}
	if !active {
		t.Fatal("failed session insert consumed or disabled the recovery authorization")
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 0 {
		t.Fatalf("sessions after transaction rollback = %d, want 0", got)
	}
	if got := recoveryIntegrationAuditActionCount(
		t,
		ctx,
		db,
		admin.Email,
		audit.ActionSuperAdminRecoveryConsumed,
	); got != 0 {
		t.Fatalf("consumption audit events after transaction rollback = %d, want 0", got)
	}
}

func TestPostgresRecoveryUsesDatabaseTimeAfterWaitingForAccountLock(t *testing.T) {
	ctx, db, store := openRecoveryIntegrationPostgres(t)
	admin, _, _ := createRecoveryIntegrationAdmin(t, ctx, db, store, "recovery-lock-expiry")

	callerCreatedAt := time.Date(2004, 5, 6, 7, 8, 9, 0, time.UTC)
	shortValidity := 2 * time.Second
	authorization, err := store.AuthorizeSuperAdminRecovery(
		ctx,
		admin.Email,
		"notification provider lock-expiry test",
		"recovery-integration-lock-operator",
		"recovery-integration-lock-authorize",
		callerCreatedAt,
		callerCreatedAt.Add(shortValidity),
	)
	if err != nil {
		t.Fatalf("authorize lock-expiry recovery: %v", err)
	}
	if authorization.ExpiresAt.Sub(authorization.CreatedAt) != shortValidity {
		t.Fatalf("lock-expiry validity = %s, want %s", authorization.ExpiresAt.Sub(authorization.CreatedAt), shortValidity)
	}

	lockTx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("begin account-lock transaction: %v", err)
	}
	defer func() { _ = lockTx.Rollback(ctx) }()
	var lockedID string
	var lockedAt time.Time
	if err := lockTx.QueryRow(
		ctx,
		`SELECT id::text, clock_timestamp()
		 FROM super_admin_accounts
		 WHERE id = $1::uuid
		 FOR UPDATE`,
		admin.ID,
	).Scan(&lockedID, &lockedAt); err != nil {
		t.Fatalf("lock recovery account: %v", err)
	}
	if lockedID != admin.ID || !lockedAt.Before(authorization.ExpiresAt) {
		t.Fatalf("account lock was not acquired before recovery expiry: id=%q locked_at=%s expires_at=%s", lockedID, lockedAt, authorization.ExpiresAt)
	}

	// This caller timestamp is valid before expiry. The store must ignore it
	// after waiting and instead read clock_timestamp() only once the account lock
	// becomes available.
	preExpiryCallerTime := authorization.CreatedAt.Add(shortValidity / 2)
	session, _ := recoveryIntegrationSession(admin.ID, time.Now().UTC(), "lock-expiry")
	consumeResult := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		close(started)
		consumeResult <- store.ConsumeSuperAdminRecoveryAndCreateSession(
			ctx,
			admin.ID,
			preExpiryCallerTime,
			session,
			"recovery-integration-lock-consume",
		)
	}()
	<-started
	waitForRecoveryIntegrationDatabaseTime(t, ctx, db, authorization.ExpiresAt)
	select {
	case premature := <-consumeResult:
		t.Fatalf("recovery consumption returned while account lock remained held: %v", premature)
	default:
	}
	if err := lockTx.Commit(ctx); err != nil {
		t.Fatalf("release account lock after recovery expiry: %v", err)
	}

	select {
	case consumeErr := <-consumeResult:
		if !errors.Is(consumeErr, ErrRecoveryNotActive) {
			t.Fatalf("post-lock elapsed consumption error = %v, want %v", consumeErr, ErrRecoveryNotActive)
		}
	case <-ctx.Done():
		t.Fatalf("post-lock elapsed consumption did not finish: %v", ctx.Err())
	}

	var consumedAt *time.Time
	if err := db.QueryRow(
		ctx,
		`SELECT consumed_at
		 FROM super_admin_recovery_authorizations
		 WHERE id = $1::uuid`,
		authorization.ID,
	).Scan(&consumedAt); err != nil {
		t.Fatalf("read lock-expired recovery: %v", err)
	}
	if consumedAt != nil {
		t.Fatalf("lock-expired recovery was consumed at %s", *consumedAt)
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 0 {
		t.Fatalf("lock-expired recovery sessions = %d, want 0", got)
	}
	if got := recoveryIntegrationAuditActionCount(
		t,
		ctx,
		db,
		admin.Email,
		audit.ActionSuperAdminRecoveryConsumed,
	); got != 0 {
		t.Fatalf("lock-expired recovery consumption audits = %d, want 0", got)
	}
}

func TestPostgresRecoveryCannotConsumeAfterAccountDeactivation(t *testing.T) {
	ctx, db, store := openRecoveryIntegrationPostgres(t)
	admin, _, _ := createRecoveryIntegrationAdmin(t, ctx, db, store, "recovery-inactive")

	service := newRecoveryIntegrationService(t, store, time.Date(2005, 6, 7, 8, 9, 10, 0, time.UTC))
	authorization, err := service.Authorize(
		requestmeta.WithCorrelationID(ctx, "recovery-integration-inactive-authorize"),
		admin.Email,
		"notification provider inactive-account test",
		"recovery-integration-inactive-operator",
	)
	if err != nil {
		t.Fatalf("authorize inactive-account recovery: %v", err)
	}
	if _, err := db.Exec(
		ctx,
		`UPDATE super_admin_accounts
		 SET is_active = FALSE,
		     updated_at = GREATEST(updated_at, clock_timestamp())
		 WHERE id = $1::uuid`,
		admin.ID,
	); err != nil {
		t.Fatalf("deactivate recovery account: %v", err)
	}

	session, _ := recoveryIntegrationSession(admin.ID, time.Now().UTC(), "inactive")
	if err := store.ConsumeSuperAdminRecoveryAndCreateSession(
		ctx,
		admin.ID,
		time.Date(1997, 1, 1, 0, 0, 0, 0, time.UTC),
		session,
		"recovery-integration-inactive-consume",
	); !errors.Is(err, ErrRecoveryUnavailable) {
		t.Fatalf("inactive account consumption error = %v, want %v", err, ErrRecoveryUnavailable)
	}

	var consumedAt *time.Time
	if err := db.QueryRow(
		ctx,
		`SELECT consumed_at
		 FROM super_admin_recovery_authorizations
		 WHERE id = $1::uuid`,
		authorization.ID,
	).Scan(&consumedAt); err != nil {
		t.Fatalf("read inactive-account recovery: %v", err)
	}
	if consumedAt != nil {
		t.Fatalf("inactive-account recovery was consumed at %s", *consumedAt)
	}
	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 0 {
		t.Fatalf("inactive-account recovery sessions = %d, want 0", got)
	}
	if got := recoveryIntegrationAuditActionCount(
		t,
		ctx,
		db,
		admin.Email,
		audit.ActionSuperAdminRecoveryConsumed,
	); got != 0 {
		t.Fatalf("inactive-account recovery consumption audits = %d, want 0", got)
	}
	active, err := store.HasActiveSuperAdminRecovery(ctx, admin.ID, time.Time{})
	if err != nil {
		t.Fatalf("check inactive-account recovery: %v", err)
	}
	if active {
		t.Fatal("recovery hint treated an inactive account as eligible")
	}
}

func TestPostgresRecoveryConcurrentConsumptionCreatesOneSessionAndAudit(t *testing.T) {
	ctx, db, store := openRecoveryIntegrationPostgres(t)
	admin, rawPassword, passwordHash := createRecoveryIntegrationAdmin(t, ctx, db, store, "recovery-concurrency")

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	reason := "notification provider concurrency outage"
	operator := "recovery-integration-concurrency-operator"
	service := newRecoveryIntegrationService(t, store, createdAt)
	authorization, err := service.Authorize(
		requestmeta.WithCorrelationID(ctx, "recovery-integration-concurrency-authorize"),
		admin.Email,
		reason,
		operator,
	)
	if err != nil {
		t.Fatalf("authorize concurrent recovery: %v", err)
	}

	consumedAtArgument := time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC)
	sessionCreatedAt := time.Now().UTC()
	type consumptionAttempt struct {
		correlation string
		rawToken    string
		session     NewSession
		err         error
	}
	attempts := make([]consumptionAttempt, 2)
	attempts[0].correlation = "recovery-integration-concurrent-a"
	attempts[0].session, attempts[0].rawToken = recoveryIntegrationSession(admin.ID, sessionCreatedAt, "concurrent-a")
	attempts[1].correlation = "recovery-integration-concurrent-b"
	attempts[1].session, attempts[1].rawToken = recoveryIntegrationSession(admin.ID, sessionCreatedAt, "concurrent-b")

	start := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(len(attempts))
	for index := range attempts {
		index := index
		go func() {
			defer wait.Done()
			<-start
			attempts[index].err = store.ConsumeSuperAdminRecoveryAndCreateSession(
				ctx,
				admin.ID,
				consumedAtArgument,
				attempts[index].session,
				attempts[index].correlation,
			)
		}()
	}
	consumeStarted := recoveryIntegrationDatabaseTime(t, ctx, db)
	close(start)
	wait.Wait()
	consumeFinished := recoveryIntegrationDatabaseTime(t, ctx, db)

	winner := -1
	notActive := 0
	for index := range attempts {
		switch {
		case attempts[index].err == nil:
			if winner != -1 {
				t.Fatalf("both concurrent recovery attempts succeeded: %#v", attempts)
			}
			winner = index
		case errors.Is(attempts[index].err, ErrRecoveryNotActive):
			notActive++
		default:
			t.Fatalf("concurrent recovery attempt %d error = %v", index, attempts[index].err)
		}
	}
	if winner == -1 || notActive != 1 {
		t.Fatalf("concurrent recovery results = winner:%d not-active:%d attempts:%#v", winner, notActive, attempts)
	}

	if got := recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		admin.ID,
	); got != 1 {
		t.Fatalf("concurrent recovery session count = %d, want 1", got)
	}
	var persistedTokenHash []byte
	var persistedCSRFToken string
	if err := db.QueryRow(
		ctx,
		`SELECT token_hash, csrf_token
		 FROM super_admin_sessions
		 WHERE super_admin_id = $1::uuid`,
		admin.ID,
	).Scan(&persistedTokenHash, &persistedCSRFToken); err != nil {
		t.Fatalf("read concurrent recovery session: %v", err)
	}
	if len(persistedTokenHash) != sha256.Size ||
		!bytes.Equal(persistedTokenHash, attempts[winner].session.TokenHash) ||
		persistedCSRFToken != attempts[winner].session.CSRFToken {
		t.Fatalf("concurrent recovery persisted the wrong session")
	}

	var consumedAtStored *time.Time
	if err := db.QueryRow(
		ctx,
		`SELECT consumed_at
		 FROM super_admin_recovery_authorizations
		 WHERE id = $1::uuid`,
		authorization.ID,
	).Scan(&consumedAtStored); err != nil {
		t.Fatalf("read concurrently consumed recovery: %v", err)
	}
	if consumedAtStored == nil || consumedAtStored.Before(consumeStarted) || consumedAtStored.After(consumeFinished) {
		t.Fatalf("concurrent recovery consumed_at = %v, expected between %s and %s", consumedAtStored, consumeStarted, consumeFinished)
	}
	active, err := store.HasActiveSuperAdminRecovery(ctx, admin.ID, consumedAtArgument)
	if err != nil {
		t.Fatalf("check concurrent recovery after consumption: %v", err)
	}
	if active {
		t.Fatal("concurrently consumed recovery remained active")
	}
	if got := recoveryIntegrationAuditActionCount(
		t,
		ctx,
		db,
		admin.Email,
		audit.ActionSuperAdminRecoveryConsumed,
	); got != 1 {
		t.Fatalf("concurrent consumed audit event count = %d, want 1", got)
	}
	events := recoveryIntegrationAuditEvents(t, ctx, db, admin.Email)
	consumedEvent := recoveryIntegrationAuditEvent(
		t,
		events,
		audit.ActionSuperAdminRecoveryConsumed,
		attempts[winner].correlation,
	)
	assertRecoveryIntegrationAuditAttribution(
		t,
		consumedEvent,
		audit.ActorTypeSuperAdmin,
		admin.Email,
		reason,
		audit.ResultSuccess,
	)
	consumedValues := recoveryIntegrationAuditValues(t, consumedEvent.NewValues)
	if consumedValues["operator_identifier"] != operator ||
		consumedValues["recovery_authorization_id"] != authorization.ID ||
		consumedValues["status"] != string(RecoveryAuthorizationStateConsumed) {
		t.Fatalf("concurrent consumption audit values = %#v", consumedValues)
	}

	assertRecoveryIntegrationAuditExcludes(
		t,
		events,
		rawPassword,
		passwordHash,
		attempts[0].rawToken,
		attempts[0].session.CSRFToken,
		hex.EncodeToString(attempts[0].session.TokenHash),
		attempts[1].rawToken,
		attempts[1].session.CSRFToken,
		hex.EncodeToString(attempts[1].session.TokenHash),
	)
}

func openRecoveryIntegrationPostgres(
	t *testing.T,
) (context.Context, *database.Postgres, *PostgresStore) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping Super Admin recovery PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	t.Cleanup(db.Close)
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	store, err := NewPostgresStore(db)
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}

	return ctx, db, store
}

func createRecoveryIntegrationAdmin(
	t *testing.T,
	ctx context.Context,
	db *database.Postgres,
	store *PostgresStore,
	prefix string,
) (SuperAdmin, string, string) {
	t.Helper()

	unique := time.Now().UnixNano()
	rawPassword := fmt.Sprintf("integration-recovery-password-%d", unique)
	passwordHash, err := NewPasswordHasher().Hash(rawPassword)
	if err != nil {
		t.Fatalf("hash recovery integration password: %v", err)
	}
	admin, err := store.CreateSuperAdmin(ctx, NewSuperAdmin{
		Email:        fmt.Sprintf("%s-%d@example.com", prefix, unique),
		PasswordHash: passwordHash,
		DisplayName:  "Recovery Integration Admin",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("create recovery integration Super Admin: %v", err)
	}
	t.Cleanup(func() {
		cleanupRecoveryIntegrationAdmin(t, db, admin.ID)
	})

	return admin, rawPassword, passwordHash
}

func cleanupRecoveryIntegrationAdmin(t *testing.T, db *database.Postgres, superAdminID string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Errorf("begin recovery integration cleanup: %v", err)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := []string{
		`DELETE FROM super_admin_auth_challenges WHERE super_admin_id = $1::uuid`,
		`DELETE FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		`DELETE FROM super_admin_recovery_authorizations WHERE super_admin_id = $1::uuid`,
		`DELETE FROM super_admin_accounts WHERE id = $1::uuid`,
	}
	for _, query := range queries {
		if _, err := tx.Exec(ctx, query, superAdminID); err != nil {
			t.Errorf("clean recovery integration data: %v", err)
			return
		}
	}
	// Platform Audit rows intentionally remain: the database makes them
	// immutable, and test cleanup must not weaken that contract.
	if err := tx.Commit(ctx); err != nil {
		t.Errorf("commit recovery integration cleanup: %v", err)
	}
}

func newRecoveryIntegrationService(
	t *testing.T,
	store *PostgresStore,
	now time.Time,
) *RecoveryService {
	t.Helper()

	service, err := NewRecoveryService(store, RecoveryServiceOptions{
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewRecoveryService() error = %v", err)
	}
	return service
}

func recoveryIntegrationSession(
	superAdminID string,
	createdAt time.Time,
	label string,
) (NewSession, string) {
	rawToken := "integration-raw-recovery-session-token-" + label + "-" + superAdminID
	tokenHash := sha256.Sum256([]byte(rawToken))
	csrfHash := sha256.Sum256([]byte("integration-recovery-csrf-" + label + "-" + superAdminID))
	return NewSession{
		SuperAdminID: superAdminID,
		TokenHash:    append([]byte(nil), tokenHash[:]...),
		CSRFToken:    base64.RawURLEncoding.EncodeToString(csrfHash[:]),
		CreatedAt:    createdAt,
		ExpiresAt:    createdAt.Add(8 * time.Hour),
		LastSeenAt:   createdAt,
	}, rawToken
}

func recoveryIntegrationAuditEvents(
	t *testing.T,
	ctx context.Context,
	db *database.Postgres,
	email string,
) []audit.Event {
	t.Helper()

	store, err := audit.NewPostgresStore(db)
	if err != nil {
		t.Fatalf("NewPostgresStore(audit) error = %v", err)
	}
	events, err := store.ListByResource(
		ctx,
		audit.ResourceTypeSuperAdminAccount,
		email,
		100,
	)
	if err != nil {
		t.Fatalf("list recovery audit events: %v", err)
	}
	return events
}

func recoveryIntegrationAuditEvent(
	t *testing.T,
	events []audit.Event,
	action audit.Action,
	correlationID string,
) audit.Event {
	t.Helper()

	var matched []audit.Event
	for _, event := range events {
		if event.Action == action && event.CorrelationID == correlationID {
			matched = append(matched, event)
		}
	}
	if len(matched) != 1 {
		t.Fatalf("audit events matching action %q and correlation %q = %d, want 1", action, correlationID, len(matched))
	}
	return matched[0]
}

func assertRecoveryIntegrationAuditAttribution(
	t *testing.T,
	event audit.Event,
	actorType audit.ActorType,
	actorIdentifier string,
	reason string,
	result audit.Result,
) {
	t.Helper()

	if event.ActorType != actorType ||
		event.ActorIdentifier != actorIdentifier ||
		event.ResourceType != audit.ResourceTypeSuperAdminAccount ||
		event.ResourceReference == "" ||
		event.Reason == nil || *event.Reason != reason ||
		event.Result != result ||
		event.OrganizationID != nil {
		t.Fatalf("recovery audit attribution = %#v", event)
	}
}

func recoveryIntegrationAuditValues(t *testing.T, encoded json.RawMessage) map[string]string {
	t.Helper()

	var values map[string]any
	if err := json.Unmarshal(encoded, &values); err != nil {
		t.Fatalf("decode recovery audit values: %v", err)
	}
	stringsOnly := make(map[string]string, len(values))
	for key, value := range values {
		if text, ok := value.(string); ok {
			stringsOnly[key] = text
		}
	}
	return stringsOnly
}

func assertRecoveryIntegrationAuditExcludes(
	t *testing.T,
	events []audit.Event,
	prohibitedValues ...string,
) {
	t.Helper()

	encoded, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("encode recovery audit events for secret scan: %v", err)
	}
	for _, value := range prohibitedValues {
		if value != "" && bytes.Contains(encoded, []byte(value)) {
			t.Fatalf("recovery audit contains prohibited credential material %q", value)
		}
	}
}

func recoveryIntegrationCount(
	t *testing.T,
	ctx context.Context,
	db *database.Postgres,
	query string,
	arguments ...any,
) int {
	t.Helper()

	var count int
	if err := db.QueryRow(ctx, query, arguments...).Scan(&count); err != nil {
		t.Fatalf("count recovery integration rows: %v", err)
	}
	return count
}

func recoveryIntegrationAuditActionCount(
	t *testing.T,
	ctx context.Context,
	db *database.Postgres,
	email string,
	action audit.Action,
) int {
	t.Helper()

	return recoveryIntegrationCount(
		t,
		ctx,
		db,
		`SELECT COUNT(*)
		 FROM platform_audit_events
		 WHERE resource_type = $1
		   AND resource_reference = $2
		   AND action = $3`,
		string(audit.ResourceTypeSuperAdminAccount),
		email,
		string(action),
	)
}

func waitForRecoveryIntegrationDatabaseTime(
	t *testing.T,
	ctx context.Context,
	db *database.Postgres,
	target time.Time,
) {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var databaseNow time.Time
		if err := db.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
			t.Fatalf("read PostgreSQL clock for recovery integration test: %v", err)
		}
		if !databaseNow.Before(target) {
			return
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatalf("wait for PostgreSQL recovery time: %v", ctx.Err())
		}
	}
}

func recoveryIntegrationDatabaseTime(
	t *testing.T,
	ctx context.Context,
	db *database.Postgres,
) time.Time {
	t.Helper()

	var databaseNow time.Time
	if err := db.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
		t.Fatalf("read PostgreSQL clock for recovery integration test: %v", err)
	}

	return databaseNow.UTC()
}
