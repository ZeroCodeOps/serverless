package main

import (
	"fmt"
	"log"
	"net/http"

	"main/config"
	"main/db"
	"main/handlers"
	"main/middleware"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig()

	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create handlers
	h := handlers.NewHandlers(cfg, db.DB)

	// Create a new mux
	mux := http.NewServeMux()

	// Register routes
	h.RegisterRoutes(mux)

	// Wrap the mux with middleware
	handler := middleware.CORS(middleware.Logging(mux))

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	fmt.Printf("Server starting on port %s...\n", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(addr, handler))
}
