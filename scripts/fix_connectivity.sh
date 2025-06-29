#!/bin/bash

# Fluentum Testnet Connectivity Fix Script
# This script helps diagnose and fix connectivity issues between testnet nodes

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

# Testnet configuration
TESTNET_CHAIN_ID="fluentum-testnet-1"

# Server configurations with correct ports
declare -A SERVERS=(
    ["fluentum-node1"]="34.44.129.207"
    ["fluentum-node2"]="34.44.82.114"
    ["fluentum-node3"]="34.68.180.153"
    ["fluentum-node4"]="34.72.252.153"
)

# Port configurations
declare -A RPC_PORTS=(
    ["fluentum-node1"]="26657"
    ["fluentum-node2"]="26658"
    ["fluentum-node3"]="26659"
    ["fluentum-node4"]="26660"
)

declare -A P2P_PORTS=(
    ["fluentum-node1"]="26656"
    ["fluentum-node2"]="26657"
    ["fluentum-node3"]="26658"
    ["fluentum-node4"]="26659"
)

echo "=========================================="
echo "    Fluentum Testnet Connectivity Fix"
echo "=========================================="
echo ""

# Function to check current node
get_current_node() {
    local current_ip=$(curl -s ifconfig.me 2>/dev/null || echo "unknown")
    local current_node=""
    
    for node_name in "${!SERVERS[@]}"; do
        if [ "${SERVERS[$node_name]}" = "$current_ip" ]; then
            current_node=$node_name
            break
        fi
    done
    
    echo "$current_node"
}

# Function to check local service status
check_local_service() {
    print_status "Checking local service status..."
    
    if systemctl is-active --quiet fluentum-testnet.service; then
        print_success "Local fluentum-testnet service is running"
        
        # Check local RPC endpoint
        local rpc_port="26657"
        if curl -s --max-time 5 "http://localhost:$rpc_port/status" > /dev/null 2>&1; then
            print_success "Local RPC endpoint (localhost:$rpc_port) is responding"
        else
            print_error "Local RPC endpoint (localhost:$rpc_port) is not responding"
        fi
        
        # Check what ports are actually listening
        print_status "Checking listening ports..."
        netstat -tulpn | grep -E "(26656|26657|26658|26659|26660)" || echo "No fluentum ports found listening"
        
    else
        print_error "Local fluentum-testnet service is not running"
        echo "Start it with: sudo systemctl start fluentum-testnet.service"
    fi
    
    echo ""
}

# Function to check firewall status
check_firewall() {
    print_status "Checking firewall status..."
    
    # Check UFW status
    if command -v ufw &> /dev/null; then
        local ufw_status=$(sudo ufw status 2>/dev/null | head -1)
        echo "UFW Status: $ufw_status"
        
        if echo "$ufw_status" | grep -q "inactive"; then
            print_warning "UFW is inactive - this might be good for testing"
        else
            print_status "UFW is active - checking rules..."
            sudo ufw status numbered | grep -E "(26656|26657|26658|26659|26660)" || echo "No fluentum ports in UFW rules"
        fi
    else
        print_warning "UFW not found"
    fi
    
    # Check iptables
    print_status "Checking iptables..."
    sudo iptables -L -n | grep -E "(26656|26657|26658|26659|26660)" || echo "No fluentum ports in iptables rules"
    
    echo ""
}

# Function to test connectivity to other nodes
test_connectivity() {
    print_status "Testing connectivity to other nodes..."
    
    local current_node=$(get_current_node)
    echo "Current node: $current_node"
    echo ""
    
    for node_name in "${!SERVERS[@]}"; do
        local ip=${SERVERS[$node_name]}
        local rpc_port=${RPC_PORTS[$node_name]}
        local p2p_port=${P2P_PORTS[$node_name]}
        
        echo "Testing $node_name ($ip)..."
        
        # Test basic connectivity
        if ping -c 1 -W 3 "$ip" > /dev/null 2>&1; then
            print_success "  Ping successful"
        else
            print_error "  Ping failed"
            continue
        fi
        
        # Test RPC port
        if timeout 5 bash -c "</dev/tcp/$ip/$rpc_port" 2>/dev/null; then
            print_success "  RPC port $rpc_port is open"
        else
            print_error "  RPC port $rpc_port is closed"
        fi
        
        # Test P2P port
        if timeout 5 bash -c "</dev/tcp/$ip/$p2p_port" 2>/dev/null; then
            print_success "  P2P port $p2p_port is open"
        else
            print_error "  P2P port $p2p_port is closed"
        fi
        
        # Test HTTP RPC endpoint
        if curl -s --max-time 5 "http://$ip:$rpc_port/status" > /dev/null 2>&1; then
            print_success "  HTTP RPC endpoint is responding"
        else
            print_error "  HTTP RPC endpoint is not responding"
        fi
        
        echo ""
    done
}

