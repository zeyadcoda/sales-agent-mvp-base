package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestRealOTPLoginDoesNotSendForInvalidAccountCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		store    *fakeStore
		verifier *fakePasswordVerifier
	}{
		{
			name:     "unknown account",
			store:    &fakeStore{adminErr: ErrSuperAdminNotFound},
			verifier: &fakePasswordVerifier{results: map[string]bool{}},
		},
		{
			name:  "wrong password",
			store: &fakeStore{admin: activeAdmin()},
			verifier: &fakePasswordVerifier{
				results: map[string]bool{"stored-hash": false},
			},
		},
		{
			name: "inactive account",
			store: &fakeStore{admin: func() SuperAdmin {
				admin := activeAdmin()
				admin.IsActive = false
				return admin
			}()},
			verifier: &fakePasswordVerifier{
				results: map[string]bool{"stored-hash": true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			sender := &fakeOTPEmailSender{}
			service := newConfiguredOTPService(
				t,
				test.store,
				&fakeLimiter{},
				&fakeOTPLimiter{},
				test.verifier,
				sender,
				time.Now().UTC(),
				otpRandomMaterial(1, 0),
			)

			_, err := service.Login(
				context.Background(),
				"admin@example.com",
				"incorrect password",
				"192.0.2.10",
			)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
			}
			if len(sender.messages) != 0 {
				t.Fatalf("invalid credentials sent %d OTP emails", len(sender.messages))
			}
			if test.store.createdChallenge.ID != "" || test.store.createdSession.SuperAdminID != "" {
				t.Fatal("invalid credentials created authentication state")
			}
		})
	}
}

func TestRealOTPLoginPersistsOnlyHashAndReturnsMinimalContext(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	store := &fakeStore{admin: activeAdmin()}
	sender := &fakeOTPEmailSender{}
	service := newConfiguredOTPService(
		t,
		store,
		&fakeLimiter{},
		&fakeOTPLimiter{},
		&fakePasswordVerifier{results: map[string]bool{"stored-hash": true}},
		sender,
		now,
		otpRandomMaterial(8, 0),
	)

	result, err := service.Login(
		context.Background(),
		"admin@example.com",
		"correct password",
		"192.0.2.10",
	)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Challenge == nil || result.Authenticated != nil {
		t.Fatalf("Login() result = %#v, want challenge only", result)
	}
	if len(sender.messages) != 1 || sender.messages[0].OTP != "000000" {
		t.Fatalf("email messages = %#v", sender.messages)
	}
	if len(store.createdChallenge.OTPHash) != sha256.Size {
		t.Fatalf("persisted OTP hash length = %d", len(store.createdChallenge.OTPHash))
	}
	if bytes.Equal(store.createdChallenge.OTPHash, []byte(sender.messages[0].OTP)) ||
		bytes.Contains(store.createdChallenge.OTPHash, []byte(sender.messages[0].OTP)) {
		t.Fatal("plaintext OTP reached persistence")
	}
	if result.Challenge.ID != store.createdChallenge.ID || !validChallengeID(result.Challenge.ID) {
		t.Fatalf("browser challenge = %#v", result.Challenge)
	}
	if result.Challenge.DestinationHint != "a***@example.com" {
		t.Fatalf("destination hint = %q", result.Challenge.DestinationHint)
	}
	if !result.Challenge.ExpiresAt.Equal(now.Add(otpValidity)) ||
		!result.Challenge.ResendAvailableAt.Equal(now.Add(otpResendCooldown)) {
		t.Fatalf("challenge timing = %#v", result.Challenge)
	}
	if store.createdSession.SuperAdminID != "" {
		t.Fatal("password plus delivery created a session before OTP verification")
	}
}

