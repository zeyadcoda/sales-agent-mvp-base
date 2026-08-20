package auth

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"salesagent.local/backend/internal/requestmeta"
)

const (
	// DefaultRecoveryAuthorizationValidity is deliberately short and cannot be
	// extended through RecoveryServiceOptions. A deployment operator must create
	// a new, reasoned authorization after this one expires.
	DefaultRecoveryAuthorizationValidity = 10 * time.Minute

	maximumRecoveryReasonRunes   = 500
	maximumRecoveryOperatorRunes = 200
	maximumRecoveryCorrelationID = 128
)

var (
	// ErrRecoveryTargetNotEligible deliberately covers malformed, unknown, and
	// inactive Super Admin targets so callers do not receive account-state detail.
	ErrRecoveryTargetNotEligible = errors.New("super admin recovery target is not eligible")
	ErrRecoveryAlreadyActive     = errors.New("an active super admin recovery authorization already exists")
	ErrRecoveryUnavailable       = errors.New("super admin recovery is unavailable")
	ErrRecoveryNotActive         = errors.New("no active super admin recovery authorization exists")

	ErrInvalidRecoveryReason        = errors.New("recovery reason is required and must be safe")
	ErrInvalidRecoveryOperator      = errors.New("recovery operator identifier is required and must be safe")
	ErrInvalidRecoveryCorrelationID = errors.New("recovery correlation ID must be safe")
	ErrInvalidRecoveryValidity      = errors.New("recovery authorization validity must be positive and at most 10 minutes")
)

// RecoveryAuthorization is deployment-created, PostgreSQL-authoritative state.
// Its identifier is an internal database/audit reference, never a browser
// credential or a code that a user enters. No reusable recovery secret exists.
type RecoveryAuthorization struct {
	ID                 string
	SuperAdminID       string
	SuperAdminEmail    string
	Reason             string
	OperatorIdentifier string
	CorrelationID      string
	CreatedAt          time.Time
	ExpiresAt          time.Time
	ConsumedAt         *time.Time
	RevokedAt          *time.Time
	ExpiredAt          *time.Time
}

type RecoveryAuthorizationState string

const (
	RecoveryAuthorizationStateNone     RecoveryAuthorizationState = "NONE"
	RecoveryAuthorizationStateActive   RecoveryAuthorizationState = "ACTIVE"
	RecoveryAuthorizationStateExpired  RecoveryAuthorizationState = "EXPIRED"
	RecoveryAuthorizationStateConsumed RecoveryAuthorizationState = "CONSUMED"
	RecoveryAuthorizationStateRevoked  RecoveryAuthorizationState = "REVOKED"
)

// RecoveryAuthorizationStatus is safe deployment-CLI state. It contains no
// password, OTP, session token, provider credential, or recovery secret.
// Authorization is the zero value when State is NONE.
type RecoveryAuthorizationStatus struct {
	State         RecoveryAuthorizationState
	Authorization RecoveryAuthorization
}

// StateAt classifies already-loaded authorization state using the trusted
// operation time supplied by its caller. PostgreSQL remains authoritative for
// loading and locking the record; this helper keeps CLI/status classification
// consistent and never grants authentication capability.
func (authorization RecoveryAuthorization) StateAt(now time.Time) RecoveryAuthorizationState {
	if authorization.ID == "" {
		return RecoveryAuthorizationStateNone
	}

	switch {
	case authorization.ConsumedAt != nil:
		return RecoveryAuthorizationStateConsumed
	case authorization.RevokedAt != nil:
		return RecoveryAuthorizationStateRevoked
	case authorization.ExpiredAt != nil || !authorization.ExpiresAt.After(now.UTC()):
		return RecoveryAuthorizationStateExpired
	default:
		return RecoveryAuthorizationStateActive
	}
}

