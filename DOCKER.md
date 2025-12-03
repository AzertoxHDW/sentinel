# Sentinel Docker Deployment

Deploy Sentinel monitoring stack with Docker Compose.

## Quick Start

1. **Clone the repository:**
```bash
   git clone https://github.com/AzertoxHDW/sentinel.git
   cd sentinel
```

2. **(Optional) Change InfluxDB token:**
   Edit `docker-compose.yml` and change `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` to a secure value.

3. **Start the stack:**
```bash
   docker-compose up -d
```

4. **Access the dashboard:**
   - Frontend: http://localhost:3000
   - InfluxDB UI: http://localhost:8086
   - API: http://localhost:8080

## Deploy Agents

Agents run on the machines you want to monitor. They are **NOT** containerized for better system access.

### Binary Installation
```bash
# Download latest release
wget https://github.com/AzertoxHDW/sentinel/releases/latest/download/sentinel-agent-linux-amd64

# Make executable
chmod +x sentinel-agent-linux-amd64

# Run
./sentinel-agent-linux-amd64
```

### Build from source
```bash
git clone https://github.com/AzertoxHDW/sentinel.git
cd sentinel
go build -o sentinel-agent ./agent
./sentinel-agent
```

The agent will:
- Collect system metrics
- Broadcast via mDNS
- Appear in dashboard's "Discover Agents"

## Configuration

### InfluxDB Credentials

Default credentials (change in production!):
- Username: `sentinel`
- Password: `sentinelpassword`
- Org: `sentinel`
- Bucket: `metrics`
- Token: `sentinel-super-secret-token-change-me-in-production`

### Ports

- `3000` - Frontend (Web UI)
- `8080` - Backend API
- `8086` - InfluxDB
- `9100` - Agent (on monitored machines)

## Stopping
```bash
docker-compose down
```

To remove data volumes:
```bash
docker-compose down -v
```

## Updating
```bash
git pull
docker-compose down
docker-compose up -d --build
```

## Troubleshooting

**Agents not appearing:**
- Ensure mDNS/Avahi is working on your network
- Check firewall rules for port 9100
- Manually add agent IP in dashboard

**Can't connect to InfluxDB:**
- Wait 30s for InfluxDB to initialize
- Check logs: `docker-compose logs influxdb`

**Dashboard not loading:**
- Check logs: `docker-compose logs dashboard-backend dashboard-frontend`
- Verify all services are running: `docker-compose ps`