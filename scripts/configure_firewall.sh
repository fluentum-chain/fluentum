#!/bin/bash

# Fluentum Testnet Firewall Configuration Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

print_status "Configuring firewall for Fluentum testnet..."

# Check if ufw is available
if ! command -v ufw &> /dev/null; then
    print_error "ufw is not installed. Please install it first:"
    echo "  sudo apt install ufw"
    exit 1
fi

# Reset ufw to default
print_status "Resetting ufw to default..."
ufw --force reset

# Set default policies
print_status "Setting default policies..."
ufw default deny incoming
ufw default allow outgoing

# Allow SSH
print_status "Allowing SSH..."
ufw allow 22/tcp

# Allow Fluentum testnet ports
print_status "Allowing Fluentum testnet ports..."

# P2P ports (26656-26660)
ufw allow 26656:26660/tcp

# RPC ports (26657-26660)
ufw allow 26657:26660/tcp

# API ports (1317-1320)
ufw allow 1317:1320/tcp

# Prometheus metrics port
ufw allow 26660/tcp

# Allow specific server IPs for better security
print_status "Allowing specific server IPs..."
ufw allow from 34.44.129.207 to any port 26656:26660
ufw allow from 34.44.82.114 to any port 26656:26660
ufw allow from 34.68.180.153 to any port 26656:26660
ufw allow from 34.72.252.153 to any port 26656:26660

# Enable ufw
print_status "Enabling ufw..."
ufw --force enable

# Show status
print_success "Firewall configuration completed!"
echo ""
print_status "Current firewall status:"
ufw status verbose

echo ""
print_status "Firewall rules summary:"
echo "  - SSH (22/tcp): Allowed"
echo "  - P2P ports (26656-26660/tcp): Allowed"
echo "  - RPC ports (26657-26660/tcp): Allowed"
echo "  - API ports (1317-1320/tcp): Allowed"
echo "  - Prometheus (26660/tcp): Allowed"
echo "  - Inter-node communication: Restricted to testnet IPs"
echo ""
print_warning "Remember to test connectivity between nodes after deployment!" 