package auth

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidEmail         = errors.New("invalid email address")
	ErrInvalidPasswordInput = errors.New("invalid password input")
	ErrWeakPassword         = errors.New("password must contain at least 12 characters")
	ErrInvalidDisplayName   = errors.New("invalid display name")
)

const (
	maxEmailBytes       = 254
	maxLoginPassword    = 1024
	minBootstrapRunes   = 12
	maxBootstrapRunes   = 128
	maxDisplayNameRunes = 100
)

// NormalizeEmail defines the one canonical account key used by login,
// throttling, bootstrap, and PostgreSQL uniqueness enforcement.
func NormalizeEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxEmailBytes || strings.ContainsAny(value, "\r\n") {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", ErrInvalidEmail
	}

	local, domain, ok := strings.Cut(parsed.Address, "@")
	if !ok || local == "" || domain == "" || strings.Contains(domain, "@") {
		return "", ErrInvalidEmail
	}

	return strings.ToLower(parsed.Address), nil
}

// ValidateLoginPassword bounds work and memory retained from an untrusted
// request without applying bootstrap policy to an existing stored account.
func ValidateLoginPassword(password string) error {
	if password == "" || len(password) > maxLoginPassword {
		return ErrInvalidPasswordInput
	}

	return nil
}

// ValidateBootstrapPassword applies the privileged-account provisioning
// policy before expensive password hashing begins.
func ValidateBootstrapPassword(password string) error {
	runeCount := utf8.RuneCountInString(password)
	if runeCount < minBootstrapRunes || runeCount > maxBootstrapRunes || strings.TrimSpace(password) == "" {
		return ErrWeakPassword
	}

	return nil
}

// ValidateDisplayName keeps the operator-facing identity useful while
// preventing control characters and unbounded database/UI values.
func ValidateDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > maxDisplayNameRunes || strings.ContainsAny(value, "\r\n\x00") {
		return "", ErrInvalidDisplayName
	}

	return value, nil
}
