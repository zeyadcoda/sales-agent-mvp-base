package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

type fakeProvisioningStore struct {
	created            auth.NewSuperAdmin
	createHadDeadline  bool
	createTimeToExpiry time.Duration
	err                error
}

func (store *fakeProvisioningStore) CreateSuperAdmin(
	ctx context.Context,
	account auth.NewSuperAdmin,
) (auth.SuperAdmin, error) {
	store.created = account
	deadline, hasDeadline := ctx.Deadline()
	store.createHadDeadline = hasDeadline
	if hasDeadline {
		store.createTimeToExpiry = time.Until(deadline)
	}
	if store.err != nil {
		return auth.SuperAdmin{}, store.err
	}

	return auth.SuperAdmin{
		Email:       account.Email,
		DisplayName: account.DisplayName,
		IsActive:    account.IsActive,
	}, nil
}

type fakePasswordHasher struct {
	hash     string
	password string
	err      error
}

func (hasher *fakePasswordHasher) Hash(password string) (string, error) {
	hasher.password = password
	return hasher.hash, hasher.err
}

func TestProvisionCreatesActiveSuperAdminWithHash(t *testing.T) {
	t.Parallel()

	store := &fakeProvisioningStore{}
	hasher := &fakePasswordHasher{hash: "$argon2id$test-hash"}

	created, err := provision(
		context.Background(),
		store,
		hasher,
		"admin@example.com",
		"Super Admin",
		"a secure local password",
		"a secure local password",
	)
	if err != nil {
		t.Fatalf("provision() error = %v", err)
	}
	if created.Email != "admin@example.com" || !created.IsActive {
		t.Fatalf("created account = %#v", created)
	}
	if store.created.PasswordHash != hasher.hash {
		t.Fatalf("stored hash = %q, want %q", store.created.PasswordHash, hasher.hash)
	}
	if store.created.PasswordHash == hasher.password {
		t.Fatal("plaintext password must never be sent to persistence")
	}
	if !store.createHadDeadline {
		t.Fatal("CreateSuperAdmin context must have a bounded deadline")
	}
	if store.createTimeToExpiry < databaseOperationTimeout-time.Second {
		t.Fatalf(
			"CreateSuperAdmin deadline had %s remaining, want a fresh %s timeout",
			store.createTimeToExpiry,
			databaseOperationTimeout,
		)
	}
}

func TestProvisionRejectsWeakOrMismatchedPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		password     string
		confirmation string
	}{
		{name: "weak", password: "short", confirmation: "short"},
		{name: "mismatch", password: "a secure local password", confirmation: "different secure pass"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeProvisioningStore{}
			hasher := &fakePasswordHasher{hash: "unused"}

			_, err := provision(
				context.Background(),
				store,
				hasher,
				"admin@example.com",
				"Super Admin",
				test.password,
				test.confirmation,
			)
			if err == nil {
				t.Fatal("provision() should reject the password")
			}
			if store.created.Email != "" || hasher.password != "" {
				t.Fatal("rejected password must not be hashed or persisted")
			}
		})
	}
}

func TestProvisionDoesNotOverwriteExistingSuperAdmin(t *testing.T) {
	t.Parallel()

	store := &fakeProvisioningStore{err: auth.ErrSuperAdminExists}
	hasher := &fakePasswordHasher{hash: "$argon2id$test-hash"}

	_, err := provision(
		context.Background(),
		store,
		hasher,
		"admin@example.com",
		"Super Admin",
		"a secure local password",
		"a secure local password",
	)
	if !errors.Is(err, auth.ErrSuperAdminExists) {
		t.Fatalf("provision() error = %v, want %v", err, auth.ErrSuperAdminExists)
	}
}

func TestCommandDoesNotDefinePasswordFlag(t *testing.T) {
	t.Parallel()

	// Parsing fails before database access, proving the CLI cannot accept a
	// password value that would be exposed in shell history/process listings.
	err := run(
		context.Background(),
		[]string{"--email", "admin@example.com", "--password", "do-not-accept"},
		os.Stdin,
		&strings.Builder{},
	)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("run() error = %v, want unknown password flag error", err)
	}
}
