package config

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/runtimeenv"
)

// setValidEnvironment creates a complete valid configuration for each test.
//
// t.Setenv automatically restores the previous environment after the test,
// which prevents one test from changing another test's configuration.
func setValidEnvironment(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "local")
	t.Setenv("EXECUTION_ENV", "TEST")
	t.Setenv("API_HOST", "127.0.0.1")
	t.Setenv("API_PORT", "8081")
	t.Setenv("APP_ORIGIN", "http://127.0.0.1:3001")
	t.Setenv("AUTH_OTP_BYPASS", "")
	t.Setenv("AUTH_SESSION_TTL", "")
	t.Setenv("AUTH_TRUSTED_PROXY_CIDRS", "")
	t.Setenv(
		"AUTH_OTP_HMAC_SECRET",
		base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
	)
	t.Setenv("AUTH_EMAIL_MODE", "smtp")
	t.Setenv("AUTH_EMAIL_FROM_ADDRESS", "no-reply@example.com")
	t.Setenv("AUTH_EMAIL_FROM_NAME", "Sales Agent")
	t.Setenv("AUTH_SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_SMTP_PORT", "465")
	t.Setenv("AUTH_SMTP_TLS_MODE", "tls")
	t.Setenv("AUTH_SMTP_USERNAME", "smtp-user")
	t.Setenv("AUTH_SMTP_PASSWORD", "smtp-password")
	t.Setenv("AUTH_SMTP_TIMEOUT", "")

	t.Setenv(
		"DATABASE_URL",
		"postgres://sales_agent:sales_agent_local@127.0.0.1:5432/sales_agent?sslmode=disable",
	)

	t.Setenv(
		"REDIS_URL",
		"redis://127.0.0.1:6379/0",
	)
}

// TestLoadValidConfiguration verifies that valid environment variables become
// one validated Config object that the application can safely use.
func TestLoadValidConfiguration(t *testing.T) {
	setValidEnvironment(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.AppEnvironment != AppLocal {
		t.Fatalf(
			"AppEnvironment = %q, want %q",
			cfg.AppEnvironment,
			AppLocal,
		)
	}

	if cfg.ExecutionEnvironment != runtimeenv.Test {
		t.Fatalf(
			"ExecutionEnvironment = %q, want %q",
			cfg.ExecutionEnvironment,
			runtimeenv.Test,
		)
	}

	if cfg.APIAddress() != "127.0.0.1:8081" {
		t.Fatalf(
			"APIAddress() = %q, want %q",
			cfg.APIAddress(),
			"127.0.0.1:8081",
		)
	}

	if cfg.AppOrigin != "http://127.0.0.1:3001" {
		t.Fatalf("AppOrigin = %q, want %q", cfg.AppOrigin, "http://127.0.0.1:3001")
	}

	if cfg.AuthOTPBypass {
		t.Fatal("AuthOTPBypass should default to false")
	}

	if cfg.AuthSessionTTL != 8*time.Hour {
		t.Fatalf("AuthSessionTTL = %s, want %s", cfg.AuthSessionTTL, 8*time.Hour)
	}
	if len(cfg.AuthOTPHMACSecret) != 32 {
		t.Fatalf("len(AuthOTPHMACSecret) = %d, want 32", len(cfg.AuthOTPHMACSecret))
	}
	if cfg.AuthEmail.Mode != AuthEmailSMTP {
		t.Fatalf("AuthEmail.Mode = %q, want %q", cfg.AuthEmail.Mode, AuthEmailSMTP)
	}
	if cfg.AuthEmail.SMTPTLSMode != SMTPTLSDirect {
		t.Fatalf("AuthEmail.SMTPTLSMode = %q, want %q", cfg.AuthEmail.SMTPTLSMode, SMTPTLSDirect)
	}
	if cfg.AuthEmail.SMTPTimeout != defaultSMTPTimeout {
		t.Fatalf("AuthEmail.SMTPTimeout = %s, want %s", cfg.AuthEmail.SMTPTimeout, defaultSMTPTimeout)
	}
	if len(cfg.AuthTrustedProxyCIDRs) != 0 {
		t.Fatalf("AuthTrustedProxyCIDRs = %v, want none", cfg.AuthTrustedProxyCIDRs)
	}

	if cfg.CookieSecure() {
		t.Fatal("CookieSecure() should be false for local HTTP development")
	}
}

func TestLoadParsesTrustedProxyCIDRs(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv(
		"AUTH_TRUSTED_PROXY_CIDRS",
		" 127.0.0.1/32, 10.1.2.3/8, ::1/128, 10.0.0.0/8 ",
	)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	want := []string{"127.0.0.1/32", "10.0.0.0/8", "::1/128"}
	if len(cfg.AuthTrustedProxyCIDRs) != len(want) {
		t.Fatalf("AuthTrustedProxyCIDRs = %v, want %v", cfg.AuthTrustedProxyCIDRs, want)
	}
	for index, prefix := range cfg.AuthTrustedProxyCIDRs {
		if prefix.String() != want[index] {
			t.Fatalf("AuthTrustedProxyCIDRs[%d] = %q, want %q", index, prefix, want[index])
		}
	}
}

func TestLoadRejectsUnsafeTrustedProxyCIDRs(t *testing.T) {
	tests := []string{
		"127.0.0.1",
		"not-a-network",
		"127.0.0.1/32,,10.0.0.0/8",
		"0.0.0.0/0",
		"::/0",
		"::ffff:127.0.0.1/128",
	}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("AUTH_TRUSTED_PROXY_CIDRS", value)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() should reject AUTH_TRUSTED_PROXY_CIDRS=%q", value)
			}
		})
	}
}

