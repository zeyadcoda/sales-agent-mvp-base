package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/requestmeta"
)

type fakeRecoveryAdministrationStore struct {
	authorizeCalls       int
	authorizeEmail       string
	authorizeReason      string
	authorizeOperator    string
	authorizeCorrelation string
	authorizeCreatedAt   time.Time
	authorizeExpiresAt   time.Time
	authorizeResult      RecoveryAuthorization
	authorizeErr         error

	statusCalls  int
	statusEmail  string
	statusNow    time.Time
	statusResult RecoveryAuthorizationStatus
	statusErr    error

	revokeCalls       int
	revokeEmail       string
	revokeReason      string
	revokeOperator    string
	revokeCorrelation string
	revokedAt         time.Time
	revokeResult      RecoveryAuthorization
	revokeErr         error
}

func (store *fakeRecoveryAdministrationStore) AuthorizeSuperAdminRecovery(
	_ context.Context,
	normalizedEmail string,
	reason string,
	operatorIdentifier string,
	correlationID string,
	createdAt time.Time,
	expiresAt time.Time,
) (RecoveryAuthorization, error) {
	store.authorizeCalls++
	store.authorizeEmail = normalizedEmail
	store.authorizeReason = reason
	store.authorizeOperator = operatorIdentifier
	store.authorizeCorrelation = correlationID
	store.authorizeCreatedAt = createdAt
	store.authorizeExpiresAt = expiresAt
	return store.authorizeResult, store.authorizeErr
}

func (store *fakeRecoveryAdministrationStore) FindSuperAdminRecoveryStatus(
	_ context.Context,
	normalizedEmail string,
	now time.Time,
) (RecoveryAuthorizationStatus, error) {
	store.statusCalls++
	store.statusEmail = normalizedEmail
	store.statusNow = now
	return store.statusResult, store.statusErr
}

func (store *fakeRecoveryAdministrationStore) RevokeSuperAdminRecovery(
	_ context.Context,
	normalizedEmail string,
	reason string,
	operatorIdentifier string,
	correlationID string,
	revokedAt time.Time,
) (RecoveryAuthorization, error) {
	store.revokeCalls++
	store.revokeEmail = normalizedEmail
	store.revokeReason = reason
	store.revokeOperator = operatorIdentifier
	store.revokeCorrelation = correlationID
	store.revokedAt = revokedAt
	return store.revokeResult, store.revokeErr
}

func TestRecoveryAuthorizeNormalizesInputsAndUsesTenMinuteDefault(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 30, 0, 0, time.FixedZone("test", 4*60*60))
	want := RecoveryAuthorization{
		ID:                 "00000000-0000-0000-0000-000000000020",
		SuperAdminID:       "00000000-0000-0000-0000-000000000001",
		SuperAdminEmail:    "admin@example.com",
		Reason:             "notification provider outage",
		OperatorIdentifier: "on-call/sre-1",
		CorrelationID:      "recovery-correlation",
		CreatedAt:          now.UTC(),
		ExpiresAt:          now.UTC().Add(DefaultRecoveryAuthorizationValidity),
	}
	store := &fakeRecoveryAdministrationStore{authorizeResult: want}
	service := newRecoveryTestService(t, store, RecoveryServiceOptions{
		Now: func() time.Time { return now },
	})
	ctx := requestmeta.WithCorrelationID(context.Background(), "recovery-correlation")

	got, err := service.Authorize(
		ctx,
		"  ADMIN@Example.COM  ",
		"  notification provider outage  ",
		"  on-call/sre-1  ",
	)
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if got != want {
		t.Fatalf("Authorize() = %#v, want %#v", got, want)
	}
	if store.authorizeCalls != 1 || store.authorizeEmail != "admin@example.com" {
		t.Fatalf("authorize calls/email = %d/%q", store.authorizeCalls, store.authorizeEmail)
	}
	if store.authorizeReason != "notification provider outage" || store.authorizeOperator != "on-call/sre-1" {
		t.Fatalf("authorize attribution = reason %q, operator %q", store.authorizeReason, store.authorizeOperator)
	}
	if store.authorizeCorrelation != "recovery-correlation" {
		t.Fatalf("authorize correlation ID = %q", store.authorizeCorrelation)
	}
	if !store.authorizeCreatedAt.Equal(now.UTC()) {
		t.Fatalf("created at = %s, want %s", store.authorizeCreatedAt, now.UTC())
	}
	if !store.authorizeExpiresAt.Equal(now.UTC().Add(10 * time.Minute)) {
		t.Fatalf("expires at = %s, want 10 minutes after creation", store.authorizeExpiresAt)
	}
}

