package collector

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMetrics holds all collected system information
type SystemMetrics struct {
	Timestamp   time.Time       `json:"timestamp"`
	Hostname    string          `json:"hostname"`
	Uptime      uint64          `json:"uptime"`
	CPU         CPUMetrics      `json:"cpu"`
	Memory      MemoryMetrics   `json:"memory"`
	Disk        []DiskMetrics   `json:"disk"`
	Network     []NetworkMetrics `json:"network"`
}

type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	CoreCount    int     `json:"core_count"`
	LoadAvg      []float64 `json:"load_avg,omitempty"` // Linux/Unix only
}

type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskMetrics struct {
	Device      string  `json:"device"`
	MountPoint  string  `json:"mount_point"`
	FSType      string  `json:"fs_type"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkMetrics struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// Collector handles metrics collection
type Collector struct {
	hostname string
}

// NewCollector creates a new metrics collector
func NewCollector() (*Collector, error) {
	hostname, err := host.Info()
	if err != nil {
		return nil, err
	}
	
	return &Collector{
		hostname: hostname.Hostname,
	}, nil
}

// Collect gathers all system metrics
func (c *Collector) Collect() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
		Hostname:  c.hostname,
	}

	// Host info (uptime)
	hostInfo, err := host.Info()
	if err == nil {
		metrics.Uptime = hostInfo.Uptime
	}

	// CPU metrics
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPU.UsagePercent = cpuPercent[0]
	}
	metrics.CPU.CoreCount = runtime.NumCPU()

	// Load average (Unix-like systems)
	if runtime.GOOS != "windows" {
		loadAvg, err := cpu.LoadAvg()
		if err == nil {
			metrics.CPU.LoadAvg = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
		}
	}

	// Memory metrics
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.Memory = MemoryMetrics{
			Total:       memInfo.Total,
			Available:   memInfo.Available,
			Used:        memInfo.Used,
			UsedPercent: memInfo.UsedPercent,
		}
	}

	// Disk metrics
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err != nil {
				continue
			}

			metrics.Disk = append(metrics.Disk, DiskMetrics{
				Device:      partition.Device,
				MountPoint:  partition.Mountpoint,
				FSType:      partition.Fstype,
				Total:       usage.Total,
				Used:        usage.Used,
				Free:        usage.Free,
				UsedPercent: usage.UsedPercent,
			})
		}
	}

	// Network metrics
	netIO, err := net.IOCounters(true)
	if err == nil {
		for _, io := range netIO {
			// Skip loopback
			if io.Name == "lo" || io.Name == "lo0" {
				continue
			}

			metrics.Network = append(metrics.Network, NetworkMetrics{
				Interface:   io.Name,
				BytesSent:   io.BytesSent,
				BytesRecv:   io.BytesRecv,
				PacketsSent: io.PacketsSent,
				PacketsRecv: io.PacketsRecv,
			})
		}
	}

	return metrics, nil
}