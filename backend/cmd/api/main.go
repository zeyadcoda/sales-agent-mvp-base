package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/httpapi"
)

func main() {
	// Keep main intentionally small. run contains the application startup
	// lifecycle so deferred cleanup still executes before the process exits.
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := "127.0.0.1:8081"

	if value := os.Getenv("API_ADDR"); value != "" {
		addr = value
	}

	// Limit startup initialization so configuration problems cannot hang the
	// application indefinitely.
	startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// DATABASE_URL comes from server-side configuration only. It is never
	// supplied by the browser or an AI Agent.
	db, err := database.Open(startupCtx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	server := &http.Server{
		Addr:    addr,
		Handler: httpapi.NewRouter(db),

		// ReadHeaderTimeout reduces exposure to clients that intentionally send
		// HTTP headers extremely slowly to consume server resources.
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("sales-agent API listening on %s", addr)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve API: %w", err)
	}

	return nil
}
