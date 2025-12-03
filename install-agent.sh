#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Sentinel Agent Installer${NC}"
echo "=========================="

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Determine binary name
if [ "$OS" = "linux" ]; then
    BINARY="sentinel-agent-linux-$ARCH"
    INSTALL_PATH="/usr/local/bin/sentinel-agent"
elif [ "$OS" = "darwin" ]; then
    BINARY="sentinel-agent-darwin-$ARCH"
    INSTALL_PATH="/usr/local/bin/sentinel-agent"
elif [ "$OS" = "freebsd" ]; then
    BINARY="sentinel-agent-freebsd-$ARCH"
    INSTALL_PATH="/usr/local/bin/sentinel-agent"
else
    echo -e "${RED}Unsupported OS: $OS${NC}"
    exit 1
fi

# Get latest release version
LATEST_VERSION=$(curl -s https://api.github.com/repos/AzertoxHDW/sentinel/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Failed to get latest version${NC}"
    exit 1
fi

echo -e "Latest version: ${GREEN}$LATEST_VERSION${NC}"
echo -e "Installing for: ${GREEN}$OS-$ARCH${NC}"

# Download binary
DOWNLOAD_URL="https://github.com/AzertoxHDW/sentinel/releases/download/$LATEST_VERSION/$BINARY"
echo -e "Downloading from: ${YELLOW}$DOWNLOAD_URL${NC}"

curl -L "$DOWNLOAD_URL" -o /tmp/sentinel-agent

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to download agent${NC}"
    exit 1
fi

# Make executable
chmod +x /tmp/sentinel-agent

# Install (requires sudo)
echo -e "${YELLOW}Installing to $INSTALL_PATH (requires sudo)...${NC}"
sudo mv /tmp/sentinel-agent "$INSTALL_PATH"

# Create systemd service (Linux only)
if [ "$OS" = "linux" ] && command -v systemctl &> /dev/null; then
    echo -e "${YELLOW}Creating systemd service...${NC}"
    
    sudo tee /etc/systemd/system/sentinel-agent.service > /dev/null <<EOF
[Unit]
Description=Sentinel Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_PATH
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable sentinel-agent
    sudo systemctl start sentinel-agent
    
    echo -e "${GREEN}✓ Agent installed and started as systemd service${NC}"
    echo -e "Check status: ${YELLOW}sudo systemctl status sentinel-agent${NC}"
    echo -e "View logs: ${YELLOW}sudo journalctl -u sentinel-agent -f${NC}"
else
    echo -e "${GREEN}✓ Agent installed to $INSTALL_PATH${NC}"
    echo -e "Start manually: ${YELLOW}sudo $INSTALL_PATH${NC}"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo "The agent will automatically broadcast via mDNS and appear in your dashboard."