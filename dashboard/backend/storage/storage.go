package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Agent struct {
	ID          string    `json:"id"`
	Hostname    string    `json:"hostname"`
	IPAddress   string    `json:"ip_address"`
	Port        int       `json:"port"`
	AddedAt     time.Time `json:"added_at"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"` // online, offline, unknown
}

type Store struct {
	agents map[string]*Agent
	mu     sync.RWMutex
	file   string
}

func NewStore(file string) (*Store, error) {
	s := &Store{
		agents: make(map[string]*Agent),
		file:   file,
	}

	// Load existing agents from file
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load agents: %w", err)
	}

	return s, nil
}

func (s *Store) AddAgent(agent *Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent.AddedAt = time.Now()
	agent.LastSeen = time.Now()
	agent.Status = "online"
	
	s.agents[agent.ID] = agent

	return s.save()
}

func (s *Store) GetAgent(id string) (*Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.agents[id]
	return agent, exists
}

func (s *Store) GetAllAgents() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents
}

func (s *Store) RemoveAgent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.agents, id)
	return s.save()
}

func (s *Store) UpdateAgentStatus(id string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agent, exists := s.agents[id]; exists {
		agent.Status = status
		agent.LastSeen = time.Now()
		return s.save()
	}

	return fmt.Errorf("agent not found: %s", id)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return err
	}

	var agents []*Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		return err
	}

	for _, agent := range agents {
		s.agents[agent.ID] = agent
	}

	return nil
}

func (s *Store) save() error {
	agents := make([]*Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}