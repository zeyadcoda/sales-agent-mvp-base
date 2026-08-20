// Package email implements authentication-email delivery without exposing a
// vendor SDK to the authentication domain.
package email

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"salesagent.local/backend/internal/config"
	"salesagent.local/backend/internal/platform/auth"
)

var (
	// ErrInvalidOTPEmail indicates an invalid application-to-adapter message.
	// It never contains the OTP or recipient value.
	ErrInvalidOTPEmail = errors.New("invalid OTP email")
	// ErrDeliveryUnavailable deliberately hides SMTP and provider details from
	// the authentication boundary.
	ErrDeliveryUnavailable = errors.New("OTP email delivery unavailable")
)

// MailpitSender is the local/test adapter. Its constructor accepts only the
// loopback Mailpit endpoint so it cannot silently become a remote plaintext
// SMTP transport.
type MailpitSender struct {
	delivery smtpDelivery
}

// SMTPSender is the provider-ready adapter. STARTTLS or direct TLS is required
// before credentials or OTP content can leave the process.
type SMTPSender struct {
	delivery smtpDelivery
}

var (
	_ auth.OTPEmailSender = (*MailpitSender)(nil)
	_ auth.OTPEmailSender = (*SMTPSender)(nil)
)

// NewSender selects only between the two explicitly configured adapters.
func NewSender(settings config.AuthEmailConfig) (auth.OTPEmailSender, error) {
	switch settings.Mode {
	case config.AuthEmailMailpit:
		return NewMailpitSender(settings)
	case config.AuthEmailSMTP:
		return NewSMTPSender(settings)
	default:
		return nil, errors.New("authentication email mode is invalid")
	}
}

func NewMailpitSender(settings config.AuthEmailConfig) (*MailpitSender, error) {
	if settings.Mode != config.AuthEmailMailpit ||
		settings.SMTPHost != "127.0.0.1" ||
		settings.SMTPPort != 1025 ||
		settings.SMTPTLSMode != config.SMTPTLSNone ||
		settings.SMTPUsername != "" ||
		settings.SMTPPassword != "" {
		return nil, errors.New("Mailpit must use unauthenticated SMTP on 127.0.0.1:1025")
	}

	delivery, err := newSMTPDelivery(settings)
	if err != nil {
		return nil, err
	}

	return &MailpitSender{delivery: delivery}, nil
}

func NewSMTPSender(settings config.AuthEmailConfig) (*SMTPSender, error) {
	if settings.Mode != config.AuthEmailSMTP ||
		(settings.SMTPTLSMode != config.SMTPTLSSTARTTLS && settings.SMTPTLSMode != config.SMTPTLSDirect) {
		return nil, errors.New("provider SMTP must use STARTTLS or direct TLS")
	}

	delivery, err := newSMTPDelivery(settings)
	if err != nil {
		return nil, err
	}

	return &SMTPSender{delivery: delivery}, nil
}

func newSMTPDelivery(settings config.AuthEmailConfig) (smtpDelivery, error) {
	fromAddress, err := validateBareMailbox(settings.FromAddress)
	if err != nil {
		return smtpDelivery{}, errors.New("authentication email sender address is invalid")
	}
	fromName, err := validateDisplayName(settings.FromName)
	if err != nil {
		return smtpDelivery{}, errors.New("authentication email sender name is invalid")
	}
	if settings.SMTPHost == "" || strings.ContainsAny(settings.SMTPHost, "\r\n\x00") {
		return smtpDelivery{}, errors.New("authentication SMTP host is invalid")
	}
	if settings.SMTPPort < 1 || settings.SMTPPort > 65535 {
		return smtpDelivery{}, errors.New("authentication SMTP port is invalid")
	}
	if settings.SMTPTimeout <= 0 {
		return smtpDelivery{}, errors.New("authentication SMTP timeout is invalid")
	}
	if (settings.SMTPUsername == "") != (settings.SMTPPassword == "") {
		return smtpDelivery{}, errors.New("authentication SMTP credentials are incomplete")
	}
	if strings.ContainsAny(settings.SMTPUsername, "\r\n\x00") || strings.ContainsAny(settings.SMTPPassword, "\r\n\x00") {
		return smtpDelivery{}, errors.New("authentication SMTP credentials are invalid")
	}
	if settings.SMTPUsername != "" && settings.SMTPTLSMode == config.SMTPTLSNone {
		return smtpDelivery{}, errors.New("authentication SMTP credentials require TLS")
	}

	return smtpDelivery{
		fromAddress: fromAddress,
		fromName:    fromName,
		host:        settings.SMTPHost,
		port:        settings.SMTPPort,
		tlsMode:     settings.SMTPTLSMode,
		username:    settings.SMTPUsername,
		password:    settings.SMTPPassword,
		timeout:     settings.SMTPTimeout,
	}, nil
}

func (s *MailpitSender) SendOTP(ctx context.Context, message auth.OTPEmail) error {
	return s.delivery.sendOTP(ctx, message)
}

func (s *SMTPSender) SendOTP(ctx context.Context, message auth.OTPEmail) error {
	return s.delivery.sendOTP(ctx, message)
}

func validateDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 100 {
		return "", ErrInvalidOTPEmail
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", ErrInvalidOTPEmail
		}
	}

	return value, nil
}

func validateOTP(value string) error {
	if len(value) != 6 {
		return ErrInvalidOTPEmail
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return ErrInvalidOTPEmail
		}
	}

	return nil
}

func safeDeliveryError(err error) error {
	if err == nil {
		return nil
	}

	// Do not wrap the transport error. SMTP replies can contain provider
	// topology, recipient data, or other details that must not cross this port.
	return ErrDeliveryUnavailable
}
