package config

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"salesagent.local/backend/internal/runtimeenv"
)

// AppEnvironment describes where this application instance is running.
//
// This is separate from ExecutionEnvironment because a local or production
// application may still run TEST Agent executions through isolated workers.
type AppEnvironment string

const (
	AppLocal      AppEnvironment = "local"
	AppTest       AppEnvironment = "test"
	AppProduction AppEnvironment = "production"

	defaultAuthSessionTTL = 8 * time.Hour
	minimumAuthSessionTTL = 15 * time.Minute
	maximumAuthSessionTTL = 24 * time.Hour
)

// Config contains validated server-side configuration.
//
// Some values may contain credentials, so the complete Config object must
// never be written to logs.
type Config struct {
	AppEnvironment       AppEnvironment
	ExecutionEnvironment runtimeenv.ExecutionEnvironment

	APIHost string
	APIPort int
	// AppOrigin is the exact browser origin allowed to make authenticated
	// state-changing requests. It is server configuration, never client input.
	AppOrigin string

	// AuthOTPBypass is a temporary local-development control. Load rejects it
	// outside APP_ENV=local so it cannot become a production auth path.
	AuthOTPBypass  bool
	AuthSessionTTL time.Duration
	// AuthTrustedProxyCIDRs defines the only network peers whose forwarded
	// client-address chain the authentication rate limiter may trust.
	AuthTrustedProxyCIDRs []netip.Prefix

	DatabaseURL string
	RedisURL    string
}

// Load reads environment variables once during startup and validates them.
//
// Centralizing configuration prevents random packages from reading environment
// variables independently and accidentally applying inconsistent defaults.
func Load() (Config, error) {
	appEnvironment, err := parseAppEnvironment(os.Getenv("APP_ENV"))
	if err != nil {
		return Config{}, err
	}

	executionEnvironment, err := runtimeenv.Parse(os.Getenv("EXECUTION_ENV"))
	if err != nil {
		return Config{}, fmt.Errorf("EXECUTION_ENV: %w", err)
	}

	apiHost := strings.TrimSpace(os.Getenv("API_HOST"))
	if apiHost == "" {
		return Config{}, fmt.Errorf("API_HOST is required")
	}

	apiPort, err := parsePort(os.Getenv("API_PORT"))
	if err != nil {
		return Config{}, err
	}

	appOrigin, err := parseAppOrigin(os.Getenv("APP_ORIGIN"), appEnvironment)
	if err != nil {
		return Config{}, err
	}

	authOTPBypass, err := parseOptionalBool("AUTH_OTP_BYPASS", os.Getenv("AUTH_OTP_BYPASS"))
	if err != nil {
		return Config{}, err
	}

	// The development bypass is intentionally guarded by the deployment
	// environment at startup. Frontend state and execution-plane selection
	// cannot weaken this boundary.
	if authOTPBypass && appEnvironment != AppLocal {
		return Config{}, fmt.Errorf("AUTH_OTP_BYPASS may be enabled only when APP_ENV=local")
	}

	authSessionTTL, err := parseAuthSessionTTL(os.Getenv("AUTH_SESSION_TTL"))
	if err != nil {
		return Config{}, err
	}

	authTrustedProxyCIDRs, err := parseTrustedProxyCIDRs(os.Getenv("AUTH_TRUSTED_PROXY_CIDRS"))
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := validateURL(
		"DATABASE_URL",
		os.Getenv("DATABASE_URL"),
		"postgres",
		"postgresql",
	)
	if err != nil {
		return Config{}, err
	}

	redisURL, err := validateURL(
		"REDIS_URL",
		os.Getenv("REDIS_URL"),
		"redis",
		"rediss",
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnvironment:        appEnvironment,
		ExecutionEnvironment:  executionEnvironment,
		APIHost:               apiHost,
		APIPort:               apiPort,
		AppOrigin:             appOrigin,
		AuthOTPBypass:         authOTPBypass,
		AuthSessionTTL:        authSessionTTL,
		AuthTrustedProxyCIDRs: authTrustedProxyCIDRs,
		DatabaseURL:           databaseURL,
		RedisURL:              redisURL,
	}, nil
}

