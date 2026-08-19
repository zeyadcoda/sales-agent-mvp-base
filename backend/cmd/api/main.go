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
	"salesagent.local/backend/internal/readiness"
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

	server := &http.Server{
		Addr:    cfg.APIAddress(),
		Handler: httpapi.NewRouter(readinessChecker),

		// ReadHeaderTimeout reduces exposure to Slowloris-style clients that
		// intentionally send HTTP headers extremely slowly to consume server
		// connections.
		ReadHeaderTimeout: 5 * time.Second,
	}

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
