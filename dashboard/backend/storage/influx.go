package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxDB struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	queryAPI api.QueryAPI
	bucket   string
	org      string
}

type InfluxConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

func NewInfluxDB(config InfluxConfig) *InfluxDB {
	client := influxdb2.NewClient(config.URL, config.Token)
	
	return &InfluxDB{
		client:   client,
		writeAPI: client.WriteAPI(config.Org, config.Bucket),
		queryAPI: client.QueryAPI(config.Org),
		bucket:   config.Bucket,
		org:      config.Org,
	}
}

// WriteMetrics writes system metrics to InfluxDB
func (db *InfluxDB) WriteMetrics(agentID string, metrics *SystemMetrics) error {
	timestamp := time.Now()

	hostname := metrics.Hostname

	// CPU metrics
	cpuPoint := influxdb2.NewPoint(
		"cpu",
		map[string]string{
			"agent_id": agentID,
			"hostname": hostname,
		},
		map[string]interface{}{
			"usage_percent": metrics.CPUPercent,
			"core_count":    metrics.CoreCount,
		},
		timestamp,
	)
	db.writeAPI.WritePoint(cpuPoint)

	// Memory metrics
	memPoint := influxdb2.NewPoint(
		"memory",
		map[string]string{
			"agent_id": agentID,
			"hostname": hostname,
		},
		map[string]interface{}{
			"total":        metrics.MemTotal,
			"used":         metrics.MemUsed,
			"available":    metrics.MemAvailable,
			"used_percent": metrics.MemPercent,
		},
		timestamp,
	)
	db.writeAPI.WritePoint(memPoint)

	// Disk metrics
	for _, disk := range metrics.Disks {
		diskPoint := influxdb2.NewPoint(
			"disk",
			map[string]string{
				"agent_id":    agentID,
				"hostname":    hostname,
				"mount_point": disk.MountPoint,
				"device":      disk.Device,
			},
			map[string]interface{}{
				"total":        disk.Total,
				"used":         disk.Used,
				"free":         disk.Free,
				"used_percent": disk.UsedPercent,
			},
			timestamp,
		)
		db.writeAPI.WritePoint(diskPoint)
	}

	// Network metrics
	for _, net := range metrics.Networks {
		netPoint := influxdb2.NewPoint(
			"network",
			map[string]string{
				"agent_id":  agentID,
				"hostname":  hostname,
				"interface": net.Interface,
			},
			map[string]interface{}{
				"bytes_sent":   net.BytesSent,
				"bytes_recv":   net.BytesRecv,
				"packets_sent": net.PacketsSent,
				"packets_recv": net.PacketsRecv,
			},
			timestamp,
		)
		db.writeAPI.WritePoint(netPoint)
	}

	// Flush writes
	db.writeAPI.Flush()

	return nil
}

// QueryMetrics retrieves historical metrics
func (db *InfluxDB) QueryMetrics(agentID string, measurement string, duration time.Duration) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -%s)
		|> filter(fn: (r) => r["_measurement"] == "%s")
		|> filter(fn: (r) => r["agent_id"] == "%s")
		|> aggregateWindow(every: 30s, fn: mean, createEmpty: false)
		|> yield(name: "mean")
	`, db.bucket, duration.String(), measurement, agentID)

	result, err := db.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next() {
		record := make(map[string]interface{})
		record["time"] = result.Record().Time()
		record["value"] = result.Record().Value()
		
		// Add all fields and tags
		for k, v := range result.Record().Values() {
			record[k] = v
		}
		
		records = append(records, record)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return records, nil
}

// Close closes the InfluxDB client
func (db *InfluxDB) Close() {
	db.client.Close()
	log.Println("InfluxDB client closed")
}

// SystemMetrics represents the metrics structure from agents
type SystemMetrics struct {
	Hostname     string
	CPUPercent   float64
	CoreCount    int
	MemTotal     uint64
	MemUsed      uint64
	MemAvailable uint64
	MemPercent   float64
	Disks        []DiskMetric
	Networks     []NetworkMetric
}

type DiskMetric struct {
	Device      string
	MountPoint  string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

type NetworkMetric struct {
	Interface   string
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
}