func TestLoadAllowsLocalOTPBypass(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("AUTH_OTP_BYPASS", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if !cfg.AuthOTPBypass {
		t.Fatal("AuthOTPBypass = false, want true")
	}
	if cfg.AuthOTPHMACSecret != nil {
		t.Fatal("AuthOTPHMACSecret should not be loaded while the local bypass is enabled")
	}
	if cfg.AuthEmail != (AuthEmailConfig{}) {
		t.Fatal("AuthEmail should not be loaded while the local bypass is enabled")
	}
}

func TestLoadRejectsOTPBypassOutsideLocal(t *testing.T) {
	for _, appEnvironment := range []AppEnvironment{AppTest, AppProduction} {
		t.Run(string(appEnvironment), func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("APP_ENV", string(appEnvironment))
			t.Setenv("AUTH_OTP_BYPASS", "true")

			if appEnvironment == AppProduction {
				t.Setenv("APP_ORIGIN", "https://admin.example.com")
			}

			if _, err := Load(); err == nil {
				t.Fatalf("Load() should reject the OTP bypass in %s", appEnvironment)
			}
		})
	}
}

func TestLoadAcceptsDisabledOTPBypassOutsideLocal(t *testing.T) {
	for _, appEnvironment := range []AppEnvironment{AppTest, AppProduction} {
		t.Run(string(appEnvironment), func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("APP_ENV", string(appEnvironment))
			t.Setenv("AUTH_OTP_BYPASS", "false")

			if appEnvironment == AppProduction {
				t.Setenv("APP_ORIGIN", "https://admin.example.com")
			} else {
				setMailpitEmailEnvironment(t)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}

			if cfg.AuthOTPBypass {
				t.Fatal("AuthOTPBypass = true, want false")
			}

			if !cfg.CookieSecure() {
				t.Fatalf("CookieSecure() should be true in %s", appEnvironment)
			}
		})
	}
}

func TestLoadRejectsInvalidOTPBypass(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("AUTH_OTP_BYPASS", "yes")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a non-boolean OTP bypass value")
	}
}

func TestLoadValidatesAuthSessionTTL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "minimum", value: "15m", want: 15 * time.Minute},
		{name: "custom", value: "12h", want: 12 * time.Hour},
		{name: "maximum", value: "24h", want: 24 * time.Hour},
		{name: "below minimum", value: "14m59s", wantErr: true},
		{name: "above maximum", value: "24h1s", wantErr: true},
		{name: "invalid duration", value: "eight hours", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("AUTH_SESSION_TTL", tt.value)

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() should reject AUTH_SESSION_TTL=%q", tt.value)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}
			if cfg.AuthSessionTTL != tt.want {
				t.Fatalf("AuthSessionTTL = %s, want %s", cfg.AuthSessionTTL, tt.want)
			}
		})
	}
}

func TestLoadValidatesAppOrigin(t *testing.T) {
	tests := []struct {
		name           string
		appEnvironment AppEnvironment
		origin         string
	}{
		{name: "missing origin", appEnvironment: AppLocal},
		{name: "relative origin", appEnvironment: AppLocal, origin: "/admin"},
		{name: "unsupported scheme", appEnvironment: AppLocal, origin: "ftp://example.com"},
		{name: "credentials", appEnvironment: AppLocal, origin: "http://user:password@example.com"},
		{name: "invalid port", appEnvironment: AppLocal, origin: "http://example.com:65536"},
		{name: "path", appEnvironment: AppLocal, origin: "http://example.com/admin"},
		{name: "query", appEnvironment: AppLocal, origin: "http://example.com?mode=admin"},
		{name: "fragment", appEnvironment: AppLocal, origin: "http://example.com#admin"},
		{name: "production HTTP", appEnvironment: AppProduction, origin: "http://admin.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("APP_ENV", string(tt.appEnvironment))
			t.Setenv("APP_ORIGIN", tt.origin)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() should reject APP_ORIGIN=%q in %s", tt.origin, tt.appEnvironment)
			}
		})
	}
}

func TestLoadAcceptsProductionHTTPSOrigin(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_ENV", string(AppProduction))
	t.Setenv("APP_ORIGIN", "https://admin.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if !cfg.CookieSecure() {
		t.Fatal("CookieSecure() should be true in production")
	}
}

// TestLoadRejectsInvalidExecutionEnvironment proves that an unknown execution
// environment cannot silently fall back to TEST or PRODUCTION.
func TestLoadRejectsInvalidExecutionEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("EXECUTION_ENV", "something-else")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an invalid execution environment")
	}
}

// TestLoadRejectsInvalidPort proves that malformed network configuration is
// rejected during startup rather than failing later while serving requests.
func TestLoadRejectsInvalidPort(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("API_PORT", "99999")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an invalid API port")
	}
}

// TestLoadRejectsMissingDatabaseURL proves that the application cannot start
// without an explicitly configured PostgreSQL database.
func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a missing database URL")
	}
}

// TestLoadRejectsInvalidDatabaseScheme prevents accidental configuration with
// an HTTP endpoint or another non-PostgreSQL service.
func TestLoadRejectsInvalidDatabaseScheme(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_URL", "https://example.com/database")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a non-PostgreSQL database URL")
	}
}

// TestLoadRejectsMissingRedisURL proves that Redis must also be configured
// explicitly before the application considers configuration valid.
func TestLoadRejectsMissingRedisURL(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("REDIS_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a missing Redis URL")
	}
}
