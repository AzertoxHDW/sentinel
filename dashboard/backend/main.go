package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AzertoxHDW/sentinel/dashboard/backend/api"
	"github.com/AzertoxHDW/sentinel/dashboard/backend/collector"
	"github.com/AzertoxHDW/sentinel/dashboard/backend/storage"
	"github.com/AzertoxHDW/sentinel/dashboard/backend/alerts"
)

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	dataFile := flag.String("data", "agents.json", "Agent storage file")
	collectInterval := flag.Duration("interval", 30*time.Second, "Metrics collection interval")
	webhookURL := flag.String("alert-webhook", "", "Discord/Slack Webhook URL for alerts")
    flag.Parse()
	
	// InfluxDB config
	influxURL := flag.String("influx-url", "http://localhost:8086", "InfluxDB URL")
	influxToken := flag.String("influx-token", "", "InfluxDB token")
	influxOrg := flag.String("influx-org", "sentinel", "InfluxDB organization")
	influxBucket := flag.String("influx-bucket", "metrics", "InfluxDB bucket")
	
	flag.Parse()

	finalInfluxToken := *influxToken // Start with the flag value (if set in command)
	
	// Explicitly check the environment variable INFLUX_TOKEN
	if envToken, exists := os.LookupEnv("INFLUX_TOKEN"); exists && envToken != "" {
		// Use the environment variable, overriding the flag if both are set
		finalInfluxToken = envToken 
	}
	
	if finalInfluxToken == "" {
		log.Fatal("InfluxDB token is required. Use -influx-token flag or set INFLUX_TOKEN env var")
	}

	// Initialize storage
	store, err := storage.NewStore(*dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize InfluxDB
	influxDB := storage.NewInfluxDB(storage.InfluxConfig{
		URL:    *influxURL,
		Token:  finalInfluxToken,
		Org:    *influxOrg,
		Bucket: *influxBucket,
	})
	defer influxDB.Close()

	// Initialize Alerts
    alerter := alerts.NewAlerter(*webhookURL)

	// Start metrics collector
	metricsCollector := collector.NewMetricsCollector(store, influxDB, *collectInterval, alerter)
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