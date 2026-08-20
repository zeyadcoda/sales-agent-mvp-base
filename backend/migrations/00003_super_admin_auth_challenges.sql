-- +goose Up

CREATE TABLE super_admin_auth_challenges (
    -- The browser sees this identifier. It is independent 256-bit random
    -- material, not an account ID, session token, or sequential database key.
    id TEXT PRIMARY KEY,
    super_admin_id UUID NOT NULL,
    -- A challenge-bound HMAC-SHA256 digest is stored instead of the six-digit
    -- code. The HMAC key remains in server-side secret management.
    otp_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    failed_attempts SMALLINT NOT NULL DEFAULT 0,
    resend_available_at TIMESTAMPTZ NOT NULL,
    -- Delivery is a two-phase state transition. A version is inactive while
    -- email delivery is pending and becomes usable only after delivery succeeds.
    delivery_version INTEGER NOT NULL DEFAULT 1,
    active_version INTEGER,
    delivery_started_at TIMESTAMPTZ NOT NULL,
    activated_at TIMESTAMPTZ,
    consumed_at TIMESTAMPTZ,
    invalidated_at TIMESTAMPTZ,
    CONSTRAINT super_admin_auth_challenges_admin_fk
        FOREIGN KEY (super_admin_id)
        REFERENCES super_admin_accounts (id)
        ON DELETE RESTRICT,
    CONSTRAINT super_admin_auth_challenges_id_format
        CHECK (
            CHAR_LENGTH(id) = 43
            AND id ~ '^[A-Za-z0-9_-]{43}$'
        ),
    CONSTRAINT super_admin_auth_challenges_otp_hash_sha256
        CHECK (OCTET_LENGTH(otp_hash) = 32),
    CONSTRAINT super_admin_auth_challenges_attempts
        CHECK (failed_attempts BETWEEN 0 AND 5),
    CONSTRAINT super_admin_auth_challenges_delivery_version
        CHECK (
            delivery_version >= 1
            AND (active_version IS NULL OR active_version = delivery_version)
        ),
    CONSTRAINT super_admin_auth_challenges_activation_consistent
        CHECK ((active_version IS NULL) = (activated_at IS NULL)),
    CONSTRAINT super_admin_auth_challenges_expiry_after_creation
        CHECK (expires_at > created_at),
    CONSTRAINT super_admin_auth_challenges_resend_after_creation
        CHECK (resend_available_at >= created_at),
    CONSTRAINT super_admin_auth_challenges_delivery_after_creation
        CHECK (delivery_started_at >= created_at),
    CONSTRAINT super_admin_auth_challenges_activation_after_delivery
        CHECK (activated_at IS NULL OR activated_at >= delivery_started_at),
    CONSTRAINT super_admin_auth_challenges_terminal_state
        CHECK (consumed_at IS NULL OR invalidated_at IS NULL),
    CONSTRAINT super_admin_auth_challenges_consumed_while_active
        CHECK (
            consumed_at IS NULL
            OR (
                active_version IS NOT NULL
                AND consumed_at >= created_at
                AND consumed_at < expires_at
            )
        ),
    CONSTRAINT super_admin_auth_challenges_invalidation_after_creation
        CHECK (invalidated_at IS NULL OR invalidated_at >= created_at)
);

-- Application code serializes challenge creation on the account row; this
-- partial unique index provides a final database guarantee that repeated or
-- concurrent password login cannot leave multiple usable challenge flows.
CREATE UNIQUE INDEX super_admin_auth_challenges_current_admin_idx
    ON super_admin_auth_challenges (super_admin_id)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

-- Supports bounded cleanup of terminal/expired authentication state.
CREATE INDEX super_admin_auth_challenges_expires_at_idx
    ON super_admin_auth_challenges (expires_at);

-- Makes abandoned pending deliveries discoverable for operational cleanup.
CREATE INDEX super_admin_auth_challenges_pending_delivery_idx
    ON super_admin_auth_challenges (delivery_started_at)
    WHERE active_version IS NULL
      AND consumed_at IS NULL
      AND invalidated_at IS NULL;

-- +goose Down

DROP TABLE super_admin_auth_challenges;
