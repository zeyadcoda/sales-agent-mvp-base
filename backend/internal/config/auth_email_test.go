package config

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func setMailpitEmailEnvironment(t *testing.T) {
	t.Helper()

	t.Setenv("AUTH_EMAIL_MODE", "mailpit")
	t.Setenv("AUTH_EMAIL_FROM_ADDRESS", "no-reply@sales-agent.local")
	t.Setenv("AUTH_EMAIL_FROM_NAME", "Sales Agent Local")
	t.Setenv("AUTH_SMTP_HOST", "127.0.0.1")
	t.Setenv("AUTH_SMTP_PORT", "1025")
	t.Setenv("AUTH_SMTP_TLS_MODE", "none")
	t.Setenv("AUTH_SMTP_USERNAME", "")
	t.Setenv("AUTH_SMTP_PASSWORD", "")
}

func TestLoadAcceptsLocalMailpitConfiguration(t *testing.T) {
	setValidEnvironment(t)
	setMailpitEmailEnvironment(t)
	t.Setenv("AUTH_SMTP_TIMEOUT", "5s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.AuthEmail.Mode != AuthEmailMailpit {
		t.Fatalf("AuthEmail.Mode = %q, want %q", cfg.AuthEmail.Mode, AuthEmailMailpit)
	}
	if cfg.AuthEmail.SMTPHost != "127.0.0.1" || cfg.AuthEmail.SMTPPort != 1025 {
		t.Fatalf("Mailpit endpoint = %s:%d, want 127.0.0.1:1025", cfg.AuthEmail.SMTPHost, cfg.AuthEmail.SMTPPort)
	}
	if cfg.AuthEmail.SMTPTimeout != 5*time.Second {
		t.Fatalf("AuthEmail.SMTPTimeout = %s, want 5s", cfg.AuthEmail.SMTPTimeout)
	}
}

func TestLoadAcceptsMailpitInTestEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_ENV", "test")
	setMailpitEmailEnvironment(t)

	if _, err := Load(); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
}

func TestLoadRejectsProviderSMTPInTestEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_ENV", "test")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a production-capable SMTP adapter in test")
	}
}

func TestLoadAllowsLocalBypassWithoutOTPOrEmailConfiguration(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("AUTH_OTP_BYPASS", "true")
	for _, name := range []string{
		"AUTH_OTP_HMAC_SECRET",
		"AUTH_EMAIL_MODE",
		"AUTH_EMAIL_FROM_ADDRESS",
		"AUTH_EMAIL_FROM_NAME",
		"AUTH_SMTP_HOST",
		"AUTH_SMTP_PORT",
		"AUTH_SMTP_TLS_MODE",
		"AUTH_SMTP_USERNAME",
		"AUTH_SMTP_PASSWORD",
		"AUTH_SMTP_TIMEOUT",
	} {
		t.Setenv(name, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.AuthOTPHMACSecret != nil || cfg.AuthEmail != (AuthEmailConfig{}) {
		t.Fatal("unused OTP/email configuration should remain empty in bypass mode")
	}
}

func TestLoadValidatesOTPHMACSecret(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "missing"},
		{name: "not base64", value: "not-base64"},
		{name: "too short", value: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", 31)))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("AUTH_OTP_HMAC_SECRET", tt.value)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() should reject AUTH_OTP_HMAC_SECRET=%q", tt.value)
			}
		})
	}
}

func TestLoadAcceptsUnpaddedOTPHMACSecret(t *testing.T) {
	setValidEnvironment(t)
	secret := []byte(strings.Repeat("s", 32))
	t.Setenv("AUTH_OTP_HMAC_SECRET", base64.RawStdEncoding.EncodeToString(secret))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if string(cfg.AuthOTPHMACSecret) != string(secret) {
		t.Fatal("AuthOTPHMACSecret did not contain the decoded key material")
	}
}

