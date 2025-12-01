package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AzertoxHDW/sentinel/agent/server"
)

func main() {
	port := flag.String("port", "9100", "Port to listen on")
	flag.Parse()

	log.Println("Starting Sentinel Agent...")
	log.Printf("Hostname: %s", getHostname())

	// Create and start HTTP server
	srv, err := server.NewServer(*port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down agent...")
		os.Exit(0)
	}()

	// Start server (blocking)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}