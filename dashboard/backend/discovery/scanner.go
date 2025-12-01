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
	Hostname string
	Instance string
	Port     int
	IPs      []string
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

	entries := make(chan *zeroconf.ServiceEntry, 10) // Buffered channel
	agents := []*DiscoveredAgent{}
	done := make(chan bool)

	go func() {
		for entry := range entries {
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
		}
		done <- true
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := resolver.Browse(ctx, ServiceType, Domain, entries); err != nil {
		close(entries)
		return nil, err
	}

	<-ctx.Done()
	close(entries) // Close before waiting
	<-done         // Wait for goroutine to finish

	return agents, nil
}