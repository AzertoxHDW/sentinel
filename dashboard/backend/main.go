package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sentinel/dashboard/backend/api"
	"sentinel/dashboard/backend/storage"
)

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	dataFile := flag.String("data", "agents.json", "Agent storage file")
	flag.Parse()

	log.Println("Starting Sentinel Dashboard...")

	// Initialize storage
	store, err := storage.NewStore(*dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create API server
	server := api.NewServer(store, *port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down dashboard...")
		os.Exit(0)
	}()

	// Start server (blocking)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}