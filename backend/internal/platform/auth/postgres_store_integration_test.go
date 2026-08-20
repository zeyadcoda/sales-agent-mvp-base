package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"salesagent.local/backend/internal/database"
)

func TestPostgresStoreAuthenticationRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping auth repository integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	t.Cleanup(db.Close)

	store, err := NewPostgresStore(db)
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}

	hasher := NewPasswordHasher()
	passwordHash, err := hasher.Hash("integration-only-password")
	if err != nil {
		t.Fatalf("hash integration password: %v", err)
	}

	email := fmt.Sprintf("auth-integration-%d@example.com", time.Now().UnixNano())
	created, err := store.CreateSuperAdmin(ctx, NewSuperAdmin{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  "Integration Admin",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("create Super Admin: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_auth_challenges WHERE super_admin_id = $1::uuid`, created.ID)
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_sessions WHERE super_admin_id = $1::uuid`, created.ID)
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_accounts WHERE id = $1::uuid`, created.ID)
	})

	loaded, err := store.FindSuperAdminByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find Super Admin: %v", err)
	}
	if loaded.PasswordHash == "integration-only-password" {
		t.Fatal("PostgreSQL contains a plaintext password")
	}
	passwordMatches, err := hasher.Verify(loaded.PasswordHash, "integration-only-password")
	if err != nil || !passwordMatches {
		t.Fatalf("verify persisted password: match=%v error=%v", passwordMatches, err)
	}

	if _, err := store.CreateSuperAdmin(ctx, NewSuperAdmin{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  "Duplicate",
		IsActive:     true,
	}); !errors.Is(err, ErrSuperAdminExists) {
		t.Fatalf("duplicate create error = %v, want %v", err, ErrSuperAdminExists)
	}

	// The injected text is a query parameter, not executable SQL. A failed
	// lookup also proves it cannot turn the predicate into a match-all clause.
	if _, err := store.FindSuperAdminByEmail(ctx, `' OR '1'='1`); !errors.Is(err, ErrSuperAdminNotFound) {
		t.Fatalf("SQL injection lookup error = %v, want not found", err)
	}

	rawToken := "integration-raw-session-token-never-store"
	tokenHash := sha256.Sum256([]byte(rawToken))
	now := time.Now().UTC().Truncate(time.Microsecond)
	if err := store.CreateSession(ctx, NewSession{
		SuperAdminID: created.ID,
		TokenHash:    tokenHash[:],
		CSRFToken:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CreatedAt:    now,
		ExpiresAt:    now.Add(8 * time.Hour),
		LastSeenAt:   now,
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	var persistedHash []byte
	if err := db.QueryRow(
		ctx,
		`SELECT token_hash FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		created.ID,
	).Scan(&persistedHash); err != nil {
		t.Fatalf("read persisted token hash: %v", err)
	}
	if !bytes.Equal(persistedHash, tokenHash[:]) || bytes.Equal(persistedHash, []byte(rawToken)) {
		t.Fatalf("persisted token material = %x", persistedHash)
	}

	session, err := store.FindSessionByTokenHash(ctx, tokenHash[:])
	if err != nil {
		t.Fatalf("find session: %v", err)
	}
	if session.SuperAdmin.Email != email || !session.SuperAdmin.IsActive {
		t.Fatalf("session identity = %#v", session.SuperAdmin)
	}
	if err := store.RevokeSession(ctx, session.ID, now.Add(time.Minute)); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	revoked, err := store.FindSessionByTokenHash(ctx, tokenHash[:])
	if err != nil || revoked.RevokedAt == nil {
		t.Fatalf("revoked session = %#v, error = %v", revoked, err)
	}
}

func TestPostgresStoreOTPChallengeLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping OTP repository integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	t.Cleanup(db.Close)

	store, err := NewPostgresStore(db)
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	passwordHash, err := NewPasswordHasher().Hash("integration-only-password")
	if err != nil {
		t.Fatalf("hash integration password: %v", err)
	}
	admin, err := store.CreateSuperAdmin(ctx, NewSuperAdmin{
		Email:        fmt.Sprintf("otp-integration-%d@example.com", time.Now().UnixNano()),
		PasswordHash: passwordHash,
		DisplayName:  "OTP Integration Admin",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("create Super Admin: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_auth_challenges WHERE super_admin_id = $1::uuid`, admin.ID)
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_sessions WHERE super_admin_id = $1::uuid`, admin.ID)
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_accounts WHERE id = $1::uuid`, admin.ID)
	})

	hasher, err := newOTPHasher(bytes.Repeat([]byte{51}, otpHashSecretMinBytes))
	if err != nil {
		t.Fatalf("newOTPHasher() error = %v", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("failed attempts lock on fifth and sixth cannot succeed", func(t *testing.T) {
		challengeID := tokenFromByte(41)
		correctDigest, _ := hasher.hash(challengeID, "001284")
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, challengeID, correctDigest[:], now)

		var persistedHash []byte
		if err := db.QueryRow(
			ctx,
			`SELECT otp_hash FROM super_admin_auth_challenges WHERE id = $1`,
			challengeID,
		).Scan(&persistedHash); err != nil {
			t.Fatalf("read OTP hash: %v", err)
		}
		if !bytes.Equal(persistedHash, correctDigest[:]) || bytes.Contains(persistedHash, []byte("001284")) {
			t.Fatal("PostgreSQL did not retain only the OTP HMAC")
		}

		wrongDigest, _ := hasher.hash(challengeID, "999999")
		for attempt := 1; attempt <= maxOTPFailedAttempts; attempt++ {
			_, verifyErr := store.VerifyOTPChallengeAndCreateSession(
				ctx,
				challengeID,
				wrongDigest[:],
				now.Add(time.Duration(attempt)*time.Second),
				integrationSession(admin.ID, now.Add(time.Duration(attempt)*time.Second), byte(attempt)),
			)
			if attempt < maxOTPFailedAttempts && !errors.Is(verifyErr, ErrOTPInvalid) {
				t.Fatalf("attempt %d error = %v, want %v", attempt, verifyErr, ErrOTPInvalid)
			}
			if attempt == maxOTPFailedAttempts && !errors.Is(verifyErr, ErrOTPAttemptsExceeded) {
				t.Fatalf("fifth attempt error = %v, want %v", verifyErr, ErrOTPAttemptsExceeded)
			}

			var failedAttempts int
			if err := db.QueryRow(
				ctx,
				`SELECT failed_attempts FROM super_admin_auth_challenges WHERE id = $1`,
				challengeID,
			).Scan(&failedAttempts); err != nil {
				t.Fatalf("read failed attempts: %v", err)
			}
			if failedAttempts != attempt {
				t.Fatalf("failed attempts after try %d = %d", attempt, failedAttempts)
			}
		}

		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			correctDigest[:],
			now.Add(10*time.Second),
			integrationSession(admin.ID, now.Add(10*time.Second), 10),
		); !errors.Is(err, ErrOTPAttemptsExceeded) {
			t.Fatalf("correct sixth attempt error = %v, want %v", err, ErrOTPAttemptsExceeded)
		}
	})

	t.Run("new password challenge invalidates previous flow", func(t *testing.T) {
		firstID := tokenFromByte(42)
		firstDigest, _ := hasher.hash(firstID, "123456")
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, firstID, firstDigest[:], now.Add(time.Minute))

		secondID := tokenFromByte(43)
		secondDigest, _ := hasher.hash(secondID, "654321")
		if err := store.CreateOTPChallenge(ctx, integrationChallenge(
			admin.ID,
			secondID,
			secondDigest[:],
			now.Add(2*time.Minute),
		)); err != nil {
			t.Fatalf("create replacement challenge: %v", err)
		}

		first, err := store.FindOTPChallenge(ctx, firstID)
		if err != nil {
			t.Fatalf("find first challenge: %v", err)
		}
		if first.InvalidatedAt == nil {
			t.Fatal("new password login left prior challenge usable")
		}
		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			firstID,
			firstDigest[:],
			now.Add(2*time.Minute),
			integrationSession(admin.ID, now.Add(2*time.Minute), 11),
		); !errors.Is(err, ErrOTPInvalidated) {
			t.Fatalf("prior challenge verification error = %v, want %v", err, ErrOTPInvalidated)
		}

		if err := store.ActivateOTPChallenge(ctx, secondID, 1, now.Add(2*time.Minute)); err != nil {
			t.Fatalf("activate replacement challenge: %v", err)
		}
	})

	t.Run("resend invalidates old code and preserves failure count", func(t *testing.T) {
		challengeID := tokenFromByte(44)
		oldDigest, _ := hasher.hash(challengeID, "222222")
		createdAt := now.Add(3 * time.Minute)
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, challengeID, oldDigest[:], createdAt)

		wrongDigest, _ := hasher.hash(challengeID, "333333")
		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			wrongDigest[:],
			createdAt.Add(time.Second),
			integrationSession(admin.ID, createdAt.Add(time.Second), 12),
		); !errors.Is(err, ErrOTPInvalid) {
			t.Fatalf("pre-resend wrong code error = %v", err)
		}

		resendAt := createdAt.Add(otpResendCooldown)
		if _, err := store.BeginOTPChallengeResend(
			ctx,
			challengeID,
			oldDigest[:],
			resendAt,
			resendAt.Add(otpValidity),
			resendAt.Add(otpResendCooldown),
		); !errors.Is(err, ErrOTPRotationCollision) {
			t.Fatalf("same-code resend error = %v, want collision", err)
		}

		newDigest, _ := hasher.hash(challengeID, "444444")
		delivery, err := store.BeginOTPChallengeResend(
			ctx,
			challengeID,
			newDigest[:],
			resendAt,
			resendAt.Add(otpValidity),
			resendAt.Add(otpResendCooldown),
		)
		if err != nil {
			t.Fatalf("begin resend: %v", err)
		}
		if delivery.DeliveryVersion != 2 {
			t.Fatalf("delivery version = %d, want 2", delivery.DeliveryVersion)
		}
		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			oldDigest[:],
			resendAt.Add(time.Second),
			integrationSession(admin.ID, resendAt.Add(time.Second), 13),
		); !errors.Is(err, ErrOTPDeliveryPending) {
			t.Fatalf("old OTP during delivery error = %v, want fail-closed pending", err)
		}
		if err := store.ActivateOTPChallenge(ctx, challengeID, delivery.DeliveryVersion, resendAt.Add(time.Second)); err != nil {
			t.Fatalf("activate resent OTP: %v", err)
		}
		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			oldDigest[:],
			resendAt.Add(2*time.Second),
			integrationSession(admin.ID, resendAt.Add(2*time.Second), 14),
		); !errors.Is(err, ErrOTPInvalid) {
			t.Fatalf("old OTP after activation error = %v, want %v", err, ErrOTPInvalid)
		}

		challenge, err := store.FindOTPChallenge(ctx, challengeID)
		if err != nil {
			t.Fatalf("find resent challenge: %v", err)
		}
		if challenge.FailedAttempts != 2 {
			t.Fatalf("failed attempts after resend = %d, want preserved+incremented 2", challenge.FailedAttempts)
		}
		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			newDigest[:],
			resendAt.Add(3*time.Second),
			integrationSession(admin.ID, resendAt.Add(3*time.Second), 15),
		); err != nil {
			t.Fatalf("new OTP verification: %v", err)
		}
	})

	t.Run("success consumes challenge and concurrent replay creates one session", func(t *testing.T) {
		challengeID := tokenFromByte(45)
		correctDigest, _ := hasher.hash(challengeID, "555555")
		createdAt := now.Add(5 * time.Minute)
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, challengeID, correctDigest[:], createdAt)

		var sessionsBefore int
		if err := db.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
			admin.ID,
		).Scan(&sessionsBefore); err != nil {
			t.Fatalf("count sessions before concurrent verify: %v", err)
		}

		type verifyResult struct{ err error }
		start := make(chan struct{})
		results := make(chan verifyResult, 2)
		for index := 0; index < 2; index++ {
			index := index
			go func() {
				<-start
				_, verifyErr := store.VerifyOTPChallengeAndCreateSession(
					ctx,
					challengeID,
					correctDigest[:],
					createdAt.Add(time.Second),
					integrationSession(admin.ID, createdAt.Add(time.Second), byte(20+index)),
				)
				results <- verifyResult{err: verifyErr}
			}()
		}
		close(start)

		successes := 0
		consumedFailures := 0
		for index := 0; index < 2; index++ {
			result := <-results
			switch {
			case result.err == nil:
				successes++
			case errors.Is(result.err, ErrOTPConsumed):
				consumedFailures++
			default:
				t.Fatalf("concurrent verification error = %v", result.err)
			}
		}
		if successes != 1 || consumedFailures != 1 {
			t.Fatalf("concurrent results = %d success, %d consumed", successes, consumedFailures)
		}

		var sessionsAfter int
		if err := db.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
			admin.ID,
		).Scan(&sessionsAfter); err != nil {
			t.Fatalf("count sessions after concurrent verify: %v", err)
		}
		if sessionsAfter != sessionsBefore+1 {
			t.Fatalf("sessions after concurrent verify = %d, want %d", sessionsAfter, sessionsBefore+1)
		}
	})

	t.Run("expired challenge rejects the correct code", func(t *testing.T) {
		challengeID := tokenFromByte(46)
		correctDigest, _ := hasher.hash(challengeID, "666666")
		createdAt := now.Add(6 * time.Minute)
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, challengeID, correctDigest[:], createdAt)

		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			correctDigest[:],
			createdAt.Add(otpValidity),
			integrationSession(admin.ID, createdAt.Add(otpValidity), 30),
		); !errors.Is(err, ErrOTPExpired) {
			t.Fatalf("expired challenge verification error = %v, want %v", err, ErrOTPExpired)
		}

		challenge, err := store.FindOTPChallenge(ctx, challengeID)
		if err != nil {
			t.Fatalf("find expired challenge: %v", err)
		}
		if challenge.ConsumedAt != nil || challenge.FailedAttempts != 0 {
			t.Fatalf("expired verification mutated challenge = %#v", challenge)
		}
	})

	t.Run("resend before cooldown preserves the active code", func(t *testing.T) {
		challengeID := tokenFromByte(47)
		oldDigest, _ := hasher.hash(challengeID, "777777")
		newDigest, _ := hasher.hash(challengeID, "888888")
		createdAt := now.Add(7 * time.Minute)
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, challengeID, oldDigest[:], createdAt)

		resendAt := createdAt.Add(30 * time.Second)
		_, err := store.BeginOTPChallengeResend(
			ctx,
			challengeID,
			newDigest[:],
			resendAt,
			resendAt.Add(otpValidity),
			resendAt.Add(otpResendCooldown),
		)
		var tooEarly *OTPResendTooEarlyError
		if !errors.As(err, &tooEarly) || tooEarly.RetryAfter != 30*time.Second {
			t.Fatalf("early resend error = %#v, want 30-second cooldown", err)
		}

		challenge, err := store.FindOTPChallenge(ctx, challengeID)
		if err != nil {
			t.Fatalf("find challenge after early resend: %v", err)
		}
		if challenge.DeliveryVersion != 1 || challenge.ActiveVersion == nil || *challenge.ActiveVersion != 1 {
			t.Fatalf("early resend changed delivery state = %#v", challenge)
		}
		if !bytes.Equal(challenge.OTPHash, oldDigest[:]) {
			t.Fatal("early resend replaced the active OTP hash")
		}
	})

	t.Run("successful verification preserves prior failed attempts", func(t *testing.T) {
		challengeID := tokenFromByte(48)
		correctDigest, _ := hasher.hash(challengeID, "123123")
		wrongDigest, _ := hasher.hash(challengeID, "321321")
		createdAt := now.Add(8 * time.Minute)
		createAndActivateIntegrationChallenge(t, ctx, store, admin.ID, challengeID, correctDigest[:], createdAt)

		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			wrongDigest[:],
			createdAt.Add(time.Second),
			integrationSession(admin.ID, createdAt.Add(time.Second), 31),
		); !errors.Is(err, ErrOTPInvalid) {
			t.Fatalf("wrong OTP error = %v, want %v", err, ErrOTPInvalid)
		}
		if _, err := store.VerifyOTPChallengeAndCreateSession(
			ctx,
			challengeID,
			correctDigest[:],
			createdAt.Add(2*time.Second),
			integrationSession(admin.ID, createdAt.Add(2*time.Second), 31),
		); err != nil {
			t.Fatalf("correct OTP verification: %v", err)
		}

		challenge, err := store.FindOTPChallenge(ctx, challengeID)
		if err != nil {
			t.Fatalf("find consumed challenge: %v", err)
		}
		if challenge.ConsumedAt == nil || challenge.FailedAttempts != 1 {
			t.Fatalf("successful verification changed failed attempts = %#v", challenge)
		}
	})

	t.Run("concurrent same-admin creation leaves one current challenge", func(t *testing.T) {
		createdAt := now.Add(9 * time.Minute)
		challengeIDs := []string{tokenFromByte(49), tokenFromByte(50)}
		challenges := make([]NewOTPChallenge, 0, len(challengeIDs))
		for index, challengeID := range challengeIDs {
			digest, hashErr := hasher.hash(challengeID, fmt.Sprintf("90000%d", index))
			if hashErr != nil {
				t.Fatalf("hash concurrent OTP: %v", hashErr)
			}
			challenges = append(challenges, integrationChallenge(admin.ID, challengeID, digest[:], createdAt))
		}

		start := make(chan struct{})
		results := make(chan error, len(challenges))
		for _, challenge := range challenges {
			challenge := challenge
			go func() {
				<-start
				results <- store.CreateOTPChallenge(ctx, challenge)
			}()
		}
		close(start)

		successes := 0
		for range challenges {
			err := <-results
			switch {
			case err == nil:
				successes++
			case isUniqueViolation(err):
				// The partial unique index is the final fail-closed guard if a
				// future transaction implementation no longer serializes on the account.
			default:
				t.Fatalf("concurrent challenge creation error = %v", err)
			}
		}
		if successes == 0 {
			t.Fatal("concurrent challenge creation produced no current flow")
		}

		var currentCount int
		var currentID string
		if err := db.QueryRow(
			ctx,
			`SELECT COUNT(*), MIN(id)
			 FROM super_admin_auth_challenges
			 WHERE super_admin_id = $1::uuid
			   AND consumed_at IS NULL
			   AND invalidated_at IS NULL`,
			admin.ID,
		).Scan(&currentCount, &currentID); err != nil {
			t.Fatalf("read current challenge count: %v", err)
		}
		if currentCount != 1 {
			t.Fatalf("current same-admin challenges = %d, want exactly one", currentCount)
		}
		if err := store.ActivateOTPChallenge(ctx, currentID, 1, createdAt); err != nil {
			t.Fatalf("activate surviving concurrent challenge: %v", err)
		}
		current, err := store.FindOTPChallenge(ctx, currentID)
		if err != nil {
			t.Fatalf("find surviving concurrent challenge: %v", err)
		}
		if err := usableOTPChallenge(current, createdAt.Add(time.Second)); err != nil {
			t.Fatalf("surviving concurrent challenge is unusable: %v", err)
		}
	})
}

