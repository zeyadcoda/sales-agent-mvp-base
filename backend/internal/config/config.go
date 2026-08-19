package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

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
		AppEnvironment:       appEnvironment,
		ExecutionEnvironment: executionEnvironment,
		APIHost:              apiHost,
		APIPort:              apiPort,
		DatabaseURL:          databaseURL,
		RedisURL:             redisURL,
	}, nil
}

// APIAddress builds the HTTP listen address safely for IPv4, IPv6, or hostnames.
func (c Config) APIAddress() string {
	return net.JoinHostPort(c.APIHost, strconv.Itoa(c.APIPort))
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
