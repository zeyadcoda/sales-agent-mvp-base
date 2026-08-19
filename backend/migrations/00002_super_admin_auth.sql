-- +goose Up

CREATE TABLE super_admin_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT super_admin_accounts_email_unique UNIQUE (email),
    CONSTRAINT super_admin_accounts_email_normalized
        CHECK (email = LOWER(BTRIM(email))),
    CONSTRAINT super_admin_accounts_email_length
        CHECK (CHAR_LENGTH(email) BETWEEN 3 AND 254),
    CONSTRAINT super_admin_accounts_password_hash_length
        CHECK (CHAR_LENGTH(password_hash) BETWEEN 1 AND 512),
    CONSTRAINT super_admin_accounts_display_name_normalized
        CHECK (
            display_name = BTRIM(display_name)
            AND CHAR_LENGTH(display_name) BETWEEN 1 AND 200
        ),
    CONSTRAINT super_admin_accounts_timestamps_ordered
        CHECK (updated_at >= created_at)
);

CREATE TABLE super_admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    super_admin_id UUID NOT NULL,
    -- The browser receives the opaque session secret; PostgreSQL retains only
    -- its fixed-length SHA-256 digest so a database read cannot replay it.
    token_hash BYTEA NOT NULL,
    -- This independent synchronizer token is returned only by the same-origin
    -- session endpoint and must never be logged or used as authentication.
    csrf_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT super_admin_sessions_admin_fk
        FOREIGN KEY (super_admin_id)
        REFERENCES super_admin_accounts (id)
        ON DELETE RESTRICT,
    CONSTRAINT super_admin_sessions_token_hash_unique UNIQUE (token_hash),
    CONSTRAINT super_admin_sessions_token_hash_sha256
        CHECK (OCTET_LENGTH(token_hash) = 32),
    CONSTRAINT super_admin_sessions_csrf_token_unique UNIQUE (csrf_token),
    CONSTRAINT super_admin_sessions_csrf_token_length
        CHECK (CHAR_LENGTH(csrf_token) BETWEEN 43 AND 128),
    CONSTRAINT super_admin_sessions_expiry_after_creation
        CHECK (expires_at > created_at),
    CONSTRAINT super_admin_sessions_revocation_after_creation
        CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CONSTRAINT super_admin_sessions_last_seen_in_lifetime
        CHECK (last_seen_at >= created_at AND last_seen_at <= expires_at)
);

-- Supports listing/revoking currently active sessions for one Super Admin.
CREATE INDEX super_admin_sessions_active_admin_idx
    ON super_admin_sessions (super_admin_id, expires_at)
    WHERE revoked_at IS NULL;

-- Supports bounded cleanup of expired session rows without scanning the table.
CREATE INDEX super_admin_sessions_expires_at_idx
    ON super_admin_sessions (expires_at);

-- +goose Down

DROP TABLE super_admin_sessions;
DROP TABLE super_admin_accounts;
