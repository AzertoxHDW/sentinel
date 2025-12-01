package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"sentinel/agent/collector"
)

type Server struct {
	collector *collector.Collector
	port      string
}

func NewServer(port string) (*Server, error) {
	col, err := collector.NewCollector()
	if err != nil {
		return nil, err
	}

	return &Server{
		collector: col,
		port:      port,
	}, nil
}

func (s *Server) Start() error {
	http.HandleFunc("/metrics", s.handleMetrics)
	http.HandleFunc("/health", s.handleHealth)

	log.Printf("Agent server starting on :%s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// CORS headers for dashboard access
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := s.collector.Collect()
	if err != nil {
		log.Printf("Error collecting metrics: %v", err)
		http.Error(w, "Failed to collect metrics", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("Error encoding metrics: %v", err)
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(health)
}