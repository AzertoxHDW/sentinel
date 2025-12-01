package discovery

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	ServiceType = "_sentinel._tcp"
	Domain      = "local."
)

type Broadcaster struct {
	server *zeroconf.Server
	port   int
}

func NewBroadcaster(port int) *Broadcaster {
	return &Broadcaster{
		port: port,
	}
}

func (b *Broadcaster) Start() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "sentinel-agent"
	}

	// Register service
	server, err := zeroconf.Register(
		hostname,           // Instance name
		ServiceType,        // Service type
		Domain,            // Domain
		b.port,            // Port
		[]string{"txtv=0", "version=1.0"}, // TXT records
		nil,               // Network interface (nil = all)
	)
	
	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}

	b.server = server
	log.Printf("mDNS service registered: %s%s on port %d", hostname, ServiceType, b.port)
	
	return nil
}

func (b *Broadcaster) Stop() {
	if b.server != nil {
		b.server.Shutdown()
		log.Println("mDNS service stopped")
	}
}

// Scanner discovers Sentinel agents on the network
type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan(timeout time.Duration) ([]*AgentInfo, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	agents := []*AgentInfo{}

	go func() {
		for entry := range entries {
			agent := &AgentInfo{
				Hostname: entry.HostName,
				Instance: entry.Instance,
				Port:     entry.Port,
				IPs:      make([]string, len(entry.AddrIPv4)),
			}
			
			for i, ip := range entry.AddrIPv4 {
				agent.IPs[i] = ip.String()
			}
			
			agents = append(agents, agent)
			log.Printf("Discovered agent: %s at %v:%d", agent.Instance, agent.IPs, agent.Port)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = resolver.Browse(ctx, ServiceType, Domain, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to browse: %w", err)
	}

	<-ctx.Done()
	
	return agents, nil
}

type AgentInfo struct {
	Hostname string
	Instance string
	Port     int
	IPs      []string
}