package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	minimumOTPHMACSecretBytes = 32
	defaultSMTPTimeout        = 10 * time.Second
	minimumSMTPTimeout        = time.Second
	maximumSMTPTimeout        = 30 * time.Second
)

// AuthEmailMode selects a deliberately narrow authentication-email adapter.
type AuthEmailMode string

const (
	AuthEmailMailpit AuthEmailMode = "mailpit"
	AuthEmailSMTP    AuthEmailMode = "smtp"
)

// SMTPTLSMode describes transport security for an SMTP connection.
type SMTPTLSMode string

const (
	SMTPTLSNone     SMTPTLSMode = "none"
	SMTPTLSSTARTTLS SMTPTLSMode = "starttls"
	SMTPTLSDirect   SMTPTLSMode = "tls"
)

// AuthEmailConfig contains validated delivery settings. SMTPPassword is a
// secret and, like the complete Config object, must never be logged.
type AuthEmailConfig struct {
	Mode         AuthEmailMode
	FromAddress  string
	FromName     string
	SMTPHost     string
	SMTPPort     int
	SMTPTLSMode  SMTPTLSMode
	SMTPUsername string
	SMTPPassword string
	SMTPTimeout  time.Duration
}

func loadAuthOTPConfiguration(appEnvironment AppEnvironment) ([]byte, AuthEmailConfig, error) {
	secret, err := parseOTPHMACSecret(os.Getenv("AUTH_OTP_HMAC_SECRET"))
	if err != nil {
		return nil, AuthEmailConfig{}, err
	}

	emailConfig, err := parseAuthEmailConfig(appEnvironment)
	if err != nil {
		return nil, AuthEmailConfig{}, err
	}

	return secret, emailConfig, nil
}

func parseOTPHMACSecret(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("AUTH_OTP_HMAC_SECRET is required when AUTH_OTP_BYPASS=false")
	}

	secret, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Raw standard base64 is accepted so secret managers that omit padding do
		// not need to transform otherwise valid key material.
		secret, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil || len(secret) < minimumOTPHMACSecretBytes {
		return nil, fmt.Errorf("AUTH_OTP_HMAC_SECRET must be base64-encoded key material of at least %d bytes", minimumOTPHMACSecretBytes)
	}

	return secret, nil
}

func parseAuthEmailConfig(appEnvironment AppEnvironment) (AuthEmailConfig, error) {
	mode, err := parseAuthEmailMode(os.Getenv("AUTH_EMAIL_MODE"))
	if err != nil {
		return AuthEmailConfig{}, err
	}
	if appEnvironment == AppProduction && mode != AuthEmailSMTP {
		return AuthEmailConfig{}, fmt.Errorf("AUTH_EMAIL_MODE must be smtp when APP_ENV=production")
	}
	if appEnvironment == AppTest && mode != AuthEmailMailpit {
		return AuthEmailConfig{}, fmt.Errorf("AUTH_EMAIL_MODE must be mailpit when APP_ENV=test")
	}

	fromAddress, err := parseMailbox("AUTH_EMAIL_FROM_ADDRESS", os.Getenv("AUTH_EMAIL_FROM_ADDRESS"))
	if err != nil {
		return AuthEmailConfig{}, err
	}
	fromName, err := parseEmailDisplayName(os.Getenv("AUTH_EMAIL_FROM_NAME"))
	if err != nil {
		return AuthEmailConfig{}, err
	}

	host, err := parseSMTPHost(os.Getenv("AUTH_SMTP_HOST"))
	if err != nil {
		return AuthEmailConfig{}, err
	}
	port, err := parseNamedPort("AUTH_SMTP_PORT", os.Getenv("AUTH_SMTP_PORT"))
	if err != nil {
		return AuthEmailConfig{}, err
	}
	tlsMode, err := parseSMTPTLSMode(os.Getenv("AUTH_SMTP_TLS_MODE"))
	if err != nil {
		return AuthEmailConfig{}, err
	}
	timeout, err := parseSMTPTimeout(os.Getenv("AUTH_SMTP_TIMEOUT"))
	if err != nil {
		return AuthEmailConfig{}, err
	}

	username := strings.TrimSpace(os.Getenv("AUTH_SMTP_USERNAME"))
	password := os.Getenv("AUTH_SMTP_PASSWORD")
	if (username == "") != (password == "") {
		return AuthEmailConfig{}, fmt.Errorf("AUTH_SMTP_USERNAME and AUTH_SMTP_PASSWORD must either both be set or both be empty")
	}
	if strings.ContainsAny(username, "\r\n\x00") || strings.ContainsAny(password, "\r\n\x00") {
		return AuthEmailConfig{}, fmt.Errorf("AUTH_SMTP credentials contain invalid characters")
	}

	switch mode {
	case AuthEmailMailpit:
		if appEnvironment != AppLocal && appEnvironment != AppTest {
			return AuthEmailConfig{}, fmt.Errorf("AUTH_EMAIL_MODE=mailpit is allowed only in local or test environments")
		}
		if host != "127.0.0.1" || port != 1025 || tlsMode != SMTPTLSNone {
			return AuthEmailConfig{}, fmt.Errorf("mailpit email must use 127.0.0.1:1025 with AUTH_SMTP_TLS_MODE=none")
		}
		if username != "" {
			return AuthEmailConfig{}, fmt.Errorf("mailpit email must not configure SMTP credentials")
		}

	case AuthEmailSMTP:
		if tlsMode != SMTPTLSSTARTTLS && tlsMode != SMTPTLSDirect {
			return AuthEmailConfig{}, fmt.Errorf("AUTH_SMTP_TLS_MODE must be starttls or tls when AUTH_EMAIL_MODE=smtp")
		}
		if appEnvironment == AppProduction && username == "" {
			return AuthEmailConfig{}, fmt.Errorf("authenticated SMTP credentials are required when APP_ENV=production")
		}
	}

	return AuthEmailConfig{
		Mode:         mode,
		FromAddress:  fromAddress,
		FromName:     fromName,
		SMTPHost:     host,
		SMTPPort:     port,
		SMTPTLSMode:  tlsMode,
		SMTPUsername: username,
		SMTPPassword: password,
		SMTPTimeout:  timeout,
	}, nil
}