func TestRecoveryAuthorizeRejectsInvalidInputsBeforePersistence(t *testing.T) {
	t.Parallel()

	invalidUTF8 := string([]byte{0xff})
	tests := []struct {
		name     string
		email    string
		reason   string
		operator string
		want     error
	}{
		{name: "malformed target", email: "not-an-email", reason: "provider outage", operator: "operator", want: ErrRecoveryTargetNotEligible},
		{name: "missing reason", email: "admin@example.com", reason: "  ", operator: "operator", want: ErrInvalidRecoveryReason},
		{name: "multiline reason", email: "admin@example.com", reason: "provider\noutage", operator: "operator", want: ErrInvalidRecoveryReason},
		{name: "reason line separator", email: "admin@example.com", reason: "provider\u2028outage", operator: "operator", want: ErrInvalidRecoveryReason},
		{name: "reason too long", email: "admin@example.com", reason: strings.Repeat("r", maximumRecoveryReasonRunes+1), operator: "operator", want: ErrInvalidRecoveryReason},
		{name: "invalid reason encoding", email: "admin@example.com", reason: invalidUTF8, operator: "operator", want: ErrInvalidRecoveryReason},
		{name: "missing operator", email: "admin@example.com", reason: "provider outage", operator: "", want: ErrInvalidRecoveryOperator},
		{name: "multiline operator", email: "admin@example.com", reason: "provider outage", operator: "operator\rname", want: ErrInvalidRecoveryOperator},
		{name: "operator too long", email: "admin@example.com", reason: "provider outage", operator: strings.Repeat("o", maximumRecoveryOperatorRunes+1), want: ErrInvalidRecoveryOperator},
		{name: "invalid operator encoding", email: "admin@example.com", reason: "provider outage", operator: invalidUTF8, want: ErrInvalidRecoveryOperator},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &fakeRecoveryAdministrationStore{}
			service := newRecoveryTestService(t, store, RecoveryServiceOptions{})
			_, err := service.Authorize(context.Background(), test.email, test.reason, test.operator)
			if !errors.Is(err, test.want) {
				t.Fatalf("Authorize() error = %v, want %v", err, test.want)
			}
			if store.authorizeCalls != 0 {
				t.Fatal("invalid recovery input reached persistence")
			}
		})
	}
}

func TestRecoveryAuthorizeMapsTargetDuplicateAndInfrastructureErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		storeErr error
		want     error
	}{
		{name: "unknown or inactive target", storeErr: ErrRecoveryTargetNotEligible, want: ErrRecoveryTargetNotEligible},
		{name: "active authorization exists", storeErr: ErrRecoveryAlreadyActive, want: ErrRecoveryAlreadyActive},
		{name: "database or audit failure", storeErr: errors.New("raw PostgreSQL detail"), want: ErrRecoveryUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &fakeRecoveryAdministrationStore{authorizeErr: test.storeErr}
			service := newRecoveryTestService(t, store, RecoveryServiceOptions{})
			_, err := service.Authorize(
				context.Background(),
				"admin@example.com",
				"notification provider outage",
				"operator",
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("Authorize() error = %v, want %v", err, test.want)
			}
			if err != test.want {
				t.Fatalf("Authorize() exposed wrapped store error %v", err)
			}
		})
	}
}