// RecoveryAdministrationStore is implemented by the deployment-authorized
// PostgreSQL repository. Each mutation must perform target lookup, account-row
// locking, authorization state change, and immutable audit creation in one
// transaction. It is intentionally not an HTTP or Agent-facing capability.
type RecoveryAdministrationStore interface {
	AuthorizeSuperAdminRecovery(
		ctx context.Context,
		normalizedEmail string,
		reason string,
		operatorIdentifier string,
		correlationID string,
		createdAt time.Time,
		expiresAt time.Time,
	) (RecoveryAuthorization, error)
	FindSuperAdminRecoveryStatus(
		ctx context.Context,
		normalizedEmail string,
		now time.Time,
	) (RecoveryAuthorizationStatus, error)
	RevokeSuperAdminRecovery(
		ctx context.Context,
		normalizedEmail string,
		reason string,
		operatorIdentifier string,
		correlationID string,
		revokedAt time.Time,
	) (RecoveryAuthorization, error)
}

type RecoveryServiceOptions struct {
	// Validity defaults to DefaultRecoveryAuthorizationValidity. A positive,
	// shorter duration is accepted for safer deployments and deterministic tests;
	// a longer authorization is rejected.
	Validity time.Duration
	Now      func() time.Time
}

// RecoveryService exposes deployment administration use cases without granting
// browser handlers or Agents direct database authority.
type RecoveryService struct {
	store    RecoveryAdministrationStore
	validity time.Duration
	now      func() time.Time
}

func NewRecoveryService(
	store RecoveryAdministrationStore,
	options RecoveryServiceOptions,
) (*RecoveryService, error) {
	if store == nil {
		return nil, errors.New("recovery administration store is required")
	}

	validity := options.Validity
	if validity == 0 {
		validity = DefaultRecoveryAuthorizationValidity
	}
	if validity <= 0 || validity > DefaultRecoveryAuthorizationValidity {
		return nil, ErrInvalidRecoveryValidity
	}
	if options.Now == nil {
		options.Now = time.Now
	}

	return &RecoveryService{
		store:    store,
		validity: validity,
		now:      options.Now,
	}, nil
}

// Authorize creates one short-lived, one-time authorization for the exact
// normalized Super Admin email. PostgreSQL owns target eligibility, duplicate
// exclusion, authorization creation, and its audit event atomically.
func (service *RecoveryService) Authorize(
	ctx context.Context,
	email string,
	reason string,
	operatorIdentifier string,
) (RecoveryAuthorization, error) {
	if service == nil || service.store == nil || service.now == nil {
		return RecoveryAuthorization{}, ErrRecoveryUnavailable
	}

	normalizedEmail, err := normalizeRecoveryTargetEmail(email)
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	reason, err = NormalizeRecoveryReason(reason)
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	operatorIdentifier, err = NormalizeRecoveryOperator(operatorIdentifier)
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	ctx, correlationID, err := recoveryOperationContext(ctx)
	if err != nil {
		return RecoveryAuthorization{}, err
	}

	createdAt := service.now().UTC()
	expiresAt := createdAt.Add(service.validity)
	if !expiresAt.After(createdAt) {
		return RecoveryAuthorization{}, ErrRecoveryUnavailable
	}

	authorization, err := service.store.AuthorizeSuperAdminRecovery(
		ctx,
		normalizedEmail,
		reason,
		operatorIdentifier,
		correlationID,
		createdAt,
		expiresAt,
	)
	if err != nil {
		return RecoveryAuthorization{}, publicRecoveryAuthorizeError(err)
	}

	return authorization, nil
}

// Status returns authoritative lifecycle state for an eligible target. NONE is
// a successful result for a known active Super Admin with no recovery history.
func (service *RecoveryService) Status(
	ctx context.Context,
	email string,
) (RecoveryAuthorizationStatus, error) {
	if service == nil || service.store == nil || service.now == nil {
		return RecoveryAuthorizationStatus{}, ErrRecoveryUnavailable
	}

	normalizedEmail, err := normalizeRecoveryTargetEmail(email)
	if err != nil {
		return RecoveryAuthorizationStatus{}, err
	}
	ctx, _, err = recoveryOperationContext(ctx)
	if err != nil {
		return RecoveryAuthorizationStatus{}, err
	}

	status, err := service.store.FindSuperAdminRecoveryStatus(
		ctx,
		normalizedEmail,
		service.now().UTC(),
	)
	if err != nil {
		if errors.Is(err, ErrRecoveryTargetNotEligible) {
			return RecoveryAuthorizationStatus{}, ErrRecoveryTargetNotEligible
		}
		return RecoveryAuthorizationStatus{}, ErrRecoveryUnavailable
	}
	if !validRecoveryAuthorizationState(status.State) {
		return RecoveryAuthorizationStatus{}, ErrRecoveryUnavailable
	}

	return status, nil
}

