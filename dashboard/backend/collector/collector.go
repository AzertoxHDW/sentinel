package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AzertoxHDW/sentinel/dashboard/backend/storage"
)

type MetricsCollector struct {
	store      *storage.Store
	influxDB   *storage.InfluxDB
	httpClient *http.Client
	interval   time.Duration
	stopChan   chan struct{}
}

func NewMetricsCollector(store *storage.Store, influxDB *storage.InfluxDB, interval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		store:    store,
		influxDB: influxDB,
		interval: interval,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

func (mc *MetricsCollector) Start() {
	ticker := time.NewTicker(mc.interval)
	go func() {
		// Collect immediately on start
		mc.collectAll()
		
		for {
			select {
			case <-ticker.C:
				mc.collectAll()
			case <-mc.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	log.Printf("Metrics collector started (interval: %v)", mc.interval)
}

func (mc *MetricsCollector) Stop() {
	close(mc.stopChan)
	log.Println("Metrics collector stopped")
}

func (mc *MetricsCollector) collectAll() {
	agents := mc.store.GetAllAgents()
	
	for _, agent := range agents {
		if err := mc.collectAgent(agent); err != nil {
			log.Printf("Failed to collect metrics for %s: %v", agent.ID, err)
		}
	}
}

func (mc *MetricsCollector) collectAgent(agent *storage.Agent) error {
	// Fetch metrics from agent
	metricsURL := fmt.Sprintf("http://%s:%d/metrics", agent.IPAddress, agent.Port)
	resp, err := mc.httpClient.Get(metricsURL)
	if err != nil {
		mc.store.UpdateAgentStatus(agent.ID, "offline")
		return err
	}
	defer resp.Body.Close()

	// Parse metrics from agent
	var agentMetrics AgentMetrics
	if err := json.NewDecoder(resp.Body).Decode(&agentMetrics); err != nil {
		return err
	}

	// Update agent status
	mc.store.UpdateAgentStatus(agent.ID, "online")

	// Convert to storage format
	metrics := convertToStorageMetrics(&agentMetrics)

	// Use the actual hostname from metrics for consistency
	hostname := agentMetrics.Hostname
	
	log.Printf("Writing metrics for agent_id=%s, hostname=%s", agent.ID, hostname)

	// Write to InfluxDB - use agent.ID consistently
	return mc.influxDB.WriteMetrics(agent.ID, metrics)
}

// AgentMetrics matches the structure from the agent's /metrics endpoint
type AgentMetrics struct {
	Hostname string `json:"hostname"`
	CPU      struct {
		UsagePercent float64 `json:"usage_percent"`
		CoreCount    int     `json:"core_count"`
	} `json:"cpu"`
	Memory struct {
		Total       uint64  `json:"total"`
		Used        uint64  `json:"used"`
		Available   uint64  `json:"available"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"memory"`
	Disk []struct {
		Device      string  `json:"device"`
		MountPoint  string  `json:"mount_point"`
		Total       uint64  `json:"total"`
		Used        uint64  `json:"used"`
		Free        uint64  `json:"free"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"disk"`
	Network []struct {
		Interface   string `json:"interface"`
		BytesSent   uint64 `json:"bytes_sent"`
		BytesRecv   uint64 `json:"bytes_recv"`
		PacketsSent uint64 `json:"packets_sent"`
		PacketsRecv uint64 `json:"packets_recv"`
	} `json:"network"`
}

func convertToStorageMetrics(am *AgentMetrics) *storage.SystemMetrics {
	metrics := &storage.SystemMetrics{
		Hostname:     am.Hostname,
		CPUPercent:   am.CPU.UsagePercent,
		CoreCount:    am.CPU.CoreCount,
		MemTotal:     am.Memory.Total,
		MemUsed:      am.Memory.Used,
		MemAvailable: am.Memory.Available,
		MemPercent:   am.Memory.UsedPercent,
		Disks:        make([]storage.DiskMetric, 0),
		Networks:     make([]storage.NetworkMetric, 0),
	}

	for _, disk := range am.Disk {
		metrics.Disks = append(metrics.Disks, storage.DiskMetric{
			Device:      disk.Device,
			MountPoint:  disk.MountPoint,
			Total:       disk.Total,
			Used:        disk.Used,
			Free:        disk.Free,
			UsedPercent: disk.UsedPercent,
		})
	}

	for _, net := range am.Network {
		metrics.Networks = append(metrics.Networks, storage.NetworkMetric{
			Interface:   net.Interface,
			BytesSent:   net.BytesSent,
			BytesRecv:   net.BytesRecv,
			PacketsSent: net.PacketsSent,
			PacketsRecv: net.PacketsRecv,
		})
	}

	return metrics
}