func TestRecoveryServiceValidityDefaultsAllowsShorterAndRejectsLonger(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		configured   time.Duration
		wantValidity time.Duration
		wantErr      error
	}{
		{name: "default", wantValidity: 10 * time.Minute},
		{name: "shorter", configured: 90 * time.Second, wantValidity: 90 * time.Second},
		{name: "negative", configured: -time.Second, wantErr: ErrInvalidRecoveryValidity},
		{name: "longer", configured: 10*time.Minute + time.Nanosecond, wantErr: ErrInvalidRecoveryValidity},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &fakeRecoveryAdministrationStore{}
			service, err := NewRecoveryService(store, RecoveryServiceOptions{
				Validity: test.configured,
				Now:      func() time.Time { return now },
			})
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) || service != nil {
					t.Fatalf("NewRecoveryService() = %#v, %v; want nil, %v", service, err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewRecoveryService() error = %v", err)
			}
			if _, err := service.Authorize(context.Background(), "admin@example.com", "reason", "operator"); err != nil {
				t.Fatalf("Authorize() error = %v", err)
			}
			if got := store.authorizeExpiresAt.Sub(store.authorizeCreatedAt); got != test.wantValidity {
				t.Fatalf("authorization validity = %s, want %s", got, test.wantValidity)
			}
		})
	}
}

func TestRecoveryStatusPassesAuthoritativeTimeAndSupportsNoneAndExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 9, 0, 0, 0, time.FixedZone("test", 4*60*60))
	expiredAt := now.UTC().Add(-time.Minute)
	tests := []struct {
		name   string
		status RecoveryAuthorizationStatus
	}{
		{name: "none", status: RecoveryAuthorizationStatus{State: RecoveryAuthorizationStateNone}},
		{name: "expired", status: RecoveryAuthorizationStatus{
			State: RecoveryAuthorizationStateExpired,
			Authorization: RecoveryAuthorization{
				ID:              "00000000-0000-0000-0000-000000000021",
				SuperAdminEmail: "admin@example.com",
				ExpiredAt:       &expiredAt,
			},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &fakeRecoveryAdministrationStore{statusResult: test.status}
			service := newRecoveryTestService(t, store, RecoveryServiceOptions{
				Now: func() time.Time { return now },
			})
			got, err := service.Status(context.Background(), " ADMIN@example.com ")
			if err != nil {
				t.Fatalf("Status() error = %v", err)
			}
			if got.State != test.status.State || got.Authorization.ID != test.status.Authorization.ID {
				t.Fatalf("Status() = %#v, want %#v", got, test.status)
			}
			if store.statusEmail != "admin@example.com" || !store.statusNow.Equal(now.UTC()) {
				t.Fatalf("status lookup = email %q at %s", store.statusEmail, store.statusNow)
			}
		})
	}
}

