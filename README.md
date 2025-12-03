# Sentinel

<div align="center">

![Sentinel Logo](dashboard/frontend/public/sentinel-icon.svg)

**Modern infrastructure monitoring with zero-config deployment**

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com/r/azertoxhdw/sentinel)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-00ADD8.svg)](https://golang.org/)

[Features](#key-features) • [Quick Start](#quick-start) • [Docker Deployment](#deploy-the-server) • [Agent Installation](#agent-installation) • [Manual Installation](#manual-installation) • [Configuration](#configuration) • [Troubleshooting](#troubleshooting)

</div>

---

## Overview

Sentinel is a lightweight, self-hosted infrastructure monitoring solution designed for simplicity and ease of use. With automatic service discovery via mDNS, beautiful real-time dashboards, and minimal configuration, Sentinel makes monitoring your homelab or small infrastructure effortless.

### Key Features

- 🔍 **Zero-Config Discovery** - Agents automatically broadcast via mDNS
- 📊 **Real-Time Metrics** - CPU, Memory, Disk, and Network monitoring
- 📈 **Historical Charts** - Track performance trends over time
- 🎨 **Minimalist UI** - Clean dashboard with dark theme
- 🐳 **Docker Ready** - Deploy the entire stack with one command
- 🚀 **Lightweight** - Minimal resource footprint on monitored systems
- 🔒 **Self-Hosted** - Your data stays on your infrastructure


### What's included

- **Agents** - Lightweight Go binaries running on monitored systems
- **Dashboard Backend** - REST API, mDNS scanner, metrics collector
- **InfluxDB** - Time-series database for historical data
- **Frontend** - Svelte-based web interface

## Quick Start

### Deploy the server:

**1. Download the [docker compose file](https://github.com/AzertoxHDW/sentinel/blob/master/docker-compose.yml)**

**2. Change the InfluxDB admin token to whatever you want**

**3. Deploy the stack with the Compose file**

That's it! The dashboard, API, and InfluxDB are now running.

## Agent Installation

### Quick Install (Linux/macOS):
```bash
curl -sSL https://raw.githubusercontent.com/AzertoxHDW/sentinel/main/install-agent.sh | sudo bash
```

### Manual Installation:

**Linux (x64):**
```bash
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-linux-amd64
chmod +x sentinel-agent-linux-amd64
sudo mv sentinel-agent-linux-amd64 /usr/local/bin/sentinel-agent
sudo sentinel-agent
```

**Raspberry Pi (ARM64):**
```bash
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-linux-arm64
chmod +x sentinel-agent-linux-arm64
sudo mv sentinel-agent-linux-arm64 /usr/local/bin/sentinel-agent
sudo sentinel-agent
```

**Windows:**
1. Download [sentinel-agent-windows-amd64.exe](https://github.com/AzertoxHDW/sentinel/releases/latest)
2. Run as Administrator
3. Add to Windows Firewall exceptions for port 9100

**macOS:**
```bash
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-darwin-arm64  # Apple Silicon
# or
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-darwin-amd64  # Intel
chmod +x sentinel-agent-darwin-*
sudo mv sentinel-agent-darwin-* /usr/local/bin/sentinel-agent
sudo sentinel-agent
```

### Uninstall
```bash
curl -sSL https://raw.githubusercontent.com/AzertoxHDW/sentinel/main/uninstall-agent.sh | sudo bash
```

Agents will automatically:
- Start collecting system metrics
- Broadcast their presence via mDNS
- Appear in the dashboard's "Discover Agents" scanner

## Manual Installation

### Prerequisites

- Go 1.25.4 or higher
- Node.js 20+
- InfluxDB 2.x
- (Optional) Avahi/mDNS daemon for service discovery

### Build from Source

```bash
# Clone repository
git clone https://github.com/AzertoxHDW/sentinel.git
cd sentinel

# Build agent
go build -o sentinel-agent ./agent

# Build dashboard backend
go build -o sentinel-dashboard ./dashboard/backend

# Build frontend
cd dashboard/frontend
npm install
npm run build
```

### Configure the stack

**InfluxDB Setup:**

```bash
# Start InfluxDB
influxdb

# Access UI at http://localhost:8086
# Create:
# - Organization: sentinel
# - Bucket: metrics
# - Generate token
```

**Start Dashboard:**

```bash
./sentinel-dashboard \
  -influx-url=http://localhost:8086 \
  -influx-token=YOUR_TOKEN \
  -influx-org=sentinel \
  -influx-bucket=metrics
```

**Serve Frontend:**

```bash
cd dashboard/frontend
npm run preview
# Or use any static file server for the dist/ directory
```

## Configuration

### Environment Variables

Dashboard backend supports these environment variables:

```bash
INFLUX_URL=http://localhost:8086
INFLUX_TOKEN=your-influx-token
INFLUX_ORG=sentinel
INFLUX_BUCKET=metrics
```

### Command Line Flags

```bash
./sentinel-dashboard \
  -port=8080 \
  -data=/path/to/agents.json \
  -interval=30s \
  -influx-url=http://localhost:8086 \
  -influx-token=TOKEN \
  -influx-org=sentinel \
  -influx-bucket=metrics
```

### Docker Compose Customization

Edit `docker-compose.yml` to customize:

- Ports
- InfluxDB credentials
- Data retention policies
- Network configuration

**Example: Change ports**

```yaml
services:
  dashboard-frontend:
    ports:
      - "8080:80"  # Change from 3000 to 8080
```
### **ATTENTION:**
**The stack has been optimized to run in Docker containers, so the network and ports configuration is set for this use case out-of-the-box. Change the network configuration at your own risks, and if you don't know exactly what you're doing... just don't.**

## Metrics Collected

### System Metrics

- **CPU**: Usage percentage, core count, model
- **Memory**: Total, used, available, percentage
- **Disk**: Usage per partition (root only by default)
- **Network**: Real-time bandwidth (upload/download) for physical interfaces
- **Uptime**: System uptime in seconds

### Collection Interval

Metrics are collected every 30 seconds by default. Configure with `-interval` flag.

## API Reference

### Endpoints

```
GET  /api/health                    - Health check
GET  /api/agents                    - List all agents
POST /api/agents                    - Register new agent
DELETE /api/agents/{id}             - Remove agent
GET  /api/agents/discover           - Scan network for agents
GET  /api/metrics/{agentID}         - Get current metrics
GET  /api/history/{agentID}/{measurement} - Get historical data
```

### Example: Get Metrics

```bash
curl http://localhost:8080/api/metrics/hostname:9100
```

## Troubleshooting

### Agents Not Discovered

**Check mDNS/Avahi:**
```bash
# Linux
systemctl status avahi-daemon

# Scan for Sentinel services
avahi-browse -t _sentinel._tcp
```

**Firewall Issues:**
- Ensure port 9100 is open agent-side
- Ensure port 8080 is open server-side
- Allow mDNS traffic (UDP 5353)

**Manual Addition:**
If discovery doesn't work, manually add agents via API:
```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{"ip_address":"192.168.1.100","port":9100}'
```

### Dashboard Can't Connect to InfluxDB

**Check InfluxDB Status:**
- Verify the container is running (`docker ps`)
- Check the container logs

**Verify Token:**
- Log into InfluxDB UI at `http://<influxdb-ip>:8086`
- Go to Settings → Tokens
- Ensure token matches docker-compose.yml

## Development

### Project Structure

```
sentinel/
├── agent/                  # Agent source code
│   ├── main.go
│   ├── collector/         # Metrics collection
│   ├── server/            # HTTP server
│   └── discovery/         # mDNS broadcasting
├── dashboard/
│   ├── backend/           # Dashboard API
│   │   ├── main.go
│   │   ├── api/
│   │   ├── storage/
│   │   └── collector/
│   └── frontend/          # Web UI (Svelte)
│       ├── src/
│       └── public/
├── docker-compose.yml
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Acknowledgments

- Built with [Go](https://golang.org/)
- UI powered by [Svelte](https://svelte.dev/) and [TailwindCSS](https://tailwindcss.com/)
- Time-series storage by [InfluxDB](https://www.influxdata.com/)
- Charts by [Chart.js](https://www.chartjs.org/)
- Icons from [Heroicons](https://heroicons.com/)

---

<div align="center">

Made with ❤️ by [AzertoxHDW](https://github.com/AzertoxHDW)

**[⬆ back to top](#sentinel)**

</div>