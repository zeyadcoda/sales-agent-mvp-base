package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestPasswordHasherHashProducesCanonicalArgon2idPHC(t *testing.T) {
	hasher := NewPasswordHasher()

	encodedHash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() returned unexpected error: %v", err)
	}

	wantPrefix := "$argon2id$v=19$m=19456,t=2,p=1$"
	if !strings.HasPrefix(encodedHash, wantPrefix) {
		t.Fatalf("Hash() = %q, want prefix %q", encodedHash, wantPrefix)
	}

	parsed, err := parsePasswordHash(encodedHash)
	if err != nil {
		t.Fatalf("parse generated hash: %v", err)
	}
	if len(parsed.salt) != defaultSaltLength {
		t.Fatalf("salt length = %d, want %d", len(parsed.salt), defaultSaltLength)
	}
	if len(parsed.hash) != defaultKeyLength {
		t.Fatalf("derived key length = %d, want %d", len(parsed.hash), defaultKeyLength)
	}
}

func TestPasswordHasherUsesUniqueRandomSalts(t *testing.T) {
	hasher := NewPasswordHasher()

	first, err := hasher.Hash("same password")
	if err != nil {
		t.Fatalf("first Hash() returned unexpected error: %v", err)
	}
	second, err := hasher.Hash("same password")
	if err != nil {
		t.Fatalf("second Hash() returned unexpected error: %v", err)
	}

	if first == second {
		t.Fatal("two hashes of the same password must use independent salts")
	}
}

func TestPasswordHasherVerifiesCorrectPassword(t *testing.T) {
	hasher := NewPasswordHasher()
	encodedHash, err := hasher.Hash("a sufficiently long password")
	if err != nil {
		t.Fatalf("Hash() returned unexpected error: %v", err)
	}

	verified, err := hasher.Verify(encodedHash, "a sufficiently long password")
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if !verified {
		t.Fatal("Verify() = false, want true for the correct password")
	}
}

func TestPasswordHasherRejectsIncorrectPassword(t *testing.T) {
	hasher := NewPasswordHasher()
	encodedHash, err := hasher.Hash("the correct password")
	if err != nil {
		t.Fatalf("Hash() returned unexpected error: %v", err)
	}

	verified, err := hasher.Verify(encodedHash, "the wrong password")
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if verified {
		t.Fatal("Verify() = true, want false for an incorrect password")
	}
}

func TestPasswordHasherRejectsMalformedHashesSafely(t *testing.T) {
	salt := base64.RawStdEncoding.EncodeToString(make([]byte, minimumDecodedSaltLength))
	derivedHash := base64.RawStdEncoding.EncodeToString(make([]byte, defaultKeyLength))
	validShape := "$argon2id$v=19$m=19456,t=2,p=1$" + salt + "$" + derivedHash

	tests := []struct {
		name        string
		encodedHash string
	}{
		{name: "empty"},
		{name: "not PHC", encodedHash: "not-a-password-hash"},
		{name: "extra field", encodedHash: validShape + "$extra"},
		{name: "wrong algorithm", encodedHash: strings.Replace(validShape, "argon2id", "argon2i", 1)},
		{name: "wrong version", encodedHash: strings.Replace(validShape, "v=19", "v=18", 1)},
		{name: "missing parameter", encodedHash: strings.Replace(validShape, "m=19456,t=2,p=1", "m=19456,t=2", 1)},
		{name: "invalid memory", encodedHash: strings.Replace(validShape, "m=19456", "m=invalid", 1)},
		{name: "non-canonical memory", encodedHash: strings.Replace(validShape, "m=19456", "m=019456", 1)},
		{name: "memory too small", encodedHash: strings.Replace(validShape, "m=19456", "m=1", 1)},
		{name: "memory below policy", encodedHash: strings.Replace(validShape, "m=19456", "m=19455", 1)},
		{name: "memory too large", encodedHash: strings.Replace(validShape, "m=19456", "m=65537", 1)},
		{name: "iterations below policy", encodedHash: strings.Replace(validShape, "t=2", "t=1", 1)},
		{name: "iterations too large", encodedHash: strings.Replace(validShape, "t=2", "t=11", 1)},
		{name: "zero parallelism", encodedHash: strings.Replace(validShape, "p=1", "p=0", 1)},
		{name: "parallelism too large", encodedHash: strings.Replace(validShape, "p=1", "p=9", 1)},
		{name: "invalid salt base64", encodedHash: strings.Replace(validShape, salt, "not*base64", 1)},
		{name: "padded salt", encodedHash: strings.Replace(validShape, salt, salt+"=", 1)},
		{name: "short salt", encodedHash: strings.Replace(validShape, salt, base64.RawStdEncoding.EncodeToString([]byte("short")), 1)},
		{name: "invalid hash base64", encodedHash: strings.TrimSuffix(validShape, derivedHash) + "not*base64"},
		{name: "short derived hash", encodedHash: strings.TrimSuffix(validShape, derivedHash) + base64.RawStdEncoding.EncodeToString([]byte("short"))},
		{name: "oversized encoding", encodedHash: strings.Repeat("x", maximumEncodedHashLength+1)},
	}

	hasher := NewPasswordHasher()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verified, err := hasher.Verify(tt.encodedHash, "password")
			if verified {
				t.Fatal("Verify() = true for malformed stored hash")
			}
			if !errors.Is(err, ErrInvalidPasswordHash) {
				t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidPasswordHash)
			}
		})
	}
}

func TestPasswordHasherReportsEntropyFailure(t *testing.T) {
	wantErr := errors.New("simulated entropy failure")
	hasher := NewPasswordHasher()
	hasher.random = failingReader{err: wantErr}

	if _, err := hasher.Hash("password"); !errors.Is(err, wantErr) {
		t.Fatalf("Hash() error = %v, want error matching %v", err, wantErr)
	}
}

func TestNilPasswordHasherFailsSafely(t *testing.T) {
	var hasher *PasswordHasher

	if _, err := hasher.Hash("password"); !errors.Is(err, ErrPasswordHasherUnavailable) {
		t.Fatalf("Hash() error = %v, want %v", err, ErrPasswordHasherUnavailable)
	}
	if verified, err := hasher.Verify("stored hash", "password"); verified || !errors.Is(err, ErrPasswordHasherUnavailable) {
		t.Fatalf("Verify() = (%v, %v), want (false, %v)", verified, err, ErrPasswordHasherUnavailable)
	}
}

type failingReader struct {
	err error
}

func (r failingReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
