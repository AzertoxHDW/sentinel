package discovery

import (
	"context"
	"log"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	ServiceType = "_sentinel._tcp"
	Domain      = "local."
)

type DiscoveredAgent struct {
	Hostname string   `json:"hostname"`
	Instance string   `json:"instance"`
	Port     int      `json:"port"`
	IPs      []string `json:"ips"`
}

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan(timeout time.Duration) ([]*DiscoveredAgent, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	agents := make([]*DiscoveredAgent, 0)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start browsing in background
	go func() {
		if err := resolver.Browse(ctx, ServiceType, Domain, entries); err != nil {
			log.Printf("Browse error: %v", err)
		}
	}()

	// Collect entries until context is done
	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				// Channel closed, we're done
				return agents, nil
			}

			agent := &DiscoveredAgent{
				Hostname: entry.HostName,
				Instance: entry.Instance,
				Port:     entry.Port,
				IPs:      make([]string, 0),
			}

			// Collect IPv4 addresses
			for _, ip := range entry.AddrIPv4 {
				agent.IPs = append(agent.IPs, ip.String())
			}

			if len(agent.IPs) > 0 {
				agents = append(agents, agent)
				log.Printf("Discovered agent: %s at %v:%d", agent.Instance, agent.IPs, agent.Port)
			}

		case <-ctx.Done():
			// Timeout reached, return what we have
			return agents, nil
		}
	}
}