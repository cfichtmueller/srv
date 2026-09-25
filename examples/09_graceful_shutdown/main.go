package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cfichtmueller/srv"
)

func main() {
	// Cancelled on SIGINT/SIGTERM to start graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := srv.NewServer().
		Use(srv.LoggingMiddleware()).
		SetReadHeaderTimeout(10 * time.Second).
		SetReadTimeout(30 * time.Second).
		SetWriteTimeout(30 * time.Second).
		SetIdleTimeout(60 * time.Second)

	s.GET("/healthz", func(c *srv.Context) *srv.Response {
		return srv.Respond().Json(map[string]any{"status": "ok"})
	})
	// A slow endpoint to demonstrate in-flight requests draining on shutdown.
	s.GET("/slow", func(c *srv.Context) *srv.Response {
		time.Sleep(5 * time.Second)
		return srv.Respond().Json(map[string]any{"status": "done"})
	})

	errCh := make(chan error, 1)
	go func() {
		log.Println("listening on :8080")
		if err := s.ListenAndServe(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	case <-ctx.Done():
		log.Println("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("stopped")
}
