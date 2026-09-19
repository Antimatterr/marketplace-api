package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Antimatterr/marketplace-api/internal/config"
	"github.com/Antimatterr/marketplace-api/internal/db"
	"github.com/Antimatterr/marketplace-api/internal/handlers"
	"github.com/Antimatterr/marketplace-api/internal/middleware"
)

func main() {

	cfg := config.MustLoad()
	db, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("main.db.connect: %v", err)
	}

	//create new logger for handling JSON and set as default
	logHandler := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo}))
	slog.SetDefault(logHandler)

	//initialize the constructor
	listingHandler := handlers.NewListingHandler(db, logHandler)

	// Create a new router instead of using the default servemux to avoid polluting the global state
	// and to have better control over registered routes and handlers
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /listings", listingHandler.List)
	mux.HandleFunc("DELETE /listings/{id}", listingHandler.Delete)
	mux.HandleFunc("POST /listings", listingHandler.Create)

	//wrap the mux handler with the requestId middleware
	muxHandler := middleware.RequestId(mux)

	// Initialize HTTP server with timeouts to prevent resource exhaustion and hanging connections
	srv := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      muxHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrs := make(chan error, 1)
	go func() {
		log.Printf("Server is listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrs <- err
		}
	}()

	select {
	case err := <-serverErrs:
		log.Fatalf("server failed: %v", err)
	case <-ctx.Done():
		log.Println("shutdown signal received, draining connections...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	err = db.Close()
	if err != nil {
		log.Printf("db close failed: %v", err)
	}

	log.Println("server stopped cleanly")

}