func TestRealOTPLoginDeliveryOrActivationFailureCannotAuthenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		senderErr   error
		activateErr error
	}{
		{name: "delivery failure", senderErr: errors.New("raw SMTP provider detail")},
		{name: "activation failure", activateErr: errors.New("raw database detail")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &fakeStore{admin: activeAdmin(), activateChallengeErr: test.activateErr}
			sender := &fakeOTPEmailSender{err: test.senderErr}
			service := newConfiguredOTPService(
				t,
				store,
				&fakeLimiter{},
				&fakeOTPLimiter{},
				&fakePasswordVerifier{results: map[string]bool{"stored-hash": true}},
				sender,
				time.Now().UTC(),
				otpRandomMaterial(9, 0),
			)

			result, err := service.Login(
				context.Background(),
				"admin@example.com",
				"correct password",
				"192.0.2.10",
			)
			if !errors.Is(err, ErrAuthenticationUnavailable) {
				t.Fatalf("Login() error = %v, want safe unavailable error", err)
			}
			if result.Authenticated != nil || result.Challenge != nil || store.createdSession.SuperAdminID != "" {
				t.Fatal("failed OTP delivery/activation returned authentication state")
			}
			if store.invalidatedChallengeID != store.createdChallenge.ID || store.invalidatedDeliveryVersion != 1 {
				t.Fatalf("invalidated delivery = (%q, %d)", store.invalidatedChallengeID, store.invalidatedDeliveryVersion)
			}
		})
	}
}

func TestVerifyOTPCreatesIndependentHashedSessionAfterStoreConsumesChallenge(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 5, 0, 0, time.UTC)
	challengeID := tokenFromByte(12)
	store := &fakeStore{verifyAdmin: activeAdmin()}
	otpLimiter := &fakeOTPLimiter{}
	service := newConfiguredOTPService(
		t,
		store,
		&fakeLimiter{},
		otpLimiter,
		&fakePasswordVerifier{},
		&fakeOTPEmailSender{},
		now,
		append(bytes.Repeat([]byte{21}, sessionTokenBytes), bytes.Repeat([]byte{22}, sessionTokenBytes)...),
	)

	authenticated, err := service.VerifyOTP(
		context.Background(),
		challengeID,
		"093221",
		"192.0.2.10",
	)
	if err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}
	if otpLimiter.verifyCalls != 1 || store.verifiedChallengeID != challengeID {
		t.Fatalf("verify calls = limiter:%d store:%q", otpLimiter.verifyCalls, store.verifiedChallengeID)
	}
	hasher, _ := newOTPHasher(bytes.Repeat([]byte{7}, otpHashSecretMinBytes))
	wantDigest, _ := hasher.hash(challengeID, "093221")
	if !bytes.Equal(store.verifiedCandidateHash, wantDigest[:]) ||
		bytes.Equal(store.verifiedCandidateHash, []byte("093221")) {
		t.Fatal("store did not receive only the challenge-bound OTP digest")
	}
	if !validOpaqueToken(authenticated.RawSessionToken) || authenticated.Session.CSRFToken == authenticated.RawSessionToken {
		t.Fatal("OTP success did not produce independent 256-bit session and CSRF tokens")
	}
	wantTokenHash := sha256.Sum256([]byte(authenticated.RawSessionToken))
	if !bytes.Equal(store.verifiedSession.TokenHash, wantTokenHash[:]) {
		t.Fatalf("stored session hash = %x, want %x", store.verifiedSession.TokenHash, wantTokenHash)
	}
	if store.verifiedSession.SuperAdminID != "" {
		t.Fatal("service supplied an account ID instead of letting the locked challenge resolve ownership")
	}
	if authenticated.Session.SuperAdmin.PasswordHash != "" ||
		!authenticated.Session.ExpiresAt.Equal(now.Add(8*time.Hour)) {
		t.Fatalf("authenticated session = %#v", authenticated.Session)
	}
}

func TestVerifyOTPRejectsMalformedInputAndLimiterFailureBeforePersistence(t *testing.T) {
	t.Parallel()

	challengeID := tokenFromByte(13)
	tests := []struct {
		name       string
		challenge  string
		otp        string
		limiterErr error
		want       error
	}{
		{name: "malformed challenge", challenge: "not-a-challenge", otp: "123456", want: ErrOTPInvalid},
		{name: "malformed code", challenge: challengeID, otp: "12345x", want: ErrOTPInvalid},
		{name: "Redis failure", challenge: challengeID, otp: "123456", limiterErr: ErrRateLimitUnavailable, want: ErrAuthenticationUnavailable},
		{name: "verification throttled", challenge: challengeID, otp: "123456", limiterErr: ErrOTPVerifyRateLimited, want: ErrOTPVerifyRateLimited},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeStore{verifyAdmin: activeAdmin()}
			limiter := &fakeOTPLimiter{verifyErr: test.limiterErr}
			service := newConfiguredOTPService(
				t,
				store,
				&fakeLimiter{},
				limiter,
				&fakePasswordVerifier{},
				&fakeOTPEmailSender{},
				time.Now().UTC(),
				bytes.Repeat([]byte{1}, sessionTokenBytes*2),
			)

			_, err := service.VerifyOTP(context.Background(), test.challenge, test.otp, "192.0.2.10")
			if !errors.Is(err, test.want) {
				t.Fatalf("VerifyOTP() error = %v, want %v", err, test.want)
			}
			if store.verifiedChallengeID != "" {
				t.Fatal("rejected verification reached PostgreSQL verification")
			}
		})
	}
}