func integrationChallenge(
	adminID string,
	challengeID string,
	otpHash []byte,
	createdAt time.Time,
) NewOTPChallenge {
	return NewOTPChallenge{
		ID:                challengeID,
		SuperAdminID:      adminID,
		OTPHash:           otpHash,
		CreatedAt:         createdAt,
		ExpiresAt:         createdAt.Add(otpValidity),
		ResendAvailableAt: createdAt.Add(otpResendCooldown),
		DeliveryVersion:   1,
		DeliveryStartedAt: createdAt,
	}
}

func createAndActivateIntegrationChallenge(
	t *testing.T,
	ctx context.Context,
	store *PostgresStore,
	adminID string,
	challengeID string,
	otpHash []byte,
	createdAt time.Time,
) {
	t.Helper()

	if err := store.CreateOTPChallenge(ctx, integrationChallenge(adminID, challengeID, otpHash, createdAt)); err != nil {
		t.Fatalf("create OTP challenge: %v", err)
	}
	if err := store.ActivateOTPChallenge(ctx, challengeID, 1, createdAt); err != nil {
		t.Fatalf("activate OTP challenge: %v", err)
	}
}

func integrationSession(adminID string, createdAt time.Time, value byte) NewSession {
	return NewSession{
		SuperAdminID: adminID,
		TokenHash:    bytes.Repeat([]byte{value}, sha256.Size),
		CSRFToken:    tokenFromByte(value + 80),
		CreatedAt:    createdAt,
		ExpiresAt:    createdAt.Add(time.Hour),
		LastSeenAt:   createdAt,
	}
}
