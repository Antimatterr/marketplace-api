package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Antimatterr/marketplace-api/internal/config"
	"github.com/Antimatterr/marketplace-api/internal/handlers"
)

func main() {

	cfg := config.MustLoad()

	// Create a new router instead of using the default servemux to avoid polluting the global state
	// and to have better control over registered routes and handlers
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)

	// Initialize HTTP server with timeouts to prevent resource exhaustion and hanging connections
	srv := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server is listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed : %v", err)
	}
}