// APIAddress builds the HTTP listen address safely for IPv4, IPv6, or hostnames.
func (c Config) APIAddress() string {
	return net.JoinHostPort(c.APIHost, strconv.Itoa(c.APIPort))
}

// CookieSecure reports whether authentication cookies must be restricted to
// HTTPS. Local HTTP is the sole exception needed for loopback development.
func (c Config) CookieSecure() bool {
	return c.AppEnvironment != AppLocal
}

// parseAppEnvironment accepts only known application environments.
//
// Unknown values fail closed rather than silently choosing a potentially unsafe
// environment.
func parseAppEnvironment(value string) (AppEnvironment, error) {
	switch AppEnvironment(strings.ToLower(strings.TrimSpace(value))) {
	case AppLocal:
		return AppLocal, nil
	case AppTest:
		return AppTest, nil
	case AppProduction:
		return AppProduction, nil
	default:
		return "", fmt.Errorf("APP_ENV must be local, test, or production")
	}
}

// parsePort validates the API port before the server starts.
func parsePort(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("API_PORT must be a valid TCP port")
	}

	return port, nil
}

// parseAppOrigin accepts one exact HTTP(S) origin rather than a URL prefix.
// Credentials, paths, queries, and fragments would make Origin comparison
// ambiguous and are rejected during startup.
func parseAppOrigin(value string, appEnvironment AppEnvironment) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("APP_ORIGIN is required")
	}

	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.Hostname() == "" || parsed.Opaque != "" {
		return "", fmt.Errorf("APP_ORIGIN must be an absolute HTTP or HTTPS origin")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("APP_ORIGIN must use HTTP or HTTPS")
	}

	if strings.HasSuffix(parsed.Host, ":") {
		return "", fmt.Errorf("APP_ORIGIN has an invalid TCP port")
	}
	if port := parsed.Port(); port != "" {
		portNumber, portErr := strconv.Atoi(port)
		if portErr != nil || portNumber < 1 || portNumber > 65535 {
			return "", fmt.Errorf("APP_ORIGIN has an invalid TCP port")
		}
	}

	if parsed.User != nil || parsed.Path != "" || parsed.RawPath != "" || parsed.ForceQuery || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.RawFragment != "" {
		return "", fmt.Errorf("APP_ORIGIN must not contain credentials, a path, query, or fragment")
	}

	if appEnvironment == AppProduction && parsed.Scheme != "https" {
		return "", fmt.Errorf("APP_ORIGIN must use HTTPS when APP_ENV=production")
	}

	return value, nil
}

func parseOptionalBool(name string, value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, fmt.Errorf("%s must be true or false", name)
	}
}

func parseAuthSessionTTL(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultAuthSessionTTL, nil
	}

	ttl, err := time.ParseDuration(value)
	if err != nil || ttl < minimumAuthSessionTTL || ttl > maximumAuthSessionTTL {
		return 0, fmt.Errorf(
			"AUTH_SESSION_TTL must be a duration between %s and %s",
			minimumAuthSessionTTL,
			maximumAuthSessionTTL,
		)
	}

	return ttl, nil
}

func parseTrustedProxyCIDRs(value string) ([]netip.Prefix, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		prefix, err := netip.ParsePrefix(part)
		if err != nil || prefix.Bits() == 0 || prefix.Addr().Is4In6() {
			// Reject an all-addresses range and IPv4-mapped IPv6 notation because
			// either can make a trusted-proxy boundary broader than it appears.
			return nil, fmt.Errorf("AUTH_TRUSTED_PROXY_CIDRS must contain valid, bounded IP CIDRs")
		}

		prefix = prefix.Masked()
		key := prefix.String()
		if _, duplicate := seen[key]; duplicate {
			continue
		}

		seen[key] = struct{}{}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}

// validateURL validates infrastructure URLs without returning their raw value
// in errors. This avoids leaking passwords or tokens embedded inside URLs.
func validateURL(name string, value string, allowedSchemes ...string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("%s has an invalid URL", name)
	}

	for _, scheme := range allowedSchemes {
		if parsed.Scheme == scheme {
			return value, nil
		}
	}

	return "", fmt.Errorf("%s uses an unsupported URL scheme", name)
}
