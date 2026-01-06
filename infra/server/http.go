package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/alifpay/temporal/app"
)

// Start starts an HTTP server on the given address (or :8080 if empty).
func Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", app.PayHandler)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
	}()

	log.Printf("HTTP server listening on %s", addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