# Function to configure firewall rules
configure_firewall() {
    print_status "Configuring firewall rules..."
    
    local current_node=$(get_current_node)
    echo "Current node: $current_node"
    
    # Get current node's ports
    local rpc_port=${RPC_PORTS[$current_node]}
    local p2p_port=${P2P_PORTS[$current_node]}
    
    echo "Configuring ports: RPC=$rpc_port, P2P=$p2p_port"
    
    # Configure UFW if available
    if command -v ufw &> /dev/null; then
        print_status "Configuring UFW rules..."
        
        # Allow RPC port
        sudo ufw allow $rpc_port/tcp
        print_success "Allowed RPC port $rpc_port"
        
        # Allow P2P port
        sudo ufw allow $p2p_port/tcp
        print_success "Allowed P2P port $p2p_port"
        
        # Allow SSH (important!)
        sudo ufw allow ssh
        print_success "Allowed SSH"
        
        # Enable UFW if not already enabled
        if ! sudo ufw status | grep -q "Status: active"; then
            print_warning "Enabling UFW..."
            echo "y" | sudo ufw enable
        fi
        
        print_success "UFW configured successfully"
    else
        print_warning "UFW not available, using iptables..."
        
        # Configure iptables
        sudo iptables -A INPUT -p tcp --dport $rpc_port -j ACCEPT
        sudo iptables -A INPUT -p tcp --dport $p2p_port -j ACCEPT
        sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT  # SSH
        
        print_success "iptables rules added"
    fi
    
    echo ""
}

# Function to check Google Cloud firewall
check_gcp_firewall() {
    print_status "Checking Google Cloud firewall configuration..."
    
    echo "IMPORTANT: You need to configure Google Cloud firewall rules!"
    echo ""
    echo "For each node, you need to create firewall rules allowing:"
    echo "1. RPC ports (26657-26660) for HTTP traffic"
    echo "2. P2P ports (26656-26659) for TCP traffic"
    echo "3. SSH (port 22) for management"
    echo ""
    echo "You can do this via:"
    echo "1. Google Cloud Console -> VPC Network -> Firewall"
    echo "2. Or use gcloud command line:"
    echo ""
    
    for node_name in "${!SERVERS[@]}"; do
        local ip=${SERVERS[$node_name]}
        local rpc_port=${RPC_PORTS[$node_name]}
        local p2p_port=${P2P_PORTS[$node_name]}
        
        echo "For $node_name ($ip):"
        echo "  gcloud compute firewall-rules create fluentum-rpc-$node_name \\"
        echo "    --allow tcp:$rpc_port \\"
        echo "    --source-ranges 0.0.0.0/0 \\"
        echo "    --description \"Fluentum RPC for $node_name\""
        echo ""
        echo "  gcloud compute firewall-rules create fluentum-p2p-$node_name \\"
        echo "    --allow tcp:$p2p_port \\"
        echo "    --source-ranges 0.0.0.0/0 \\"
        echo "    --description \"Fluentum P2P for $node_name\""
        echo ""
    done
}

# Function to restart services
restart_services() {
    print_status "Restarting services..."
    
    sudo systemctl restart fluentum-testnet.service
    
    # Wait a moment for service to start
    sleep 5
    
    if systemctl is-active --quiet fluentum-testnet.service; then
        print_success "Service restarted successfully"
    else
        print_error "Service failed to restart"
        sudo systemctl status fluentum-testnet.service --no-pager
    fi
    
    echo ""
}

# Function to show manual steps
show_manual_steps() {
    print_status "Manual steps to fix connectivity:"
    echo ""
    echo "1. Configure Google Cloud Firewall Rules:"
    echo "   - Go to Google Cloud Console"
    echo "   - Navigate to VPC Network -> Firewall"
    echo "   - Create rules for each node's RPC and P2P ports"
    echo ""
    echo "2. Check node configurations:"
    echo "   - Verify each node is running: sudo systemctl status fluentum-testnet.service"
    echo "   - Check logs: sudo journalctl -u fluentum-testnet.service -f"
    echo "   - Verify ports are listening: netstat -tulpn | grep fluentum"
    echo ""
    echo "3. Test connectivity manually:"
    echo "   - From each node, test: curl http://[NODE_IP]:[RPC_PORT]/status"
    echo "   - Test P2P: telnet [NODE_IP] [P2P_PORT]"
    echo ""
    echo "4. Update persistent peers in config:"
    echo "   - Edit /opt/fluentum/config/config.toml"
    echo "   - Update persistent_peers with correct addresses"
    echo ""
}

# Main execution
main() {
    echo "What would you like to do?"
    echo "1. Check local service status"
    echo "2. Check firewall status"
    echo "3. Test connectivity to other nodes"
    echo "4. Configure local firewall rules"
    echo "5. Show Google Cloud firewall instructions"
    echo "6. Restart services"
    echo "7. Show manual steps"
    echo "8. Run all checks"
    echo ""
    read -p "Enter your choice (1-8): " choice
    
    case $choice in
        1)
            check_local_service
            ;;
        2)
            check_firewall
            ;;
        3)
            test_connectivity
            ;;
        4)
            configure_firewall
            ;;
        5)
            check_gcp_firewall
            ;;
        6)
            restart_services
            ;;
        7)
            show_manual_steps
            ;;
        8)
            check_local_service
            check_firewall
            test_connectivity
            check_gcp_firewall
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