func parseAuthEmailMode(value string) (AuthEmailMode, error) {
	switch AuthEmailMode(strings.ToLower(strings.TrimSpace(value))) {
	case AuthEmailMailpit:
		return AuthEmailMailpit, nil
	case AuthEmailSMTP:
		return AuthEmailSMTP, nil
	default:
		return "", fmt.Errorf("AUTH_EMAIL_MODE must be mailpit or smtp")
	}
}

func parseMailbox(name string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		return "", fmt.Errorf("%s must be a valid email address", name)
	}

	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Name != "" || parsed.Address != value {
		return "", fmt.Errorf("%s must be a valid email address without a display name", name)
	}

	return value, nil
}

func parseEmailDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 100 {
		return "", fmt.Errorf("AUTH_EMAIL_FROM_NAME must contain between 1 and 100 characters")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", fmt.Errorf("AUTH_EMAIL_FROM_NAME contains invalid characters")
		}
	}

	return value, nil
}

func parseSMTPHost(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, " \t\r\n/\\@?#\x00") {
		return "", fmt.Errorf("AUTH_SMTP_HOST must be a hostname or IP address without a scheme or port")
	}

	if strings.ContainsAny(value, "[]") {
		return "", fmt.Errorf("AUTH_SMTP_HOST must be a hostname or IP address without brackets")
	}
	if net.ParseIP(value) != nil {
		return value, nil
	}
	if !validDNSHostname(value) {
		return "", fmt.Errorf("AUTH_SMTP_HOST must be a valid DNS hostname or IP address")
	}

	return value, nil
}

func validDNSHostname(value string) bool {
	if len(value) > 253 || strings.HasSuffix(value, ".") {
		return false
	}

	for _, label := range strings.Split(value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') &&
				(character < 'A' || character > 'Z') &&
				(character < '0' || character > '9') &&
				character != '-' {
				return false
			}
		}
	}

	return true
}

func parseNamedPort(name string, value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be a valid TCP port", name)
	}

	return port, nil
}

func parseSMTPTLSMode(value string) (SMTPTLSMode, error) {
	switch SMTPTLSMode(strings.ToLower(strings.TrimSpace(value))) {
	case SMTPTLSNone:
		return SMTPTLSNone, nil
	case SMTPTLSSTARTTLS:
		return SMTPTLSSTARTTLS, nil
	case SMTPTLSDirect:
		return SMTPTLSDirect, nil
	default:
		return "", fmt.Errorf("AUTH_SMTP_TLS_MODE must be none, starttls, or tls")
	}
}

func parseSMTPTimeout(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultSMTPTimeout, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil || timeout < minimumSMTPTimeout || timeout > maximumSMTPTimeout {
		return 0, fmt.Errorf("AUTH_SMTP_TIMEOUT must be a duration between %s and %s", minimumSMTPTimeout, maximumSMTPTimeout)
	}

	return timeout, nil
}
