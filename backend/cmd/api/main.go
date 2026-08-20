package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"salesagent.local/backend/internal/cache"
	"salesagent.local/backend/internal/config"
	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/httpapi"
	notificationemail "salesagent.local/backend/internal/notification/email"
	"salesagent.local/backend/internal/platform/auth"
	"salesagent.local/backend/internal/readiness"
)

const (
	apiReadHeaderTimeout = 5 * time.Second
	apiReadTimeout       = 15 * time.Second
	// Authentication may include a provider SMTP operation bounded at 30
	// seconds plus the final PostgreSQL activation. Keep the transport budget
	// above the handler budget so valid configured delivery cannot be cut off.
	apiWriteTimeout   = 45 * time.Second
	apiIdleTimeout    = 60 * time.Second
	apiMaxHeaderBytes = 32 * 1024
)

func main() {
	// Keep main intentionally small.
	// run() owns startup and cleanup so infrastructure resources are always
	// released correctly before the process exits.
	processCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(processCtx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	// Read and validate all server-side configuration before starting
	// infrastructure or accepting HTTP requests.
	//
	// This prevents individual packages from reading arbitrary environment
	// variables and applying inconsistent or unsafe defaults.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Bound startup initialization so an unavailable dependency cannot leave
	// application startup hanging indefinitely.
	startupCtx, cancel := context.WithTimeout(
		ctx,
		5*time.Second,
	)
	defer cancel()

	// The database URL comes only from validated server-side configuration.
	//
	// A browser request or AI Agent must never be able to choose which
	// database the application connects to.
	db, err := database.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	// Redis is created only from validated server-side configuration. Neither
	// browser input nor an AI Agent may select an infrastructure endpoint.
	redisClient, err := cache.Open(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("initialize Redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			// Avoid logging the client error because dependency errors can include
			// connection metadata. The full Redis URL must never reach logs.
			log.Print("sales-agent API could not close Redis cleanly")
		}
	}()

	readinessChecker := readiness.New(db, redisClient)

	authStore, err := auth.NewPostgresStore(db)
	if err != nil {
		return fmt.Errorf("initialize authentication repository: %w", err)
	}

	passwordHasher := auth.NewPasswordHasher()
	// Unknown accounts are verified against a real hash so the most obvious
	// password-timing distinction does not disclose whether an email exists.
	// The dummy password is not an account credential and is never persisted.
	dummyPasswordHash, err := passwordHasher.Hash("timing-defense-only-not-a-real-password")
	if err != nil {
		return errors.New("initialize password timing defense")
	}

	var otpEmailSender auth.OTPEmailSender
	var otpRateLimiter auth.OTPRateLimiter
	if !cfg.AuthOTPBypass {
		otpEmailSender, err = notificationemail.NewSender(cfg.AuthEmail)
		if err != nil {
			return fmt.Errorf("initialize authentication email sender: %w", err)
		}
		otpRateLimiter = auth.NewOTPRateLimiter(redisClient)
	}

	authService, err := auth.NewService(
		authStore,
		auth.NewLoginRateLimiter(redisClient),
		otpRateLimiter,
		passwordHasher,
		otpEmailSender,
		auth.ServiceOptions{
			OTPBypassEnabled:  cfg.AuthOTPBypass,
			OTPHashSecret:     cfg.AuthOTPHMACSecret,
			SessionTTL:        cfg.AuthSessionTTL,
			DummyPasswordHash: dummyPasswordHash,
		},
	)
	if err != nil {
		return fmt.Errorf("initialize authentication service: %w", err)
	}

	authHandler, err := httpapi.NewAuthHandler(authService, httpapi.AuthHandlerOptions{
		ApplicationOrigin: cfg.AppOrigin,
		CookieSecure:      cfg.CookieSecure(),
		SessionTTL:        cfg.AuthSessionTTL,
		LocalDevelopment:  cfg.AppEnvironment == config.AppLocal,
		TrustedProxyCIDRs: cfg.AuthTrustedProxyCIDRs,
	})
	if err != nil {
		return fmt.Errorf("initialize authentication HTTP handler: %w", err)
	}

	dashboardHandler, err := httpapi.NewDashboardHandler(authService, readinessChecker)
	if err != nil {
		return fmt.Errorf("initialize dashboard HTTP handler: %w", err)
	}

	server := newAPIServer(
		cfg.APIAddress(),
		httpapi.NewRouter(readinessChecker, authHandler, dashboardHandler),
	)

	// Bind synchronously so shutdown can always close a known listener, even
	// when the process context is canceled just as serving begins.
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen for API requests: %w", err)
	}

	// Environment names are safe to log and useful for operations.
	//
	// Never log the complete Config object because it contains infrastructure
	// URLs that may contain credentials.
	log.Printf(
		"sales-agent API listening on %s (app_env=%s execution_env=%s)",
		cfg.APIAddress(),
		cfg.AppEnvironment,
		cfg.ExecutionEnvironment,
	)

	return serveUntilShutdown(ctx, server, listener)
}

func newAPIServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: handler,

		// These transport bounds complement endpoint body limits. They cap slow
		// reads, stalled handlers/writes, idle keep-alives, and oversized headers.
		ReadHeaderTimeout: apiReadHeaderTimeout,
		ReadTimeout:       apiReadTimeout,
		WriteTimeout:      apiWriteTimeout,
		IdleTimeout:       apiIdleTimeout,
		MaxHeaderBytes:    apiMaxHeaderBytes,
	}
}

// serveUntilShutdown drains in-flight HTTP requests before run returns and its
// dependency cleanup executes. Without signal handling, normal container
// termination would bypass deferred PostgreSQL and Redis cleanup.
func serveUntilShutdown(
	ctx context.Context,
	server *http.Server,
	listener net.Listener,
) error {
	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- server.Serve(listener)
	}()

	select {
	case err := <-serveErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve API: %w", err)
		}

		return nil

	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// Stop accepting/serving connections before dependency defers run even
		// when graceful draining exceeds its deadline.
		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(
				fmt.Errorf("gracefully shut down API: %w", err),
				fmt.Errorf("force close API: %w", closeErr),
			)
		}

		return fmt.Errorf("gracefully shut down API: %w", err)
	}

	if err := <-serveErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve API during shutdown: %w", err)
	}

	return nil
}