func TestRecoveryAuthorizationStateAtUsesTerminalAndExpiryState(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	terminalAt := now.Add(-time.Second)
	base := RecoveryAuthorization{
		ID:        "00000000-0000-0000-0000-000000000023",
		CreatedAt: now.Add(-time.Minute),
		ExpiresAt: now.Add(time.Minute),
	}
	tests := []struct {
		name          string
		authorization RecoveryAuthorization
		want          RecoveryAuthorizationState
	}{
		{name: "no history", want: RecoveryAuthorizationStateNone},
		{name: "active", authorization: base, want: RecoveryAuthorizationStateActive},
		{name: "expiry boundary", authorization: func() RecoveryAuthorization {
			authorization := base
			authorization.ExpiresAt = now
			return authorization
		}(), want: RecoveryAuthorizationStateExpired},
		{name: "explicitly expired", authorization: func() RecoveryAuthorization {
			authorization := base
			authorization.ExpiredAt = &terminalAt
			return authorization
		}(), want: RecoveryAuthorizationStateExpired},
		{name: "consumed", authorization: func() RecoveryAuthorization {
			authorization := base
			authorization.ConsumedAt = &terminalAt
			return authorization
		}(), want: RecoveryAuthorizationStateConsumed},
		{name: "revoked", authorization: func() RecoveryAuthorization {
			authorization := base
			authorization.RevokedAt = &terminalAt
			return authorization
		}(), want: RecoveryAuthorizationStateRevoked},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.authorization.StateAt(now); got != test.want {
				t.Fatalf("StateAt() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRecoveryStatusMapsErrorsAndRejectsUnknownState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		store    *fakeRecoveryAdministrationStore
		email    string
		want     error
		wantCall bool
	}{
		{name: "malformed target", store: &fakeRecoveryAdministrationStore{}, email: "bad", want: ErrRecoveryTargetNotEligible},
		{name: "unknown or inactive target", store: &fakeRecoveryAdministrationStore{statusErr: ErrRecoveryTargetNotEligible}, email: "admin@example.com", want: ErrRecoveryTargetNotEligible, wantCall: true},
		{name: "database failure", store: &fakeRecoveryAdministrationStore{statusErr: errors.New("raw SQL")}, email: "admin@example.com", want: ErrRecoveryUnavailable, wantCall: true},
		{name: "malformed persisted state", store: &fakeRecoveryAdministrationStore{statusResult: RecoveryAuthorizationStatus{State: "BROKEN"}}, email: "admin@example.com", want: ErrRecoveryUnavailable, wantCall: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := newRecoveryTestService(t, test.store, RecoveryServiceOptions{})
			_, err := service.Status(context.Background(), test.email)
			if !errors.Is(err, test.want) {
				t.Fatalf("Status() error = %v, want %v", err, test.want)
			}
			if got := test.store.statusCalls > 0; got != test.wantCall {
				t.Fatalf("status store called = %v, want %v", got, test.wantCall)
			}
		})
	}
}

func TestRecoveryRevokeRequiresAttributionAndMapsLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 14, 0, 0, 0, time.UTC)
	revokedAt := now
	want := RecoveryAuthorization{
		ID:              "00000000-0000-0000-0000-000000000022",
		SuperAdminEmail: "admin@example.com",
		RevokedAt:       &revokedAt,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		store := &fakeRecoveryAdministrationStore{revokeResult: want}
		service := newRecoveryTestService(t, store, RecoveryServiceOptions{Now: func() time.Time { return now }})
		ctx := requestmeta.WithCorrelationID(context.Background(), "revoke-correlation")
		got, err := service.Revoke(ctx, " ADMIN@example.com ", " no longer needed ", " sre-2 ")
		if err != nil {
			t.Fatalf("Revoke() error = %v", err)
		}
		if got.ID != want.ID || store.revokeCalls != 1 || store.revokeEmail != "admin@example.com" {
			t.Fatalf("Revoke() = %#v, store = %#v", got, store)
		}
		if store.revokeReason != "no longer needed" || store.revokeOperator != "sre-2" || store.revokeCorrelation != "revoke-correlation" {
			t.Fatalf("revoke attribution = %q/%q/%q", store.revokeReason, store.revokeOperator, store.revokeCorrelation)
		}
		if !store.revokedAt.Equal(now) {
			t.Fatalf("revoked at = %s, want %s", store.revokedAt, now)
		}
	})

	for _, test := range []struct {
		name     string
		reason   string
		operator string
		storeErr error
		want     error
		wantCall bool
	}{
		{name: "reason required", operator: "operator", want: ErrInvalidRecoveryReason},
		{name: "operator required", reason: "reason", want: ErrInvalidRecoveryOperator},
		{name: "unknown target", reason: "reason", operator: "operator", storeErr: ErrRecoveryTargetNotEligible, want: ErrRecoveryTargetNotEligible, wantCall: true},
		{name: "not active", reason: "reason", operator: "operator", storeErr: ErrRecoveryNotActive, want: ErrRecoveryNotActive, wantCall: true},
		{name: "database failure", reason: "reason", operator: "operator", storeErr: errors.New("raw SQL"), want: ErrRecoveryUnavailable, wantCall: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &fakeRecoveryAdministrationStore{revokeErr: test.storeErr}
			service := newRecoveryTestService(t, store, RecoveryServiceOptions{})
			_, err := service.Revoke(context.Background(), "admin@example.com", test.reason, test.operator)
			if !errors.Is(err, test.want) {
				t.Fatalf("Revoke() error = %v, want %v", err, test.want)
			}
			if got := store.revokeCalls > 0; got != test.wantCall {
				t.Fatalf("revoke store called = %v, want %v", got, test.wantCall)
			}
		})
	}
}

