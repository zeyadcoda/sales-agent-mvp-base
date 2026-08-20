package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"salesagent.local/backend/internal/database"
)

func TestPostgresAuditAppendAndListRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping platform audit integration test")
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

	reason := "notification provider outage"
	resourceReference := fmt.Sprintf("audit-integration-%d@example.com", time.Now().UnixNano())
	occurredAt := time.Now().UTC().Truncate(time.Microsecond)
	if err := Append(ctx, db, Event{
		OccurredAt:        occurredAt,
		ActorType:         ActorTypeDeploymentOperator,
		ActorIdentifier:   "audit-integration-operator",
		Action:            ActionSuperAdminRecoveryAuthorized,
		ResourceType:      ResourceTypeSuperAdminAccount,
		ResourceReference: resourceReference,
		NewValues: json.RawMessage(`{
			"recovery_authorization_id":"00000000-0000-0000-0000-000000000099",
			"status":"ACTIVE"
		}`),
		Reason:        &reason,
		Result:        ResultSuccess,
		CorrelationID: "audit-integration-correlation",
	}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	events, err := store.ListByResource(
		ctx,
		ResourceTypeSuperAdminAccount,
		resourceReference,
		10,
	)
	if err != nil {
		t.Fatalf("ListByResource() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("ListByResource() event count = %d, want 1", len(events))
	}
	loaded := events[0]
	if loaded.ID == "" ||
		loaded.Action != ActionSuperAdminRecoveryAuthorized ||
		loaded.ResourceReference != resourceReference ||
		loaded.Reason == nil || *loaded.Reason != reason ||
		!loaded.OccurredAt.Equal(occurredAt) {
		t.Fatalf("ListByResource() event = %#v", loaded)
	}
	if loaded.OldValues != nil || !json.Valid(loaded.NewValues) {
		t.Fatalf("ListByResource() structured values = old %q, new %q", loaded.OldValues, loaded.NewValues)
	}

	if _, err := db.Exec(
		ctx,
		`UPDATE platform_audit_events SET result = 'FAILURE' WHERE id = $1::uuid`,
		loaded.ID,
	); err == nil {
		t.Fatal("database permitted platform audit update")
	}
	if _, err := db.Exec(
		ctx,
		`DELETE FROM platform_audit_events WHERE id = $1::uuid`,
		loaded.ID,
	); err == nil {
		t.Fatal("database permitted platform audit delete")
	}
}
