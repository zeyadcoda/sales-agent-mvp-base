package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor is deliberately small so a pgx transaction can append the audit
// event in the exact transaction that protects the associated mutation.
type Executor interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

// QueryRower is the read-only capability required by the deployment CLI's
// bounded audit verification command.
type QueryRower interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

// Append writes exactly one immutable platform audit event. It intentionally
// exposes no update or delete capability.
func Append(ctx context.Context, executor Executor, event Event) error {
	if executor == nil {
		return ErrAuditStoreRequired
	}
	if err := validateEvent(event); err != nil {
		return err
	}

	const query = `
		INSERT INTO platform_audit_events (
			occurred_at,
			actor_type,
			actor_identifier,
			action,
			resource_type,
			resource_reference,
			organization_id,
			old_values,
			new_values,
			reason,
			result,
			correlation_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7::uuid,
			$8::jsonb,
			$9::jsonb,
			$10,
			$11,
			$12
		)
	`

	result, err := executor.Exec(
		ctx,
		query,
		event.OccurredAt.UTC(),
		string(event.ActorType),
		event.ActorIdentifier,
		string(event.Action),
		string(event.ResourceType),
		event.ResourceReference,
		nullableString(event.OrganizationID),
		nullableJSON(event.OldValues),
		nullableJSON(event.NewValues),
		nullableString(event.Reason),
		string(event.Result),
		event.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("append platform audit event: %w", err)
	}
	if result.RowsAffected() != 1 {
		return errors.New("append platform audit event affected an unexpected number of rows")
	}

	return nil
}

// PostgresStore provides the sole minimal audit read needed by the deployment
// CLI. It does not expose arbitrary filters or any mutation operation.
type PostgresStore struct {
	db QueryRower
}

func NewPostgresStore(db QueryRower) (*PostgresStore, error) {
	if db == nil {
		return nil, ErrAuditStoreRequired
	}

	return &PostgresStore{db: db}, nil
}

// ListByResource returns at most 100 newest-first events for one exact resource.
// Recovery callers use SUPER_ADMIN_ACCOUNT and the normalized account email, so
// verification never needs dynamic JSON filtering.
func (store *PostgresStore) ListByResource(
	ctx context.Context,
	resourceType ResourceType,
	resourceReference string,
	limit int,
) ([]Event, error) {
	if store == nil || store.db == nil {
		return nil, ErrAuditStoreRequired
	}
	if err := validateListInput(resourceType, resourceReference, limit); err != nil {
		return nil, err
	}

	const query = `
		SELECT COALESCE(
			JSONB_AGG(
				events.payload
				ORDER BY events.occurred_at DESC, events.id DESC
			),
			'[]'::jsonb
		)
		FROM (
			SELECT
				id,
				occurred_at,
				JSONB_BUILD_OBJECT(
					'id', id::text,
					'occurred_at', occurred_at,
					'actor_type', actor_type,
					'actor_identifier', actor_identifier,
					'action', action,
					'resource_type', resource_type,
					'resource_reference', resource_reference,
					'organization_id', organization_id::text,
					'old_values', old_values,
					'new_values', new_values,
					'reason', reason,
					'result', result,
					'correlation_id', correlation_id
				) AS payload
			FROM platform_audit_events
			WHERE resource_type = $1
			  AND resource_reference = $2
			ORDER BY occurred_at DESC, id DESC
			LIMIT $3
		) AS events
	`

	var encoded []byte
	if err := store.db.QueryRow(
		ctx,
		query,
		string(resourceType),
		resourceReference,
		limit,
	).Scan(&encoded); err != nil {
		return nil, fmt.Errorf("list platform audit events: %w", err)
	}

	var events []Event
	if err := json.Unmarshal(encoded, &events); err != nil {
		return nil, errors.New("decode platform audit events")
	}
	for index := range events {
		events[index].OldValues = normalizeNullableJSON(events[index].OldValues)
		events[index].NewValues = normalizeNullableJSON(events[index].NewValues)
		if err := validateLoadedEvent(events[index]); err != nil {
			// Fail closed rather than displaying malformed or credential-bearing
			// structured values inserted outside the trusted append boundary.
			return nil, errors.New("decode platform audit events")
		}
	}
	if events == nil {
		events = []Event{}
	}

	return events, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}

	return string(value)
}

func normalizeNullableJSON(value json.RawMessage) json.RawMessage {
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil
	}

	return value
}