func TestLoadDoesNotExposeOTPOrSMTPSecretsInErrors(t *testing.T) {
	setValidEnvironment(t)
	otpSecret := "sensitive-secret-that-is-not-base64"
	t.Setenv("AUTH_OTP_HMAC_SECRET", otpSecret)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should reject an invalid OTP HMAC secret")
	}
	if strings.Contains(err.Error(), otpSecret) {
		t.Fatal("configuration error exposed the OTP HMAC secret")
	}

	setValidEnvironment(t)
	smtpPassword := "sensitive\npassword"
	t.Setenv("AUTH_SMTP_PASSWORD", smtpPassword)
	_, err = Load()
	if err == nil {
		t.Fatal("Load() should reject an invalid SMTP password")
	}
	if strings.Contains(err.Error(), smtpPassword) {
		t.Fatal("configuration error exposed the SMTP password")
	}
}

func TestLoadRejectsUnsafeMailpitConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
	}{
		{name: "non-loopback host", envName: "AUTH_SMTP_HOST", envValue: "mailpit"},
		{name: "wrong port", envName: "AUTH_SMTP_PORT", envValue: "25"},
		{name: "TLS", envName: "AUTH_SMTP_TLS_MODE", envValue: "starttls"},
		{name: "credentials", envName: "AUTH_SMTP_USERNAME", envValue: "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("AUTH_EMAIL_MODE", "mailpit")
			t.Setenv("AUTH_SMTP_HOST", "127.0.0.1")
			t.Setenv("AUTH_SMTP_PORT", "1025")
			t.Setenv("AUTH_SMTP_TLS_MODE", "none")
			t.Setenv("AUTH_SMTP_USERNAME", "")
			t.Setenv("AUTH_SMTP_PASSWORD", "")
			t.Setenv(tt.envName, tt.envValue)
			if tt.envName == "AUTH_SMTP_USERNAME" {
				t.Setenv("AUTH_SMTP_PASSWORD", "password")
			}

			if _, err := Load(); err == nil {
				t.Fatal("Load() should reject unsafe Mailpit configuration")
			}
		})
	}
}

func TestLoadRejectsMailpitInProduction(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_ORIGIN", "https://admin.example.com")
	setMailpitEmailEnvironment(t)

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject Mailpit in production")
	}
}

func TestLoadRequiresTLSForSMTP(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("AUTH_SMTP_TLS_MODE", "none")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject plaintext provider SMTP")
	}
}

func TestLoadRequiresCompleteSMTPCredentials(t *testing.T) {
	for _, missing := range []string{"AUTH_SMTP_USERNAME", "AUTH_SMTP_PASSWORD"} {
		t.Run(missing, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv(missing, "")

			if _, err := Load(); err == nil {
				t.Fatal("Load() should reject incomplete SMTP credentials")
			}
		})
	}
}

func TestLoadRequiresAuthenticatedSMTPInProduction(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_ORIGIN", "https://admin.example.com")
	t.Setenv("AUTH_SMTP_USERNAME", "")
	t.Setenv("AUTH_SMTP_PASSWORD", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject unauthenticated production SMTP")
	}
}

func TestLoadRejectsInvalidEmailHeadersAndSMTPValues(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
	}{
		{name: "from address display name", envName: "AUTH_EMAIL_FROM_ADDRESS", envValue: "Sender <sender@example.com>"},
		{name: "from address injection", envName: "AUTH_EMAIL_FROM_ADDRESS", envValue: "sender@example.com\r\nBcc: victim@example.com"},
		{name: "from name injection", envName: "AUTH_EMAIL_FROM_NAME", envValue: "Sender\nBcc"},
		{name: "SMTP URL", envName: "AUTH_SMTP_HOST", envValue: "smtp://example.com"},
		{name: "SMTP host with port", envName: "AUTH_SMTP_HOST", envValue: "smtp.example.com:587"},
		{name: "SMTP empty label", envName: "AUTH_SMTP_HOST", envValue: "smtp..example.com"},
		{name: "SMTP leading hyphen", envName: "AUTH_SMTP_HOST", envValue: "-smtp.example.com"},
		{name: "invalid port", envName: "AUTH_SMTP_PORT", envValue: "65536"},
		{name: "invalid TLS mode", envName: "AUTH_SMTP_TLS_MODE", envValue: "optional"},
		{name: "short timeout", envName: "AUTH_SMTP_TIMEOUT", envValue: "999ms"},
		{name: "long timeout", envName: "AUTH_SMTP_TIMEOUT", envValue: "31s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv(tt.envName, tt.envValue)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() should reject %s=%q", tt.envName, tt.envValue)
			}
		})
	}
}
