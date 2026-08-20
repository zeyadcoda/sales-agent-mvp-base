// Package audit owns the platform-wide immutable AuditEvent contract.
// Trusted application services append events; browsers and Agents never receive
// a generic capability to create, update, or delete this history.
package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type ActorType string
type Action string
type ResourceType string
type Result string

const (
	ActorTypeDeploymentOperator ActorType = "DEPLOYMENT_OPERATOR"
	ActorTypeSuperAdmin         ActorType = "SUPER_ADMIN"
	ActorTypeSystem             ActorType = "SYSTEM"

	ActionSuperAdminRecoveryAuthorized          Action = "SUPER_ADMIN_RECOVERY_AUTHORIZED"
	ActionSuperAdminRecoveryAuthorizationFailed Action = "SUPER_ADMIN_RECOVERY_AUTHORIZATION_FAILED"
	ActionSuperAdminRecoveryConsumed            Action = "SUPER_ADMIN_RECOVERY_CONSUMED"
	ActionSuperAdminRecoveryRevoked             Action = "SUPER_ADMIN_RECOVERY_REVOKED"
	ActionSuperAdminRecoveryRevocationFailed    Action = "SUPER_ADMIN_RECOVERY_REVOCATION_FAILED"

	ResourceTypeSuperAdminAccount ResourceType = "SUPER_ADMIN_ACCOUNT"

	ResultSuccess Result = "SUCCESS"
	ResultFailure Result = "FAILURE"
)

const (
	maxActorIdentifierLength    = 254
	maxResourceReferenceLength  = 512
	maxReasonLength             = 1000
	maxCorrelationIDLength      = 128
	maxPlatformAuditListResults = 100
)

var (
	ErrInvalidEvent       = errors.New("invalid platform audit event")
	ErrAuditStoreRequired = errors.New("platform audit store is required")

	auditCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

// Event is the reusable platform AuditEvent model. ID is output-only for this
// append API: PostgreSQL assigns it, and ListByResource populates it. Nullable
// values remain explicit so callers cannot confuse absent audit context with an
// empty value.
type Event struct {
	ID                string          `json:"id"`
	OccurredAt        time.Time       `json:"occurred_at"`
	ActorType         ActorType       `json:"actor_type"`
	ActorIdentifier   string          `json:"actor_identifier"`
	Action            Action          `json:"action"`
	ResourceType      ResourceType    `json:"resource_type"`
	ResourceReference string          `json:"resource_reference"`
	OrganizationID    *string         `json:"organization_id"`
	OldValues         json.RawMessage `json:"old_values"`
	NewValues         json.RawMessage `json:"new_values"`
	Reason            *string         `json:"reason"`
	Result            Result          `json:"result"`
	CorrelationID     string          `json:"correlation_id"`
}

func validateEvent(event Event) error {
	if event.ID != "" {
		return invalidEvent("ID is assigned by PostgreSQL")
	}
	if event.OccurredAt.IsZero() {
		return invalidEvent("occurred_at is required")
	}
	if err := validateCode("actor_type", string(event.ActorType), 64); err != nil {
		return err
	}
	if err := validateText("actor_identifier", event.ActorIdentifier, maxActorIdentifierLength); err != nil {
		return err
	}
	if err := validateCode("action", string(event.Action), 128); err != nil {
		return err
	}
	if err := validateCode("resource_type", string(event.ResourceType), 64); err != nil {
		return err
	}
	if err := validateText("resource_reference", event.ResourceReference, maxResourceReferenceLength); err != nil {
		return err
	}
	if event.OrganizationID != nil {
		if err := validateText("organization_id", *event.OrganizationID, 36); err != nil {
			return err
		}
	}
	if err := validateSafeJSONObject("old_values", event.OldValues); err != nil {
		return err
	}
	if err := validateSafeJSONObject("new_values", event.NewValues); err != nil {
		return err
	}
	if event.Reason != nil {
		if err := validateText("reason", *event.Reason, maxReasonLength); err != nil {
			return err
		}
	}
	if strings.HasPrefix(string(event.Action), "SUPER_ADMIN_RECOVERY_") && event.Reason == nil {
		return invalidEvent("reason is required for recovery events")
	}
	if event.Result != ResultSuccess && event.Result != ResultFailure {
		return invalidEvent("result is not recognized")
	}
	if err := validateText("correlation_id", event.CorrelationID, maxCorrelationIDLength); err != nil {
		return err
	}

	return nil
}

func validateLoadedEvent(event Event) error {
	if event.ID == "" {
		return invalidEvent("persisted ID is required")
	}
	event.ID = ""
	return validateEvent(event)
}

func validateListInput(resourceType ResourceType, resourceReference string, limit int) error {
	if err := validateCode("resource_type", string(resourceType), 64); err != nil {
		return err
	}
	if err := validateText("resource_reference", resourceReference, maxResourceReferenceLength); err != nil {
		return err
	}
	if limit < 1 || limit > maxPlatformAuditListResults {
		return invalidEvent("list limit must be between 1 and 100")
	}

	return nil
}

func validateCode(field string, value string, maximum int) error {
	if utf8.RuneCountInString(value) < 1 ||
		utf8.RuneCountInString(value) > maximum ||
		!auditCodePattern.MatchString(value) {
		return invalidEvent(field + " has an invalid format")
	}

	return nil
}

func validateText(field string, value string, maximum int) error {
	length := utf8.RuneCountInString(value)
	if value != strings.TrimSpace(value) || length < 1 || length > maximum {
		return invalidEvent(field + " has an invalid format")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return invalidEvent(field + " contains a control character")
		}
	}

	return nil
}

func validateSafeJSONObject(field string, value json.RawMessage) error {
	if len(value) == 0 {
		return nil
	}

	var object map[string]any
	if err := json.Unmarshal(value, &object); err != nil || object == nil {
		return invalidEvent(field + " must be a JSON object")
	}
	if containsForbiddenAuditKey(object) {
		return invalidEvent(field + " contains a prohibited credential field")
	}

	return nil
}

func containsForbiddenAuditKey(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if forbiddenAuditKey(key) || containsForbiddenAuditKey(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsForbiddenAuditKey(nested) {
				return true
			}
		}
	}

	return false
}

func forbiddenAuditKey(key string) bool {
	var normalized strings.Builder
	for _, character := range strings.ToLower(key) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			normalized.WriteRune(character)
		}
	}

	switch normalized.String() {
	case "password",
		"passwordhash",
		"otp",
		"otphash",
		"sessiontoken",
		"rawsessiontoken",
		"recoverytoken",
		"apikey",
		"smtppassword",
		"secret",
		"csrftoken",
		"hmacsecret":
		return true
	default:
		return false
	}
}

func invalidEvent(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidEvent, message)
}