func TestVerifyOTPMapsChallengeLifecycleWithoutInternalLeakage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		storeErr error
		want     error
	}{
		{name: "unknown", storeErr: ErrOTPChallengeNotFound, want: ErrOTPInvalid},
		{name: "wrong code", storeErr: ErrOTPInvalid, want: ErrOTPInvalid},
		{name: "expired", storeErr: ErrOTPExpired, want: ErrOTPExpired},
		{name: "attempts exhausted", storeErr: ErrOTPAttemptsExceeded, want: ErrOTPAttemptsExceeded},
		{name: "invalidated", storeErr: ErrOTPInvalidated, want: ErrOTPInvalidated},
		{name: "consumed", storeErr: ErrOTPConsumed, want: ErrOTPConsumed},
		{name: "pending delivery", storeErr: ErrOTPDeliveryPending, want: ErrAuthenticationUnavailable},
		{name: "database failure", storeErr: errors.New("raw SQL detail"), want: ErrAuthenticationUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeStore{verifyErr: test.storeErr}
			service := newConfiguredOTPService(
				t,
				store,
				&fakeLimiter{},
				&fakeOTPLimiter{},
				&fakePasswordVerifier{},
				&fakeOTPEmailSender{},
				time.Now().UTC(),
				bytes.Repeat([]byte{1}, sessionTokenBytes*2),
			)

			_, err := service.VerifyOTP(context.Background(), tokenFromByte(14), "123456", "192.0.2.10")
			if !errors.Is(err, test.want) {
				t.Fatalf("VerifyOTP() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestResendOTPRotatesBeforeDeliveryAndActivatesOnlyNewestVersion(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 2, 0, 0, time.UTC)
	challengeID := tokenFromByte(15)
	store := &fakeStore{resendDelivery: OTPDelivery{
		ChallengeID:       challengeID,
		SuperAdmin:        activeAdmin(),
		DeliveryVersion:   2,
		ExpiresAt:         now.Add(otpValidity),
		ResendAvailableAt: now.Add(otpResendCooldown),
	}}
	sender := &fakeOTPEmailSender{}
	service := newConfiguredOTPService(
		t,
		store,
		&fakeLimiter{},
		&fakeOTPLimiter{},
		&fakePasswordVerifier{},
		sender,
		now,
		[]byte{0, 0, 0},
	)

	pending, err := service.ResendOTP(context.Background(), challengeID, "192.0.2.10")
	if err != nil {
		t.Fatalf("ResendOTP() error = %v", err)
	}
	if len(store.resendHashes) != 1 || len(store.resendHashes[0]) != sha256.Size {
		t.Fatalf("resend hashes = %#v", store.resendHashes)
	}
	if len(sender.messages) != 1 || sender.messages[0].OTP != "000000" {
		t.Fatalf("resend messages = %#v", sender.messages)
	}
	if store.activatedChallengeID != challengeID || store.activatedDeliveryVersion != 2 {
		t.Fatalf("activation = (%q, %d)", store.activatedChallengeID, store.activatedDeliveryVersion)
	}
	if pending.ID != challengeID || pending.DestinationHint != "a***@example.com" {
		t.Fatalf("pending challenge = %#v", pending)
	}
}

func TestResendOTPRejectsGeneratedCodeCollisionAndGeneratesAnother(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 2, 0, 0, time.UTC)
	challengeID := tokenFromByte(16)
	store := &fakeStore{
		resendDelivery: OTPDelivery{
			ChallengeID:       challengeID,
			SuperAdmin:        activeAdmin(),
			DeliveryVersion:   2,
			ExpiresAt:         now.Add(otpValidity),
			ResendAvailableAt: now.Add(otpResendCooldown),
		},
		resendErrors: []error{ErrOTPRotationCollision, nil},
	}
	sender := &fakeOTPEmailSender{}
	service := newConfiguredOTPService(
		t,
		store,
		&fakeLimiter{},
		&fakeOTPLimiter{},
		&fakePasswordVerifier{},
		sender,
		now,
		[]byte{0, 0, 0, 0, 0, 1},
	)

	if _, err := service.ResendOTP(context.Background(), challengeID, "192.0.2.10"); err != nil {
		t.Fatalf("ResendOTP() error = %v", err)
	}
	if len(store.resendHashes) != 2 || bytes.Equal(store.resendHashes[0], store.resendHashes[1]) {
		t.Fatalf("resend did not replace colliding digest: %#v", store.resendHashes)
	}
	if len(sender.messages) != 1 || sender.messages[0].OTP != "000001" {
		t.Fatalf("delivered OTP = %#v, want newly generated 000001", sender.messages)
	}
}

func TestResendOTPDeliveryFailureInvalidatesPendingRotation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	challengeID := tokenFromByte(17)
	store := &fakeStore{resendDelivery: OTPDelivery{
		ChallengeID:       challengeID,
		SuperAdmin:        activeAdmin(),
		DeliveryVersion:   4,
		ExpiresAt:         now.Add(otpValidity),
		ResendAvailableAt: now.Add(otpResendCooldown),
	}}
	service := newConfiguredOTPService(
		t,
		store,
		&fakeLimiter{},
		&fakeOTPLimiter{},
		&fakePasswordVerifier{},
		&fakeOTPEmailSender{err: errors.New("raw provider error")},
		now,
		[]byte{0, 0, 0},
	)

	_, err := service.ResendOTP(context.Background(), challengeID, "192.0.2.10")
	if !errors.Is(err, ErrAuthenticationUnavailable) {
		t.Fatalf("ResendOTP() error = %v", err)
	}
	if store.invalidatedChallengeID != challengeID || store.invalidatedDeliveryVersion != 4 {
		t.Fatalf("invalidated rotation = (%q, %d)", store.invalidatedChallengeID, store.invalidatedDeliveryVersion)
	}
	if store.activatedChallengeID != "" {
		t.Fatal("failed email delivery activated a code")
	}
}

func TestResendOTPReturnsAuthoritativeCooldownWithoutEmail(t *testing.T) {
	t.Parallel()

	retry := 37 * time.Second
	store := &fakeStore{resendErr: &OTPResendTooEarlyError{RetryAfter: retry}}
	sender := &fakeOTPEmailSender{}
	service := newConfiguredOTPService(
		t,
		store,
		&fakeLimiter{},
		&fakeOTPLimiter{},
		&fakePasswordVerifier{},
		sender,
		time.Now().UTC(),
		[]byte{0, 0, 0},
	)

	_, err := service.ResendOTP(context.Background(), tokenFromByte(18), "192.0.2.10")
	var cooldown *OTPResendTooEarlyError
	if !errors.As(err, &cooldown) || cooldown.RetryAfter != retry {
		t.Fatalf("ResendOTP() error = %#v, want retry after %s", err, retry)
	}
	if len(sender.messages) != 0 {
		t.Fatal("cooldown rejection sent an email")
	}
}

func TestGetOTPChallengeStatusReturnsSafeLifecycleStates(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 5, 0, 0, time.UTC)
	activeVersion := 2
	base := OTPChallenge{
		ID:                tokenFromByte(19),
		SuperAdmin:        activeAdmin(),
		CreatedAt:         now.Add(-time.Minute),
		ExpiresAt:         now.Add(9 * time.Minute),
		ResendAvailableAt: now.Add(30 * time.Second),
		DeliveryVersion:   2,
		ActiveVersion:     &activeVersion,
	}
	terminalAt := now.Add(-time.Second)
	tests := []struct {
		name      string
		mutate    func(*OTPChallenge)
		wantState OTPChallengeState
	}{
		{name: "active", wantState: OTPChallengePending},
		{name: "expired", mutate: func(c *OTPChallenge) { c.ExpiresAt = now }, wantState: OTPChallengeExpired},
		{name: "invalidated", mutate: func(c *OTPChallenge) { c.InvalidatedAt = &terminalAt }, wantState: OTPChallengeInvalid},
		{name: "consumed", mutate: func(c *OTPChallenge) { c.ConsumedAt = &terminalAt }, wantState: OTPChallengeUsed},
		{name: "attempts exhausted", mutate: func(c *OTPChallenge) { c.FailedAttempts = 5 }, wantState: OTPChallengeLocked},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			challenge := base
			if test.mutate != nil {
				test.mutate(&challenge)
			}
			service := newConfiguredOTPService(
				t,
				&fakeStore{challenge: challenge},
				&fakeLimiter{},
				&fakeOTPLimiter{},
				&fakePasswordVerifier{},
				&fakeOTPEmailSender{},
				now,
				nil,
			)

			state, err := service.GetOTPChallengeStatus(context.Background(), challenge.ID)
			if err != nil {
				t.Fatalf("GetOTPChallengeStatus() error = %v", err)
			}
			if state.State != test.wantState || state.DestinationHint != "a***@example.com" {
				t.Fatalf("challenge state = %#v", state)
			}
		})
	}
}

