package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/audit"
	"salesagent.local/backend/internal/platform/auth"
	"salesagent.local/backend/internal/requestmeta"
)

type fakeRecoveryAdministrator struct {
	authorizeEmail       string
	authorizeReason      string
	authorizeOperator    string
	authorizeCorrelation string
	authorizeResult      auth.RecoveryAuthorization
	authorizeErr         error
	authorizeCalls       int

	statusEmail  string
	statusResult auth.RecoveryAuthorizationStatus
	statusErr    error
	statusCalls  int

	revokeEmail       string
	revokeReason      string
	revokeOperator    string
	revokeCorrelation string
	revokeResult      auth.RecoveryAuthorization
	revokeErr         error
	revokeCalls       int
}

func (recovery *fakeRecoveryAdministrator) Authorize(
	ctx context.Context,
	email string,
	reason string,
	operatorIdentifier string,
) (auth.RecoveryAuthorization, error) {
	recovery.authorizeCalls++
	recovery.authorizeEmail = email
	recovery.authorizeReason = reason
	recovery.authorizeOperator = operatorIdentifier
	recovery.authorizeCorrelation = requestmeta.CorrelationID(ctx)
	return recovery.authorizeResult, recovery.authorizeErr
}

func (recovery *fakeRecoveryAdministrator) Status(
	_ context.Context,
	email string,
) (auth.RecoveryAuthorizationStatus, error) {
	recovery.statusCalls++
	recovery.statusEmail = email
	return recovery.statusResult, recovery.statusErr
}

func (recovery *fakeRecoveryAdministrator) Revoke(
	ctx context.Context,
	email string,
	reason string,
	operatorIdentifier string,
) (auth.RecoveryAuthorization, error) {
	recovery.revokeCalls++
	recovery.revokeEmail = email
	recovery.revokeReason = reason
	recovery.revokeOperator = operatorIdentifier
	recovery.revokeCorrelation = requestmeta.CorrelationID(ctx)
	return recovery.revokeResult, recovery.revokeErr
}

type fakeRecoveryAuditReader struct {
	resourceType      audit.ResourceType
	resourceReference string
	limit             int
	events            []audit.Event
	err               error
	calls             int
}

func (reader *fakeRecoveryAuditReader) ListByResource(
	_ context.Context,
	resourceType audit.ResourceType,
	resourceReference string,
	limit int,
) ([]audit.Event, error) {
	reader.calls++
	reader.resourceType = resourceType
	reader.resourceReference = resourceReference
	reader.limit = limit
	return reader.events, reader.err
}

type fakePlatformAdminFactory struct {
	services platformAdminServices
	err      error
	calls    int
	closed   int
}

func (factory *fakePlatformAdminFactory) open(_ context.Context) (platformAdminServices, error) {
	factory.calls++
	services := factory.services
	if services.close == nil {
		services.close = func() { factory.closed++ }
	}
	return services, factory.err
}

