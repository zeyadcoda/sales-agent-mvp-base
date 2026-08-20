package audit

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeExecutor struct {
	query string
	args  []any
	tag   pgconn.CommandTag
	err   error
}

func (executor *fakeExecutor) Exec(
	_ context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {
	executor.query = query
	executor.args = append([]any(nil), args...)
	return executor.tag, executor.err
}

type fakeQueryRower struct {
	query string
	args  []any
	row   pgx.Row
}

func (queryer *fakeQueryRower) QueryRow(
	_ context.Context,
	query string,
	args ...any,
) pgx.Row {
	queryer.query = query
	queryer.args = append([]any(nil), args...)
	return queryer.row
}

type encodedAuditRow struct {
	encoded []byte
	err     error
}

func (row encodedAuditRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 1 {
		return errors.New("unexpected destination count")
	}
	destination, ok := destinations[0].(*[]byte)
	if !ok {
		return errors.New("unexpected destination type")
	}
	*destination = append([]byte(nil), row.encoded...)
	return nil
}

func TestAppendWritesOneParameterizedImmutableEvent(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{tag: pgconn.NewCommandTag("INSERT 0 1")}
	event := validRecoveryAuditEvent()
	event.NewValues = json.RawMessage(`{
		"recovery_authorization_id":"00000000-0000-0000-0000-000000000020",
		"authorization_correlation_id":"recovery-correlation",
		"operator_identifier":"deployment-on-call"
	}`)

	if err := Append(context.Background(), executor, event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if !strings.Contains(executor.query, "INSERT INTO platform_audit_events") {
		t.Fatalf("Append() query = %q", executor.query)
	}
	if strings.Contains(executor.query, event.ResourceReference) ||
		strings.Contains(executor.query, *event.Reason) ||
		strings.Contains(executor.query, event.ActorIdentifier) {
		t.Fatal("Append() interpolated event data into SQL")
	}
	if len(executor.args) != 12 {
		t.Fatalf("Append() argument count = %d, want 12", len(executor.args))
	}
	if executor.args[3] != string(ActionSuperAdminRecoveryAuthorized) ||
		executor.args[4] != string(ResourceTypeSuperAdminAccount) ||
		executor.args[5] != "admin@example.com" ||
		executor.args[10] != string(ResultSuccess) {
		t.Fatalf("Append() arguments = %#v", executor.args)
	}
	if executor.args[6] != nil || executor.args[7] != nil {
		t.Fatalf("nullable Append() arguments = %#v", executor.args[6:8])
	}
}

func TestAppendRejectsUnsafeOrIncompleteEventsBeforePersistence(t *testing.T) {
	t.Parallel()

	reason := "notification provider outage"
	tests := []struct {
		name   string
		mutate func(*Event)
	}{
		{
			name: "database ID is output only",
			mutate: func(event *Event) {
				event.ID = "00000000-0000-0000-0000-000000000001"
			},
		},
		{
			name: "recovery reason required",
			mutate: func(event *Event) {
				event.Reason = nil
			},
		},
		{
			name: "reason cannot be blank",
			mutate: func(event *Event) {
				blank := "  "
				event.Reason = &blank
			},
		},
		{
			name: "structured values must be objects",
			mutate: func(event *Event) {
				event.NewValues = json.RawMessage(`["not", "an", "object"]`)
			},
		},
		{
			name: "nested password prohibited",
			mutate: func(event *Event) {
				event.NewValues = json.RawMessage(`{"safe":{"password":"must-not-persist"}}`)
			},
		},
		{
			name: "OTP hash prohibited",
			mutate: func(event *Event) {
				event.OldValues = json.RawMessage(`{"otp_hash":"must-not-persist"}`)
			},
		},
		{
			name: "unknown result",
			mutate: func(event *Event) {
				event.Result = "MAYBE"
			},
		},
		{
			name: "control character",
			mutate: func(event *Event) {
				event.ActorIdentifier = "operator\nspoof"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			event := validRecoveryAuditEvent()
			event.Reason = &reason
			test.mutate(&event)
			executor := &fakeExecutor{tag: pgconn.NewCommandTag("INSERT 0 1")}

			err := Append(context.Background(), executor, event)
			if !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("Append() error = %v, want %v", err, ErrInvalidEvent)
			}
			if executor.query != "" || len(executor.args) != 0 {
				t.Fatal("invalid event reached persistence")
			}
		})
	}
}

func TestAppendRejectsMissingOrFailedPersistence(t *testing.T) {
	t.Parallel()

	if err := Append(context.Background(), nil, validRecoveryAuditEvent()); !errors.Is(err, ErrAuditStoreRequired) {
		t.Fatalf("Append(nil) error = %v, want %v", err, ErrAuditStoreRequired)
	}

	databaseError := errors.New("simulated database failure")
	failed := &fakeExecutor{err: databaseError}
	if err := Append(context.Background(), failed, validRecoveryAuditEvent()); !errors.Is(err, databaseError) {
		t.Fatalf("Append(database failure) error = %v", err)
	}

	noRows := &fakeExecutor{tag: pgconn.NewCommandTag("INSERT 0 0")}
	if err := Append(context.Background(), noRows, validRecoveryAuditEvent()); err == nil {
		t.Fatal("Append() accepted an insert that affected no rows")
	}
}

func TestPostgresStoreListByResourceReturnsNewestAuditEvents(t *testing.T) {
	t.Parallel()

	encoded := []byte(`[
		{
			"id":"00000000-0000-0000-0000-000000000002",
			"occurred_at":"2026-08-20T10:01:00Z",
			"actor_type":"SUPER_ADMIN",
			"actor_identifier":"admin@example.com",
			"action":"SUPER_ADMIN_RECOVERY_CONSUMED",
			"resource_type":"SUPER_ADMIN_ACCOUNT",
			"resource_reference":"admin@example.com",
			"organization_id":null,
			"old_values":null,
			"new_values":{"status":"CONSUMED"},
			"reason":"notification provider outage",
			"result":"SUCCESS",
			"correlation_id":"recovery-correlation"
		},
		{
			"id":"00000000-0000-0000-0000-000000000001",
			"occurred_at":"2026-08-20T10:00:00Z",
			"actor_type":"DEPLOYMENT_OPERATOR",
			"actor_identifier":"deployment-on-call",
			"action":"SUPER_ADMIN_RECOVERY_AUTHORIZED",
			"resource_type":"SUPER_ADMIN_ACCOUNT",
			"resource_reference":"admin@example.com",
			"organization_id":null,
			"old_values":null,
			"new_values":{"status":"ACTIVE"},
			"reason":"notification provider outage",
			"result":"SUCCESS",
			"correlation_id":"recovery-correlation"
		}
	]`)
	queryer := &fakeQueryRower{row: encodedAuditRow{encoded: encoded}}
	store, err := NewPostgresStore(queryer)
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}

	events, err := store.ListByResource(
		context.Background(),
		ResourceTypeSuperAdminAccount,
		"admin@example.com",
		25,
	)
	if err != nil {
		t.Fatalf("ListByResource() error = %v", err)
	}
	if len(events) != 2 || events[0].Action != ActionSuperAdminRecoveryConsumed {
		t.Fatalf("ListByResource() events = %#v", events)
	}
	if len(events[0].OldValues) != 0 || string(events[0].NewValues) != `{"status":"CONSUMED"}` {
		t.Fatalf("ListByResource() structured values = old %q, new %q", events[0].OldValues, events[0].NewValues)
	}
	if !strings.Contains(queryer.query, "WHERE resource_type = $1") ||
		strings.Contains(queryer.query, "admin@example.com") {
		t.Fatalf("ListByResource() query is not safely fixed/parameterized: %q", queryer.query)
	}
	if len(queryer.args) != 3 ||
		queryer.args[0] != string(ResourceTypeSuperAdminAccount) ||
		queryer.args[1] != "admin@example.com" ||
		queryer.args[2] != 25 {
		t.Fatalf("ListByResource() arguments = %#v", queryer.args)
	}
}

func TestPostgresStoreListByResourceValidatesAndReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	queryer := &fakeQueryRower{row: encodedAuditRow{encoded: []byte(`[]`)}}
	store, err := NewPostgresStore(queryer)
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}

	events, err := store.ListByResource(
		context.Background(),
		ResourceTypeSuperAdminAccount,
		"admin@example.com",
		1,
	)
	if err != nil {
		t.Fatalf("ListByResource() error = %v", err)
	}
	if events == nil || len(events) != 0 {
		t.Fatalf("ListByResource() empty events = %#v", events)
	}

	for _, limit := range []int{0, 101} {
		queryer.query = ""
		_, err := store.ListByResource(
			context.Background(),
			ResourceTypeSuperAdminAccount,
			"admin@example.com",
			limit,
		)
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("ListByResource(limit=%d) error = %v", limit, err)
		}
		if queryer.query != "" {
			t.Fatalf("invalid list limit %d reached persistence", limit)
		}
	}
}

func TestPostgresStoreListByResourceFailsClosedForUnsafePersistedValues(t *testing.T) {
	t.Parallel()

	encoded := []byte(`[{
		"id":"00000000-0000-0000-0000-000000000003",
		"occurred_at":"2026-08-20T10:01:00Z",
		"actor_type":"SUPER_ADMIN",
		"actor_identifier":"admin@example.com",
		"action":"SUPER_ADMIN_RECOVERY_CONSUMED",
		"resource_type":"SUPER_ADMIN_ACCOUNT",
		"resource_reference":"admin@example.com",
		"organization_id":null,
		"old_values":null,
		"new_values":{"session_token":"must-not-display"},
		"reason":"notification provider outage",
		"result":"SUCCESS",
		"correlation_id":"recovery-correlation"
	}]`)
	store, err := NewPostgresStore(&fakeQueryRower{row: encodedAuditRow{encoded: encoded}})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}

	events, err := store.ListByResource(
		context.Background(),
		ResourceTypeSuperAdminAccount,
		"admin@example.com",
		10,
	)
	if err == nil || events != nil {
		t.Fatalf("ListByResource() = (%#v, %v), want fail-closed error", events, err)
	}
	if strings.Contains(err.Error(), "must-not-display") {
		t.Fatalf("ListByResource() leaked unsafe persisted content: %v", err)
	}
}

func validRecoveryAuditEvent() Event {
	reason := "notification provider outage"
	return Event{
		OccurredAt:        time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC),
		ActorType:         ActorTypeDeploymentOperator,
		ActorIdentifier:   "deployment-on-call",
		Action:            ActionSuperAdminRecoveryAuthorized,
		ResourceType:      ResourceTypeSuperAdminAccount,
		ResourceReference: "admin@example.com",
		NewValues:         json.RawMessage(`{"status":"ACTIVE"}`),
		Reason:            &reason,
		Result:            ResultSuccess,
		CorrelationID:     "recovery-correlation",
	}
}
