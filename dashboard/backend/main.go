package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sentinel/dashboard/backend/api"
	"sentinel/dashboard/backend/collector"
	"sentinel/dashboard/backend/storage"
)

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	dataFile := flag.String("data", "agents.json", "Agent storage file")
	collectInterval := flag.Duration("interval", 30*time.Second, "Metrics collection interval")
	
	// InfluxDB config
	influxURL := flag.String("influx-url", "http://localhost:8086", "InfluxDB URL")
	influxToken := flag.String("influx-token", "", "InfluxDB token")
	influxOrg := flag.String("influx-org", "sentinel", "InfluxDB organization")
	influxBucket := flag.String("influx-bucket", "metrics", "InfluxDB bucket")
	
	flag.Parse()

	if *influxToken == "" {
		log.Fatal("InfluxDB token is required. Use -influx-token flag or set INFLUX_TOKEN env var")
	}

	log.Println("Starting Sentinel Dashboard...")

	// Initialize storage
	store, err := storage.NewStore(*dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize InfluxDB
	influxDB := storage.NewInfluxDB(storage.InfluxConfig{
		URL:    *influxURL,
		Token:  *influxToken,
		Org:    *influxOrg,
		Bucket: *influxBucket,
	})
	defer influxDB.Close()

	// Start metrics collector
	metricsCollector := collector.NewMetricsCollector(store, influxDB, *collectInterval)
	metricsCollector.Start()
	defer metricsCollector.Stop()

	// Create API server
	server := api.NewServer(store, influxDB, *port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down dashboard...")
		metricsCollector.Stop()
		influxDB.Close()
		os.Exit(0)
	}()

	// Start server (blocking)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}