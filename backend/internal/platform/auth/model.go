package auth

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrSuperAdminNotFound and ErrSessionNotFound are repository boundary
	// errors. HTTP callers must never receive either error directly because
	// doing so would disclose account or security-state details.
	ErrSuperAdminNotFound = errors.New("super admin not found")
	ErrSessionNotFound    = errors.New("super admin session not found")
	ErrSuperAdminExists   = errors.New("super admin already exists")

	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrOTPRequired               = errors.New("email OTP required")
	ErrUnauthenticated           = errors.New("valid authentication session required")
	ErrInvalidCSRFToken          = errors.New("invalid CSRF token")
	ErrAuthenticationUnavailable = errors.New("authentication unavailable")
)

// SuperAdmin contains the account fields needed by authentication. The
// password hash is intentionally unexported from all HTTP response DTOs.
type SuperAdmin struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session is authoritative server-side authentication state loaded from
// PostgreSQL. A raw session token is never represented in this model.
type Session struct {
	ID         string
	SuperAdmin SuperAdmin
	CSRFToken  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	LastSeenAt time.Time
}

// NewSession contains the server-derived fields needed to persist a session.
// TokenHash is SHA-256 output; only the raw token sent in the HttpOnly cookie
// can reproduce it.
type NewSession struct {
	SuperAdminID string
	TokenHash    []byte
	CSRFToken    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastSeenAt   time.Time
}

// NewSuperAdmin is used only by the deployment/local bootstrap command. It is
// deliberately separate from browser request DTOs so privileged account
// fields can never become mass assignable.
type NewSuperAdmin struct {
	Email        string
	PasswordHash string
	DisplayName  string
	IsActive     bool
}

// Store is the narrow persistence contract owned by the authentication
// domain. PostgreSQL implements it; Redis must never implement or replace it.
type Store interface {
	FindSuperAdminByEmail(ctx context.Context, normalizedEmail string) (SuperAdmin, error)
	CreateSession(ctx context.Context, session NewSession) error
	FindSessionByTokenHash(ctx context.Context, tokenHash []byte) (Session, error)
	TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
}

// ProvisioningStore is intentionally separate from Store because normal API
// authentication has no authority to create privileged accounts.
type ProvisioningStore interface {
	CreateSuperAdmin(ctx context.Context, account NewSuperAdmin) (SuperAdmin, error)
}

// AuthenticatedSession is the safe subset returned to the same-origin web
// application. It contains CSRF material but never the raw session token or
// password hash.
type AuthenticatedSession struct {
	SuperAdmin SuperAdmin
	CSRFToken  string
	ExpiresAt  time.Time
}
