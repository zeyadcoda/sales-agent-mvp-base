-- +goose Up

-- Platform Audit is the single append-only source of truth for protected
-- administrative/security mutations. It is intentionally generic so later
-- platform modules can append to the same history without creating parallel
-- audit stores.
CREATE TABLE platform_audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at TIMESTAMPTZ NOT NULL,
    actor_type TEXT NOT NULL,
    actor_identifier TEXT NOT NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_reference TEXT NOT NULL,
    organization_id UUID,
    old_values JSONB,
    new_values JSONB,
    reason TEXT,
    result TEXT NOT NULL,
    correlation_id TEXT NOT NULL,
    CONSTRAINT platform_audit_events_actor_type_format
        CHECK (
            actor_type ~ '^[A-Z][A-Z0-9_]{0,63}$'
        ),
    CONSTRAINT platform_audit_events_actor_identifier_format
        CHECK (
            actor_identifier = BTRIM(actor_identifier)
            AND CHAR_LENGTH(actor_identifier) BETWEEN 1 AND 254
            AND actor_identifier !~ '[[:cntrl:]]'
        ),
    CONSTRAINT platform_audit_events_action_format
        CHECK (
            action ~ '^[A-Z][A-Z0-9_]{0,127}$'
        ),
    CONSTRAINT platform_audit_events_resource_type_format
        CHECK (
            resource_type ~ '^[A-Z][A-Z0-9_]{0,63}$'
        ),
    CONSTRAINT platform_audit_events_resource_reference_format
        CHECK (
            resource_reference = BTRIM(resource_reference)
            AND CHAR_LENGTH(resource_reference) BETWEEN 1 AND 512
            AND resource_reference !~ '[[:cntrl:]]'
        ),
    CONSTRAINT platform_audit_events_old_values_object
        CHECK (
            old_values IS NULL
            OR JSONB_TYPEOF(old_values) = 'object'
        ),
    CONSTRAINT platform_audit_events_new_values_object
        CHECK (
            new_values IS NULL
            OR JSONB_TYPEOF(new_values) = 'object'
        ),
    CONSTRAINT platform_audit_events_reason_format
        CHECK (
            reason IS NULL
            OR (
                reason = BTRIM(reason)
                AND CHAR_LENGTH(reason) BETWEEN 1 AND 1000
                AND reason !~ '[[:cntrl:]]'
            )
        ),
    CONSTRAINT platform_audit_events_recovery_reason_required
        CHECK (
            action NOT LIKE 'SUPER_ADMIN_RECOVERY_%'
            OR reason IS NOT NULL
        ),
    CONSTRAINT platform_audit_events_result_known
        CHECK (result IN ('SUCCESS', 'FAILURE')),
    CONSTRAINT platform_audit_events_correlation_id_format
        CHECK (
            correlation_id = BTRIM(correlation_id)
            AND CHAR_LENGTH(correlation_id) BETWEEN 1 AND 128
            AND correlation_id !~ '[[:cntrl:]]'
        )
);

CREATE INDEX platform_audit_events_resource_timeline_idx
    ON platform_audit_events (
        resource_type,
        resource_reference,
        occurred_at DESC,
        id DESC
    );

CREATE INDEX platform_audit_events_organization_timeline_idx
    ON platform_audit_events (organization_id, occurred_at DESC, id DESC)
    WHERE organization_id IS NOT NULL;

-- Application behavior has no update/delete operation for AuditEvent. The
-- trigger provides a database-level final guard even if future application
-- code accidentally obtains a broad query executor.
-- +goose StatementBegin
CREATE FUNCTION platform_audit_events_reject_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'platform audit events are append-only'
        USING ERRCODE = '55000';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER platform_audit_events_immutable
    BEFORE UPDATE OR DELETE ON platform_audit_events
    FOR EACH ROW
    EXECUTE FUNCTION platform_audit_events_reject_mutation();