func TestRecoveryUsesSafeCorrelationFallback(t *testing.T) {
	t.Parallel()

	for _, ctx := range []context.Context{
		nil,
		context.Background(),
		requestmeta.WithCorrelationID(context.Background(), "bad\ncorrelation"),
	} {
		store := &fakeRecoveryAdministrationStore{}
		service := newRecoveryTestService(t, store, RecoveryServiceOptions{})
		if _, err := service.Authorize(ctx, "admin@example.com", "reason", "operator"); err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if store.authorizeCorrelation != "unavailable" {
			t.Fatalf("correlation ID = %q, want safe fallback", store.authorizeCorrelation)
		}
	}
}

func TestRecoveryRejectsUnsafeServerCorrelationBeforePersistence(t *testing.T) {
	t.Parallel()

	// Context values cannot be forged through the CLI or HTTP, but validation at
	// this trust boundary still fails closed if an internal caller supplies text
	// that is not safe to persist as a single-line audit identifier.
	ctx := requestmeta.WithCorrelationID(context.Background(), "bad\u2028correlation")
	store := &fakeRecoveryAdministrationStore{}
	service := newRecoveryTestService(t, store, RecoveryServiceOptions{})
	_, err := service.Authorize(ctx, "admin@example.com", "reason", "operator")
	if !errors.Is(err, ErrInvalidRecoveryCorrelationID) {
		t.Fatalf("Authorize() error = %v, want %v", err, ErrInvalidRecoveryCorrelationID)
	}
	if store.authorizeCalls != 0 {
		t.Fatal("unsafe correlation ID reached persistence")
	}
}

func TestNewRecoveryServiceRequiresStoreAndNilReceiverFailsClosed(t *testing.T) {
	t.Parallel()

	if service, err := NewRecoveryService(nil, RecoveryServiceOptions{}); err == nil || service != nil {
		t.Fatalf("NewRecoveryService(nil) = %#v, %v; want error", service, err)
	}

	var service *RecoveryService
	if _, err := service.Authorize(context.Background(), "admin@example.com", "reason", "operator"); !errors.Is(err, ErrRecoveryUnavailable) {
		t.Fatalf("nil Authorize() error = %v", err)
	}
	if _, err := service.Status(context.Background(), "admin@example.com"); !errors.Is(err, ErrRecoveryUnavailable) {
		t.Fatalf("nil Status() error = %v", err)
	}
	if _, err := service.Revoke(context.Background(), "admin@example.com", "reason", "operator"); !errors.Is(err, ErrRecoveryUnavailable) {
		t.Fatalf("nil Revoke() error = %v", err)
	}
}

func newRecoveryTestService(
	t *testing.T,
	store RecoveryAdministrationStore,
	options RecoveryServiceOptions,
) *RecoveryService {
	t.Helper()

	service, err := NewRecoveryService(store, options)
	if err != nil {
		t.Fatalf("NewRecoveryService() error = %v", err)
	}
	return service
}
