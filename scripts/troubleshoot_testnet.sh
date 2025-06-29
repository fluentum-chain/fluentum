#!/bin/bash

# Fluentum Testnet Troubleshooting Script
# This script helps diagnose and fix issues with testnet nodes

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

echo "=========================================="
echo "    Fluentum Testnet Troubleshooting"
echo "=========================================="
echo ""

# Function to check systemd services
check_systemd_services() {
    print_status "Checking systemd services..."
    
    echo "Available fluentum services:"
    sudo systemctl list-unit-files | grep fluentum || echo "No fluentum services found"
    echo ""
    
    echo "Running fluentum services:"
    sudo systemctl list-units | grep fluentum || echo "No fluentum services running"
    echo ""
    
    # Check specific service
    if sudo systemctl list-unit-files | grep -q fluentum-testnet.service; then
        print_status "fluentum-testnet.service found, checking status..."
        sudo systemctl status fluentum-testnet.service --no-pager
        echo ""
    else
        print_error "fluentum-testnet.service not found"
    fi
}

# Function to check processes
check_processes() {
    print_status "Checking running processes..."
    
    echo "fluentumd processes:"
    ps aux | grep fluentumd | grep -v grep || echo "No fluentumd processes found"
    echo ""
    
    echo "Network connections:"
    netstat -tulpn | grep -E "(26656|26657|26658)" || echo "No fluentum ports listening"
    echo ""
}

# Function to check files and directories
check_files() {
    print_status "Checking files and directories..."
    
    # Check binary
    echo "fluentumd binary:"
    if command -v fluentumd &> /dev/null; then
        print_success "fluentumd found: $(which fluentumd)"
        ls -la $(which fluentumd)
    else
        print_error "fluentumd not found in PATH"
    fi
    echo ""
    
    # Check home directories
    echo "Home directories:"
    for dir in "/opt/fluentum" "/tmp/fluentum-node1" "$HOME/.fluentum"; do
        if [ -d "$dir" ]; then
            print_success "Found: $dir"
            ls -la "$dir"
            if [ -d "$dir/config" ]; then
                echo "  Config files:"
                ls -la "$dir/config/"
            fi
        else
            print_warning "Not found: $dir"
        fi
        echo ""
    done
}

# Function to check logs
check_logs() {
    print_status "Checking logs..."
    
    echo "Recent systemd logs for fluentum services:"
    sudo journalctl -u fluentum-testnet.service --no-pager -n 20 || echo "No logs found"
    echo ""
    
    echo "Recent system logs containing 'fluentum':"
    sudo journalctl | grep fluentum | tail -10 || echo "No fluentum logs found"
    echo ""
}

# Function to test RPC endpoints
test_rpc_endpoints() {
    print_status "Testing RPC endpoints..."
    
    local endpoints=("localhost:26657" "localhost:26658" "localhost:26659" "localhost:26660")
    
    for endpoint in "${endpoints[@]}"; do
        local port=$(echo "$endpoint" | cut -d: -f2)
        echo "Testing $endpoint..."
        if curl -s --max-time 5 "http://$endpoint/status" > /dev/null 2>&1; then
            print_success "RPC endpoint $endpoint is responding"
            curl -s "http://$endpoint/status" | jq '.result.sync_info' 2>/dev/null || echo "  Response received but not JSON"
        else
            print_warning "RPC endpoint $endpoint is not responding"
        fi
        echo ""
    done
}

# Function to fix common issues
fix_common_issues() {
    print_status "Attempting to fix common issues..."
    
    # Check if service file exists
    if [ ! -f "/etc/systemd/system/fluentum-testnet.service" ]; then
        print_error "Service file not found. You need to run the setup script first."
        echo "Run: ./scripts/setup_testnet.sh fluentum-node1 1"
        return 1
    fi
    
    # Check if fluentumd binary exists
    if [ ! -f "./build/fluentumd" ]; then
        print_error "fluentumd binary not found. Please build the project first:"
        echo "  make build"
        return 1
    fi
    
    # Check if home directory exists
    if [ ! -d "/opt/fluentum" ]; then
        print_error "Home directory /opt/fluentum not found. You need to run the setup script first."
        echo "Run: ./scripts/setup_testnet.sh fluentum-node1 1"
        return 1
    fi
    
    # Try to start the service
    print_status "Attempting to start the service..."
    sudo systemctl daemon-reload
    sudo systemctl start fluentum-testnet.service
    
    # Wait a moment and check status
    sleep 3
    if sudo systemctl is-active --quiet fluentum-testnet.service; then
        print_success "Service started successfully!"
    else
        print_error "Failed to start service. Checking logs..."
        sudo systemctl status fluentum-testnet.service --no-pager
        echo ""
        print_status "Recent logs:"
        sudo journalctl -u fluentum-testnet.service --no-pager -n 10
    fi
}

# Function to show manual start options
show_manual_options() {
    print_status "Manual start options:"
    echo ""
    echo "1. Start with systemd service:"
    echo "   sudo systemctl start fluentum-testnet.service"
    echo ""
    echo "2. Start manually:"
    echo "   ./build/fluentumd start --home /opt/fluentum --moniker fluentum-node1 --chain-id fluentum-testnet-1 --testnet"
    echo ""
    echo "3. Start with start script (if exists):"
    echo "   /opt/fluentum/start_node.sh"
    echo ""
    echo "4. Check service status:"
    echo "   sudo systemctl status fluentum-testnet.service"
    echo ""
    echo "5. View logs:"
    echo "   sudo journalctl -u fluentum-testnet.service -f"
}

# Main function
main() {
    echo "What would you like to do?"
    echo "1. Check systemd services"
    echo "2. Check processes and network"
    echo "3. Check files and directories"
    echo "4. Check logs"
    echo "5. Test RPC endpoints"
    echo "6. Fix common issues"
    echo "7. Show manual start options"
    echo "8. Run all checks"
    echo "9. Exit"
    echo ""
    
    read -p "Enter your choice (1-9): " choice
    
    case $choice in
        1)
            check_systemd_services
            ;;
        2)
            check_processes
            ;;
        3)
            check_files
            ;;
        4)
            check_logs
            ;;
        5)
            test_rpc_endpoints
            ;;
        6)
            fix_common_issues
            ;;
        7)
            show_manual_options
            ;;
        8)
            check_systemd_services
            check_processes
            check_files
            check_logs
            test_rpc_endpoints
            ;;
        9)
            print_status "Exiting..."
            exit 0
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

# Run main function
main "$@" 