func TestGetOTPChallengeStatusRejectsUnknownAndPendingDelivery(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	challengeID := tokenFromByte(20)
	tests := []struct {
		name  string
		store *fakeStore
		want  error
	}{
		{name: "unknown", store: &fakeStore{challengeErr: ErrOTPChallengeNotFound}, want: ErrOTPInvalid},
		{name: "pending delivery", store: &fakeStore{challenge: OTPChallenge{
			ID:                challengeID,
			SuperAdmin:        activeAdmin(),
			ExpiresAt:         now.Add(time.Minute),
			DeliveryVersion:   2,
			ActiveVersion:     nil,
			ResendAvailableAt: now,
		}}, want: ErrAuthenticationUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := newConfiguredOTPService(
				t,
				test.store,
				&fakeLimiter{},
				&fakeOTPLimiter{},
				&fakePasswordVerifier{},
				&fakeOTPEmailSender{},
				now,
				nil,
			)
			_, err := service.GetOTPChallengeStatus(context.Background(), challengeID)
			if !errors.Is(err, test.want) {
				t.Fatalf("GetOTPChallengeStatus() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestServiceOTPDependencyValidationAndBypassFailClosed(t *testing.T) {
	t.Parallel()

	baseOptions := ServiceOptions{
		SessionTTL:        time.Hour,
		DummyPasswordHash: "dummy",
		OTPHashSecret:     bytes.Repeat([]byte{1}, otpHashSecretMinBytes),
	}
	store := &fakeStore{}
	loginLimiter := &fakeLimiter{}
	passwords := &fakePasswordVerifier{}

	for _, test := range []struct {
		name       string
		otpLimiter OTPRateLimiter
		sender     OTPEmailSender
		options    ServiceOptions
	}{
		{name: "missing OTP limiter", sender: &fakeOTPEmailSender{}, options: baseOptions},
		{name: "missing email sender", otpLimiter: &fakeOTPLimiter{}, options: baseOptions},
		{name: "missing hash secret", otpLimiter: &fakeOTPLimiter{}, sender: &fakeOTPEmailSender{}, options: func() ServiceOptions {
			options := baseOptions
			options.OTPHashSecret = nil
			return options
		}()},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewService(store, loginLimiter, test.otpLimiter, passwords, test.sender, test.options); err == nil {
				t.Fatal("NewService() error = nil, want missing OTP dependency failure")
			}
		})
	}

	bypassOptions := baseOptions
	bypassOptions.OTPBypassEnabled = true
	bypassOptions.OTPHashSecret = nil
	service, err := NewService(store, loginLimiter, nil, passwords, nil, bypassOptions)
	if err != nil {
		t.Fatalf("NewService() local bypass error = %v", err)
	}
	if _, err := service.VerifyOTP(context.Background(), tokenFromByte(1), "123456", "192.0.2.1"); !errors.Is(err, ErrAuthenticationUnavailable) {
		t.Fatalf("VerifyOTP() under bypass error = %v", err)
	}
}

func newConfiguredOTPService(
	t *testing.T,
	store Store,
	loginLimiter LoginRateLimiter,
	otpLimiter OTPRateLimiter,
	passwords PasswordVerifier,
	sender OTPEmailSender,
	now time.Time,
	randomMaterial []byte,
) *Service {
	t.Helper()

	service, err := NewService(
		store,
		loginLimiter,
		otpLimiter,
		passwords,
		sender,
		ServiceOptions{
			OTPHashSecret:     bytes.Repeat([]byte{7}, otpHashSecretMinBytes),
			SessionTTL:        8 * time.Hour,
			DummyPasswordHash: "dummy-hash",
			Random:            bytes.NewReader(randomMaterial),
			Now:               func() time.Time { return now },
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}
