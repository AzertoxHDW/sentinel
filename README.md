# Sentinel

<div align="center">

![Sentinel Logo](dashboard/frontend/public/sentinel-icon.svg)

**Modern infrastructure monitoring with zero-config deployment**

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com/r/azertoxhdw/sentinel)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)

[Features](#features) • [Quick Start](#quick-start) • [Docker Deployment](#docker-deployment) • [Manual Installation](#manual-installation) • [Documentation](#documentation)

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

## Architecture

```
┌─────────────┐     mDNS       ┌──────────────┐
│   Agent 1   │◄──────────────►│              │
└─────────────┘                │              │
                               │  Dashboard   │──► InfluxDB
┌─────────────┐     HTTP       │   Backend    │
│   Agent 2   │◄──────────────►│              │
└─────────────┘                │              │
                               └──────────────┘
┌─────────────┐                       │
│   Agent N   │                       ▼
└─────────────┘                ┌──────────────┐
                               │   Frontend   │
                               │   (Web UI)   │
                               └──────────────┘
```

- **Agents** - Lightweight Go binaries running on monitored systems
- **Dashboard Backend** - REST API, mDNS scanner, metrics collector
- **InfluxDB** - Time-series database for historical data
- **Frontend** - Svelte-based web interface

## Quick Start

### Docker Deployment (Recommended)

The fastest way to get Sentinel running:

```bash
# Clone the repository
git clone https://github.com/AzertoxHDW/sentinel.git
cd sentinel

# Start the stack
docker-compose up -d

# Access the dashboard
open http://localhost:3000
```

That's it! The dashboard, API, and InfluxDB are now running.

### Deploy Agents

Agents run on the systems you want to monitor. Download the latest release for your platform:

**Linux (x64):**
```bash
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-linux-amd64
chmod +x sentinel-agent-linux-amd64
./sentinel-agent-linux-amd64
```

**Linux (ARM64 - Raspberry Pi):**
```bash
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-linux-arm64
chmod +x sentinel-agent-linux-arm64
./sentinel-agent-linux-arm64
```

**Windows:**
```powershell
# Download from releases page
.\sentinel-agent-windows-amd64.exe
```

**macOS:**
```bash
# Intel
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-darwin-amd64
chmod +x sentinel-agent-darwin-amd64
./sentinel-agent-darwin-amd64

# Apple Silicon
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-darwin-arm64
chmod +x sentinel-agent-darwin-arm64
./sentinel-agent-darwin-arm64
```

Agents will automatically:
- Start collecting system metrics
- Broadcast their presence via mDNS
- Appear in the dashboard's "Discover Agents" scanner

## Manual Installation

### Prerequisites

- Go 1.21 or higher
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

### Configuration

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
- Ensure port 9100 (agent) is open
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
```bash
docker-compose logs influxdb
```

**Verify Token:**
- Log into InfluxDB UI at http://localhost:8086
- Go to Settings → Tokens
- Ensure token matches docker-compose.yml

### High Memory Usage

Agents use ~10-20MB RAM. If higher:
- Check for memory leaks
- Reduce collection interval
- Update to latest version

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

### Building

```bash
# Agent
go build -o sentinel-agent ./agent

# Dashboard
go build -o sentinel-dashboard ./dashboard/backend

# Frontend
cd dashboard/frontend
npm run build
```

### Running Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Go](https://golang.org/)
- UI powered by [Svelte](https://svelte.dev/) and [Tailwind CSS](https://tailwindcss.com/)
- Time-series storage by [InfluxDB](https://www.influxdata.com/)
- Charts by [Chart.js](https://www.chartjs.org/)
- Icons from [Heroicons](https://heroicons.com/)

---

<div align="center">

Made with ❤️ by [AzertoxHDW](https://github.com/AzertoxHDW)

**[⬆ back to top](#sentinel)**

</div>