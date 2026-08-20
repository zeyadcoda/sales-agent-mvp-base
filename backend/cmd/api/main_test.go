package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestServeUntilShutdownWithCanceledContext guards the startup/shutdown race:
// cancellation must close a synchronously bound listener even if Serve has not
// begun accepting requests yet.
func TestServeUntilShutdownWithCanceledContext(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for test server: %v", err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := make(chan error, 1)
	go func() {
		result <- serveUntilShutdown(ctx, &http.Server{}, listener)
	}()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("serveUntilShutdown() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("serveUntilShutdown() did not stop after context cancellation")
	}
}

func TestAPIServerHasBoundedTransportSettings(t *testing.T) {
	t.Parallel()

	server := newAPIServer("127.0.0.1:8081", http.NotFoundHandler())
	if server.ReadHeaderTimeout != apiReadHeaderTimeout ||
		server.ReadTimeout != apiReadTimeout ||
		server.WriteTimeout != apiWriteTimeout ||
		server.IdleTimeout != apiIdleTimeout ||
		server.MaxHeaderBytes != apiMaxHeaderBytes {
		t.Fatalf(
			"server transport bounds = headers %s read %s write %s idle %s bytes %d",
			server.ReadHeaderTimeout,
			server.ReadTimeout,
			server.WriteTimeout,
			server.IdleTimeout,
			server.MaxHeaderBytes,
		)
	}
	if server.WriteTimeout <= 40*time.Second {
		t.Fatalf("server write timeout = %s, must exceed the authentication handler budget", server.WriteTimeout)
	}
}
