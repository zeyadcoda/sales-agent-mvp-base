-- +goose Up

-- This first migration intentionally creates only a tiny schema marker.
-- We are proving that the migration lifecycle works before introducing
-- business tables such as Super Admin, Organization, Package, or Agent data.
CREATE TABLE schema_marker (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

-- Rollback removes the marker table so we can verify that migrations are
-- reversible in local development before adding real application schema.
DROP TABLE schema_marker;