func TestAuthorizeRequiresAttributionAndExplicitConfirmation(t *testing.T) {
	t.Parallel()

	baseArgs := []string{
		"auth-recovery", "authorize",
		"--email", " ADMIN@example.com ",
		"--reason", " notification provider outage ",
		"--operator", " on-call-sre ",
	}

	t.Run("missing reason", func(t *testing.T) {
		t.Parallel()

		factory := &fakePlatformAdminFactory{}
		err := run(
			context.Background(),
			[]string{"auth-recovery", "authorize", "--email", "admin@example.com", "--operator", "sre"},
			strings.NewReader("y\n"),
			&strings.Builder{},
			factory.open,
		)
		if err == nil || !strings.Contains(err.Error(), "--reason is required") {
			t.Fatalf("run() error = %v, want required reason", err)
		}
		if factory.calls != 0 {
			t.Fatal("invalid command opened deployment services")
		}
	})

	t.Run("default no", func(t *testing.T) {
		t.Parallel()

		factory := &fakePlatformAdminFactory{}
		var output strings.Builder
		if err := run(context.Background(), baseArgs, strings.NewReader("\n"), &output, factory.open); err != nil {
			t.Fatalf("run() error = %v", err)
		}
		if factory.calls != 0 || !strings.Contains(output.String(), "Confirm? [y/N]") ||
			!strings.Contains(output.String(), "no recovery authorization was created") {
			t.Fatalf("declined command = calls %d, output %q", factory.calls, output.String())
		}
	})

	t.Run("confirmed", func(t *testing.T) {
		t.Parallel()

		expiresAt := time.Date(2026, 8, 20, 11, 10, 0, 0, time.UTC)
		recovery := &fakeRecoveryAdministrator{authorizeResult: auth.RecoveryAuthorization{
			ID:              "internal-authorization-id-must-not-print",
			SuperAdminEmail: "admin@example.com",
			ExpiresAt:       expiresAt,
		}}
		factory := &fakePlatformAdminFactory{services: platformAdminServices{recovery: recovery}}
		var output strings.Builder
		if err := run(context.Background(), baseArgs, strings.NewReader("yes\n"), &output, factory.open); err != nil {
			t.Fatalf("run() error = %v", err)
		}
		if recovery.authorizeCalls != 1 || recovery.authorizeEmail != "admin@example.com" ||
			recovery.authorizeReason != "notification provider outage" ||
			recovery.authorizeOperator != "on-call-sre" {
			t.Fatalf("Authorize() input = %#v", recovery)
		}
		if len(recovery.authorizeCorrelation) != 32 {
			t.Fatalf("correlation ID = %q, want 32 hex characters", recovery.authorizeCorrelation)
		}
		printed := output.String()
		if !strings.Contains(printed, "next successful password login bypasses email OTP once") ||
			!strings.Contains(printed, expiresAt.Format(time.RFC3339)) {
			t.Fatalf("impact/success output = %q", printed)
		}
		if strings.Contains(printed, recovery.authorizeResult.ID) {
			t.Fatal("command printed an internal recovery authorization identifier")
		}
		if factory.calls != 1 || factory.closed != 1 {
			t.Fatalf("factory lifecycle = calls %d, closed %d", factory.calls, factory.closed)
		}
	})
}

func TestAuthorizeRejectsPasswordOrNonInteractiveOverrideFlags(t *testing.T) {
	t.Parallel()

	for _, forbidden := range []string{"--password", "--yes"} {
		factory := &fakePlatformAdminFactory{}
		err := run(
			context.Background(),
			[]string{
				"auth-recovery", "authorize", "--email", "admin@example.com",
				"--reason", "provider outage", "--operator", "sre", forbidden, "unsafe",
			},
			strings.NewReader("y\n"),
			&strings.Builder{},
			factory.open,
		)
		if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
			t.Fatalf("run(%s) error = %v", forbidden, err)
		}
		if factory.calls != 0 {
			t.Fatalf("forbidden flag %s opened deployment services", forbidden)
		}
	}
}

func TestAuthorizeDuplicateAndUnknownBehaviorIsClearAndSanitized(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		storeErr error
		want     string
	}{
		{name: "unknown", storeErr: auth.ErrRecoveryTargetNotEligible, want: "not found or is not eligible"},
		{name: "duplicate", storeErr: auth.ErrRecoveryAlreadyActive, want: "already exists"},
		{name: "database", storeErr: errors.New("raw SQL credential=must-not-print"), want: "could not create"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			recovery := &fakeRecoveryAdministrator{authorizeErr: test.storeErr}
			factory := &fakePlatformAdminFactory{services: platformAdminServices{recovery: recovery}}
			err := run(
				context.Background(),
				[]string{
					"auth-recovery", "authorize", "--email", "admin@example.com",
					"--reason", "provider outage", "--operator", "sre",
				},
				strings.NewReader("y\n"),
				&strings.Builder{},
				factory.open,
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("run() error = %v, want %q", err, test.want)
			}
			if strings.Contains(err.Error(), "raw SQL") || strings.Contains(err.Error(), "credential") {
				t.Fatalf("run() leaked infrastructure detail: %v", err)
			}
		})
	}
}

