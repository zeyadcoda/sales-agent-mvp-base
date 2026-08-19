package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"salesagent.local/backend/internal/config"
	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/httpapi"
)

func main() {
	// Keep main intentionally small.
	// run() owns startup and cleanup so resources such as database
	// connections are always released correctly before the process exits.
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
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
		context.Background(),
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

	server := &http.Server{
		Addr:    cfg.APIAddress(),
		Handler: httpapi.NewRouter(db),

		// ReadHeaderTimeout reduces exposure to Slowloris-style clients that
		// intentionally send HTTP headers extremely slowly to consume server
		// connections.
		ReadHeaderTimeout: 5 * time.Second,
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

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve API: %w", err)
	}

	return nil
}