// Revoke terminates the current authorization through the deployment-only
// PostgreSQL operation and requires a fresh reason and operator attribution.
func (service *RecoveryService) Revoke(
	ctx context.Context,
	email string,
	reason string,
	operatorIdentifier string,
) (RecoveryAuthorization, error) {
	if service == nil || service.store == nil || service.now == nil {
		return RecoveryAuthorization{}, ErrRecoveryUnavailable
	}

	normalizedEmail, err := normalizeRecoveryTargetEmail(email)
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	reason, err = NormalizeRecoveryReason(reason)
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	operatorIdentifier, err = NormalizeRecoveryOperator(operatorIdentifier)
	if err != nil {
		return RecoveryAuthorization{}, err
	}
	ctx, correlationID, err := recoveryOperationContext(ctx)
	if err != nil {
		return RecoveryAuthorization{}, err
	}

	authorization, err := service.store.RevokeSuperAdminRecovery(
		ctx,
		normalizedEmail,
		reason,
		operatorIdentifier,
		correlationID,
		service.now().UTC(),
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecoveryTargetNotEligible):
			return RecoveryAuthorization{}, ErrRecoveryTargetNotEligible
		case errors.Is(err, ErrRecoveryNotActive):
			return RecoveryAuthorization{}, ErrRecoveryNotActive
		default:
			return RecoveryAuthorization{}, ErrRecoveryUnavailable
		}
	}

	return authorization, nil
}

func NormalizeRecoveryReason(value string) (string, error) {
	return normalizeRecoveryText(value, maximumRecoveryReasonRunes, ErrInvalidRecoveryReason)
}

func NormalizeRecoveryOperator(value string) (string, error) {
	return normalizeRecoveryText(value, maximumRecoveryOperatorRunes, ErrInvalidRecoveryOperator)
}

func normalizeRecoveryTargetEmail(value string) (string, error) {
	normalized, err := NormalizeEmail(value)
	if err != nil {
		return "", ErrRecoveryTargetNotEligible
	}
	return normalized, nil
}

func normalizeRecoveryText(value string, maximumRunes int, invalidError error) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maximumRunes {
		return "", invalidError
	}
	for _, character := range value {
		if unicode.IsControl(character) || unicode.Is(unicode.Zl, character) || unicode.Is(unicode.Zp, character) {
			return "", invalidError
		}
	}

	return value, nil
}

func recoveryOperationContext(ctx context.Context) (context.Context, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	correlationID := requestmeta.CorrelationID(ctx)
	if correlationID == "" ||
		!utf8.ValidString(correlationID) ||
		len(correlationID) > maximumRecoveryCorrelationID ||
		containsUnsafeSingleLineCharacter(correlationID) {
		return ctx, "", ErrInvalidRecoveryCorrelationID
	}

	return ctx, correlationID, nil
}

func containsUnsafeSingleLineCharacter(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) || unicode.Is(unicode.Zl, character) || unicode.Is(unicode.Zp, character) {
			return true
		}
	}
	return false
}

func publicRecoveryAuthorizeError(err error) error {
	switch {
	case errors.Is(err, ErrRecoveryTargetNotEligible):
		return ErrRecoveryTargetNotEligible
	case errors.Is(err, ErrRecoveryAlreadyActive):
		return ErrRecoveryAlreadyActive
	default:
		return ErrRecoveryUnavailable
	}
}

func validRecoveryAuthorizationState(state RecoveryAuthorizationState) bool {
	switch state {
	case RecoveryAuthorizationStateNone,
		RecoveryAuthorizationStateActive,
		RecoveryAuthorizationStateExpired,
		RecoveryAuthorizationStateConsumed,
		RecoveryAuthorizationStateRevoked:
		return true
	default:
		return false
	}
}
