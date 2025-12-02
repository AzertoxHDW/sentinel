package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/AzertoxHDW/sentinel/dashboard/backend/discovery"
	"github.com/AzertoxHDW/sentinel/dashboard/backend/storage"
)

type Server struct {
	store      *storage.Store
	scanner    *discovery.Scanner
	influxDB   *storage.InfluxDB
	port       string
	httpClient *http.Client
}

func NewServer(store *storage.Store, influxDB *storage.InfluxDB, port string) *Server {
	return &Server{
		store:    store,
		scanner:  discovery.NewScanner(),
		influxDB: influxDB,
		port:     port,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	// CORS middleware
	mux := http.NewServeMux()

	// Agent management endpoints
	mux.HandleFunc("/api/agents", s.handleAgents)
	mux.HandleFunc("/api/agents/", s.handleAgent)
	mux.HandleFunc("/api/agents/discover", s.handleDiscover)

	// Metrics proxy endpoint
	mux.HandleFunc("/api/metrics/", s.handleMetrics)

	// History endpoint
	mux.HandleFunc("/api/history/", s.handleHistory)

	// Health check
	mux.HandleFunc("/api/health", s.handleHealth)

	handler := corsMiddleware(mux)

	log.Printf("Dashboard API starting on :%s", s.port)
	return http.ListenAndServe(":"+s.port, handler)
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GET /api/agents - List all agents
// POST /api/agents - Add new agent
func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		agents := s.store.GetAllAgents()
		s.respondJSON(w, http.StatusOK, agents)

	case http.MethodPost:
		// Decode only IP and port from request
		var req struct {
			IPAddress string `json:"ip_address"`
			Port      int    `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Fetch actual hostname from the agent
		metricsURL := fmt.Sprintf("http://%s:%d/metrics", req.IPAddress, req.Port)
		resp, err := s.httpClient.Get(metricsURL)
		if err != nil {
			log.Printf("Failed to reach agent at %s: %v", metricsURL, err)
			s.respondError(w, http.StatusServiceUnavailable, "Cannot reach agent")
			return
		}
		defer resp.Body.Close()

		// Parse metrics to get hostname
		var metricsResp struct {
			Hostname string `json:"hostname"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&metricsResp); err != nil {
			log.Printf("Failed to parse agent metrics: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Invalid agent response")
			return
		}

		// Create agent with real hostname
		agent := &storage.Agent{
			Hostname:  metricsResp.Hostname,
			IPAddress: req.IPAddress,
			Port:      req.Port,
			ID:        fmt.Sprintf("%s:%d", metricsResp.Hostname, req.Port),
		}

		log.Printf("Adding agent: ID=%s, Hostname=%s, IP=%s:%d",
			agent.ID, agent.Hostname, agent.IPAddress, agent.Port)

		if err := s.store.AddAgent(agent); err != nil {
			s.respondError(w, http.StatusInternalServerError, "Failed to add agent")
			return
		}

		s.respondJSON(w, http.StatusCreated, agent)

	default:
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// DELETE /api/agents/{id} - Remove agent
func (s *Server) handleAgent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	id := r.URL.Path[len("/api/agents/"):]

	if id == "" {
		s.respondError(w, http.StatusBadRequest, "Agent ID required")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := s.store.RemoveAgent(id); err != nil {
			s.respondError(w, http.StatusInternalServerError, "Failed to remove agent")
			return
		}
		s.respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// GET /api/agents/discover - Discover agents via mDNS
func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	log.Println("Scanning network for Sentinel agents...")

	discovered, err := s.scanner.Scan(3 * time.Second)
	if err != nil {
		log.Printf("Discovery error: %v", err)
		s.respondError(w, http.StatusInternalServerError, "Discovery failed")
		return
	}

	log.Printf("Found %d agents via mDNS", len(discovered))

	// Get already registered agents
	registeredAgents := s.store.GetAllAgents()
	
	// Filter out already registered agents
	var newAgents []*discovery.DiscoveredAgent
	for _, disc := range discovered {
		isRegistered := false
		
		// Check against registered agents
		for _, registered := range registeredAgents {
			// 1. Check if Hostname matches (Primary ID check)
			// 'disc.Instance' usually contains the pure hostname (e.g. "my-server")
			// 'registered.Hostname' also contains the hostname.
			if registered.Hostname == disc.Instance {
				isRegistered = true
				break
			}

			// 2. Check if any IP matches (Fallback)
			for _, ip := range disc.IPs {
				if registered.IPAddress == ip && registered.Port == disc.Port {
					isRegistered = true
					break
				}
			}
			if isRegistered {
				break
			}
		}
		
		if !isRegistered {
			newAgents = append(newAgents, disc)
		}
	}

	log.Printf("Returning %d new agents (filtered %d already registered)", len(newAgents), len(discovered)-len(newAgents))
	s.respondJSON(w, http.StatusOK, newAgents)
}

// GET /api/metrics/{agentID} - Proxy metrics from agent
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract agent ID from path
	agentID := r.URL.Path[len("/api/metrics/"):]

	if agentID == "" {
		s.respondError(w, http.StatusBadRequest, "Agent ID required")
		return
	}

	agent, exists := s.store.GetAgent(agentID)
	if !exists {
		s.respondError(w, http.StatusNotFound, "Agent not found")
		return
	}

	// Fetch metrics from agent
	metricsURL := fmt.Sprintf("http://%s:%d/metrics", agent.IPAddress, agent.Port)
	resp, err := s.httpClient.Get(metricsURL)
	if err != nil {
		log.Printf("Failed to fetch metrics from %s: %v", agentID, err)
		s.store.UpdateAgentStatus(agentID, "offline")
		s.respondError(w, http.StatusServiceUnavailable, "Agent unreachable")
		return
	}
	defer resp.Body.Close()

	// Update agent status
	s.store.UpdateAgentStatus(agentID, "online")

	// Proxy response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// GET /api/history/{agentID}/{measurement}?duration=1h - Get historical metrics
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse URL: /api/history/{agentID}/{measurement}?duration=1h
	path := r.URL.Path[len("/api/history/"):]
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		s.respondError(w, http.StatusBadRequest, "Invalid path. Use /api/history/{agentID}/{measurement}")
		return
	}

	agentID := parts[0]
	measurement := parts[1]

	// Parse duration from query params (default 1 hour)
	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid duration format")
		return
	}

	// Query InfluxDB
	records, err := s.influxDB.QueryMetrics(agentID, measurement, duration)
	if err != nil {
		log.Printf("Failed to query metrics: %v", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to query metrics")
		return
	}

	s.respondJSON(w, http.StatusOK, records)
}

// Health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
	})
}

// Helper: Respond with JSON
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper: Respond with error
func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, map[string]string{"error": message})
}