-- TRUNCATE does not fire row-level DELETE triggers. Reject it separately so a
-- broad application query executor cannot erase the immutable Audit source of
-- truth with a table-level operation.
CREATE TRIGGER platform_audit_events_immutable_truncate
    BEFORE TRUNCATE ON platform_audit_events
    FOR EACH STATEMENT
    EXECUTE FUNCTION platform_audit_events_reject_mutation();

-- Recovery authorization is server-side deployment state only. Its UUID is an
-- internal database reference, not a browser identifier or user-entered token.
CREATE TABLE super_admin_recovery_authorizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    super_admin_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    reason TEXT NOT NULL,
    operator_identifier TEXT NOT NULL,
    correlation_id TEXT NOT NULL,
    CONSTRAINT super_admin_recovery_authorizations_admin_fk
        FOREIGN KEY (super_admin_id)
        REFERENCES super_admin_accounts (id)
        ON DELETE RESTRICT,
    CONSTRAINT super_admin_recovery_authorizations_expiry_after_creation
        CHECK (expires_at > created_at),
    CONSTRAINT super_admin_recovery_authorizations_maximum_lifetime
        CHECK (expires_at <= created_at + INTERVAL '10 minutes'),
    CONSTRAINT super_admin_recovery_authorizations_terminal_state
        CHECK (NUM_NONNULLS(consumed_at, revoked_at, expired_at) <= 1),
    CONSTRAINT super_admin_recovery_authorizations_consumed_in_lifetime
        CHECK (
            consumed_at IS NULL
            OR (
                consumed_at >= created_at
                AND consumed_at < expires_at
            )
        ),
    CONSTRAINT super_admin_recovery_authorizations_revoked_after_creation
        CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CONSTRAINT super_admin_recovery_authorizations_expired_at_expiry
        CHECK (expired_at IS NULL OR expired_at >= expires_at),
    CONSTRAINT super_admin_recovery_authorizations_reason_format
        CHECK (
            reason = BTRIM(reason)
            AND CHAR_LENGTH(reason) BETWEEN 1 AND 1000
            AND reason !~ '[[:cntrl:]]'
        ),
    CONSTRAINT super_admin_recovery_authorizations_operator_format
        CHECK (
            operator_identifier = BTRIM(operator_identifier)
            AND CHAR_LENGTH(operator_identifier) BETWEEN 1 AND 254
            AND operator_identifier !~ '[[:cntrl:]]'
        ),
    CONSTRAINT super_admin_recovery_authorizations_correlation_id_format
        CHECK (
            correlation_id = BTRIM(correlation_id)
            AND CHAR_LENGTH(correlation_id) BETWEEN 1 AND 128
            AND correlation_id !~ '[[:cntrl:]]'
        )
);

-- PostgreSQL cannot use the current time in an immutable partial-index
-- predicate. The application therefore records expired_at for elapsed grants
-- before creating a replacement. This index then gives a final database
-- guarantee that no Super Admin has two unresolved authorizations.
CREATE UNIQUE INDEX super_admin_recovery_authorizations_unresolved_admin_idx
    ON super_admin_recovery_authorizations (super_admin_id)
    WHERE consumed_at IS NULL
      AND revoked_at IS NULL
      AND expired_at IS NULL;

-- Active reads must additionally require expires_at > the trusted operation
-- timestamp; unresolved alone is never proof that authorization is usable.
CREATE INDEX super_admin_recovery_authorizations_admin_expiry_idx
    ON super_admin_recovery_authorizations (super_admin_id, expires_at)
    WHERE consumed_at IS NULL
      AND revoked_at IS NULL
      AND expired_at IS NULL;

CREATE INDEX super_admin_recovery_authorizations_expiry_idx
    ON super_admin_recovery_authorizations (expires_at)
    WHERE consumed_at IS NULL
      AND revoked_at IS NULL
      AND expired_at IS NULL;

-- +goose Down

DROP TABLE super_admin_recovery_authorizations;
DROP TABLE platform_audit_events;
DROP FUNCTION platform_audit_events_reject_mutation();