func TestStatusAndRevokeAreDeploymentOnlyAndClear(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(10 * time.Minute)
	recovery := &fakeRecoveryAdministrator{statusResult: auth.RecoveryAuthorizationStatus{
		State: auth.RecoveryAuthorizationStateExpired,
		Authorization: auth.RecoveryAuthorization{
			SuperAdminEmail:    "admin@example.com",
			CreatedAt:          createdAt,
			ExpiresAt:          expiresAt,
			OperatorIdentifier: "sre-1",
			Reason:             "provider outage",
		},
	}}
	factory := &fakePlatformAdminFactory{services: platformAdminServices{recovery: recovery}}
	var statusOutput strings.Builder
	if err := run(
		context.Background(),
		[]string{"auth-recovery", "status", "--email", "ADMIN@example.com"},
		strings.NewReader(""),
		&statusOutput,
		factory.open,
	); err != nil {
		t.Fatalf("status run() error = %v", err)
	}
	if recovery.statusEmail != "admin@example.com" ||
		!strings.Contains(statusOutput.String(), "Recovery status: EXPIRED") ||
		!strings.Contains(statusOutput.String(), "provider outage") {
		t.Fatalf("status output = %q", statusOutput.String())
	}

	t.Run("revoke defaults no", func(t *testing.T) {
		t.Parallel()
		localRecovery := &fakeRecoveryAdministrator{}
		localFactory := &fakePlatformAdminFactory{services: platformAdminServices{recovery: localRecovery}}
		if err := run(
			context.Background(),
			[]string{
				"auth-recovery", "revoke", "--email", "admin@example.com",
				"--reason", "provider restored", "--operator", "sre-2",
			},
			strings.NewReader("n\n"),
			&strings.Builder{},
			localFactory.open,
		); err != nil {
			t.Fatalf("revoke run() error = %v", err)
		}
		if localRecovery.revokeCalls != 0 || localFactory.calls != 0 {
			t.Fatal("declined revocation reached deployment services")
		}
	})

	t.Run("confirmed revoke", func(t *testing.T) {
		t.Parallel()
		localRecovery := &fakeRecoveryAdministrator{}
		localFactory := &fakePlatformAdminFactory{services: platformAdminServices{recovery: localRecovery}}
		var output strings.Builder
		if err := run(
			context.Background(),
			[]string{
				"auth-recovery", "revoke", "--email", "admin@example.com",
				"--reason", "provider restored", "--operator", "sre-2",
			},
			strings.NewReader("y\n"),
			&output,
			localFactory.open,
		); err != nil {
			t.Fatalf("revoke run() error = %v", err)
		}
		if localRecovery.revokeCalls != 1 || localRecovery.revokeReason != "provider restored" ||
			localRecovery.revokeOperator != "sre-2" || len(localRecovery.revokeCorrelation) != 32 {
			t.Fatalf("Revoke() input = %#v", localRecovery)
		}
		if !strings.Contains(output.String(), "was revoked") {
			t.Fatalf("revoke output = %q", output.String())
		}
	})
}

func TestAuditCommandListsSafeRecoveryEventsWithoutSQL(t *testing.T) {
	t.Parallel()

	reason := "notification provider outage"
	reader := &fakeRecoveryAuditReader{events: []audit.Event{
		{
			OccurredAt:        time.Date(2026, 8, 20, 13, 0, 0, 0, time.UTC),
			ActorType:         audit.ActorTypeDeploymentOperator,
			ActorIdentifier:   "sre-1",
			Action:            audit.ActionSuperAdminRecoveryAuthorized,
			ResourceType:      audit.ResourceTypeSuperAdminAccount,
			ResourceReference: "admin@example.com",
			Reason:            &reason,
			Result:            audit.ResultSuccess,
			CorrelationID:     "safe-correlation",
		},
	}}
	factory := &fakePlatformAdminFactory{services: platformAdminServices{audit: reader}}
	var output strings.Builder
	if err := run(
		context.Background(),
		[]string{"auth-recovery", "audit", "--email", "ADMIN@example.com"},
		strings.NewReader(""),
		&output,
		factory.open,
	); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if reader.calls != 1 || reader.resourceType != audit.ResourceTypeSuperAdminAccount ||
		reader.resourceReference != "admin@example.com" || reader.limit != recoveryAuditListLimit {
		t.Fatalf("audit filter = %#v", reader)
	}
	printed := output.String()
	if !strings.Contains(printed, string(audit.ActionSuperAdminRecoveryAuthorized)) ||
		!strings.Contains(printed, "sre-1") ||
		!strings.Contains(printed, reason) ||
		!strings.Contains(printed, "safe-correlation") {
		t.Fatalf("audit output = %q", printed)
	}
	if strings.Contains(printed, "SELECT ") {
		t.Fatal("audit command exposed SQL")
	}
}

func TestCommandRequiresOnlyKnownCLIShape(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		nil,
		{"auth-recovery"},
		{"unknown", "authorize"},
		{"auth-recovery", "unknown"},
	} {
		factory := &fakePlatformAdminFactory{}
		if err := run(context.Background(), args, strings.NewReader(""), &strings.Builder{}, factory.open); err == nil {
			t.Fatalf("run(%v) error = nil", args)
		}
		if factory.calls != 0 {
			t.Fatalf("run(%v) opened deployment services", args)
		}
	}
}
