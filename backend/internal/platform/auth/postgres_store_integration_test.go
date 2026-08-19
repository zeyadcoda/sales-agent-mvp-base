package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"salesagent.local/backend/internal/database"
)

func TestPostgresStoreAuthenticationRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping auth repository integration test")
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

	hasher := NewPasswordHasher()
	passwordHash, err := hasher.Hash("integration-only-password")
	if err != nil {
		t.Fatalf("hash integration password: %v", err)
	}

	email := fmt.Sprintf("auth-integration-%d@example.com", time.Now().UnixNano())
	created, err := store.CreateSuperAdmin(ctx, NewSuperAdmin{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  "Integration Admin",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("create Super Admin: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_sessions WHERE super_admin_id = $1::uuid`, created.ID)
		_, _ = db.Exec(cleanupCtx, `DELETE FROM super_admin_accounts WHERE id = $1::uuid`, created.ID)
	})

	loaded, err := store.FindSuperAdminByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find Super Admin: %v", err)
	}
	if loaded.PasswordHash == "integration-only-password" {
		t.Fatal("PostgreSQL contains a plaintext password")
	}
	passwordMatches, err := hasher.Verify(loaded.PasswordHash, "integration-only-password")
	if err != nil || !passwordMatches {
		t.Fatalf("verify persisted password: match=%v error=%v", passwordMatches, err)
	}

	if _, err := store.CreateSuperAdmin(ctx, NewSuperAdmin{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  "Duplicate",
		IsActive:     true,
	}); !errors.Is(err, ErrSuperAdminExists) {
		t.Fatalf("duplicate create error = %v, want %v", err, ErrSuperAdminExists)
	}

	// The injected text is a query parameter, not executable SQL. A failed
	// lookup also proves it cannot turn the predicate into a match-all clause.
	if _, err := store.FindSuperAdminByEmail(ctx, `' OR '1'='1`); !errors.Is(err, ErrSuperAdminNotFound) {
		t.Fatalf("SQL injection lookup error = %v, want not found", err)
	}

	rawToken := "integration-raw-session-token-never-store"
	tokenHash := sha256.Sum256([]byte(rawToken))
	now := time.Now().UTC().Truncate(time.Microsecond)
	if err := store.CreateSession(ctx, NewSession{
		SuperAdminID: created.ID,
		TokenHash:    tokenHash[:],
		CSRFToken:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CreatedAt:    now,
		ExpiresAt:    now.Add(8 * time.Hour),
		LastSeenAt:   now,
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	var persistedHash []byte
	if err := db.QueryRow(
		ctx,
		`SELECT token_hash FROM super_admin_sessions WHERE super_admin_id = $1::uuid`,
		created.ID,
	).Scan(&persistedHash); err != nil {
		t.Fatalf("read persisted token hash: %v", err)
	}
	if !bytes.Equal(persistedHash, tokenHash[:]) || bytes.Equal(persistedHash, []byte(rawToken)) {
		t.Fatalf("persisted token material = %x", persistedHash)
	}

	session, err := store.FindSessionByTokenHash(ctx, tokenHash[:])
	if err != nil {
		t.Fatalf("find session: %v", err)
	}
	if session.SuperAdmin.Email != email || !session.SuperAdmin.IsActive {
		t.Fatalf("session identity = %#v", session.SuperAdmin)
	}
	if err := store.RevokeSession(ctx, session.ID, now.Add(time.Minute)); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	revoked, err := store.FindSessionByTokenHash(ctx, tokenHash[:])
	if err != nil || revoked.RevokedAt == nil {
		t.Fatalf("revoked session = %#v, error = %v", revoked, err)
	}
}
