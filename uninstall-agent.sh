#!/bin/bash
set -e

echo "Uninstalling Sentinel Agent..."

# Stop and disable service (Linux)
if command -v systemctl &> /dev/null && systemctl list-unit-files | grep -q sentinel-agent; then
    echo "Stopping service..."
    sudo systemctl stop sentinel-agent
    sudo systemctl disable sentinel-agent
    sudo rm /etc/systemd/system/sentinel-agent.service
    sudo systemctl daemon-reload
fi

# Remove binary
if [ -f /usr/local/bin/sentinel-agent ]; then
    echo "Removing binary..."
    sudo rm /usr/local/bin/sentinel-agent
fi

echo "Sentinel Agent uninstalled successfully!"