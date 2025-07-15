#!/bin/bash

# Fluentum Testnet Connectivity Fix Script

# Root privilege check
auto_fix=false
if [ "$EUID" -ne 0 ]; then
    echo -e "\033[1;33m[WARNING]\033[0m Some checks (firewall, netstat, iptables) may require root privileges. Consider rerunning with: sudo $0 $@\n"
fi
# Parse --auto-fix flag
for arg in "$@"; do
    if [ "$arg" = "--auto-fix" ]; then
        auto_fix=true
    fi
done
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
    ["fluentum-node1"]="35.184.255.225"
    ["fluentum-node2"]="34.44.82.114"
    ["fluentum-node3"]="34.68.180.153"
    ["fluentum-node4"]="34.72.252.153"
    ["fluentum-node5"]="35.225.118.226"
)

# Port configurations
declare -A RPC_PORTS=(
    ["fluentum-node1"]="26657"
    ["fluentum-node2"]="26657"
    ["fluentum-node3"]="26657"
    ["fluentum-node4"]="26657"
    ["fluentum-node5"]="26657"
)

declare -A P2P_PORTS=(
    ["fluentum-node1"]="26656"
    ["fluentum-node2"]="26656"
    ["fluentum-node3"]="26656"
    ["fluentum-node4"]="26656"
    ["fluentum-node5"]="26656"
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
detect_fluentum_service() {
    # Try to find the correct fluentum service name
    local service="fluentum-testnet.service"
    if systemctl list-units --type=service | grep -q "$service"; then
        echo "$service"
        return
    fi
    # Fallback: try to find any fluentum-* service
    local alt_service=$(systemctl list-units --type=service | grep 'fluentum-' | awk '{print $1}' | head -n1)
    if [ -n "$alt_service" ]; then
        echo "$alt_service"
        return
    fi
    # Not found
    echo ""
}

check_local_service() {
    print_status "Checking local service status..."
    local service_name=$(detect_fluentum_service)
    if [ -z "$service_name" ]; then
        print_error "No fluentum service found (fluentum-testnet.service or fluentum-*)!"
        echo "Please check installation or service naming."
        echo "Try: sudo systemctl list-units --type=service | grep fluentum"
        return
    fi
    if systemctl is-active --quiet "$service_name"; then
        print_success "Local $service_name is running"
        # Check local RPC endpoint
        local rpc_port="26657"
        if curl -s --max-time 5 "http://localhost:$rpc_port/status" > /dev/null 2>&1; then
            print_success "Local RPC endpoint (localhost:$rpc_port) is responding"
        else
            print_error "Local RPC endpoint (localhost:$rpc_port) is not responding"
            echo "  Suggestion: Check service logs with: sudo journalctl -u $service_name -f"
        fi
        # Check what ports are actually listening
        print_status "Checking listening ports..."
        netstat -tulpn | grep -E "(26656|26657|26658|26659|26660)" || echo "No fluentum ports found listening"
    else
        print_error "Local $service_name is not running"
        echo "Start it with: sudo systemctl start $service_name"
        echo "Check logs: sudo journalctl -u $service_name -f"
    fi
    echo ""
}

# Function to check firewall status
check_firewall() {
    print_status "Checking firewall status..."
    local required_ports=(26656 26657)
    local missing_ufw=()
    local missing_iptables=()
    # Check UFW status
    if command -v ufw &> /dev/null; then
        local ufw_status=$(sudo ufw status 2>/dev/null | head -1)
        echo "UFW Status: $ufw_status"
        if echo "$ufw_status" | grep -q "inactive"; then
            print_warning "UFW is inactive - this might be good for testing"
        else
            print_status "UFW is active - checking rules..."
            local ufw_rules=$(sudo ufw status numbered)
            for port in "${required_ports[@]}"; do
                if ! echo "$ufw_rules" | grep -q "$port"; then
                    missing_ufw+=("$port")
                fi
            done
            echo "$ufw_rules" | grep -E "(26656|26657|26658|26659|26660)" || echo "No fluentum ports in UFW rules"
        fi
    else
        print_warning "UFW not found"
        missing_ufw=(26656 26657)
    fi
    # Check iptables
    print_status "Checking iptables..."
    local iptables_rules=$(sudo iptables -L -n)
    for port in "${required_ports[@]}"; do
        if ! echo "$iptables_rules" | grep -q "$port"; then
            missing_iptables+=("$port")
        fi
    done
    echo "$iptables_rules" | grep -E "(26656|26657|26658|26659|26660)" || echo "No fluentum ports in iptables rules"
    # Suggest fixes if missing
    if [ ${#missing_ufw[@]} -gt 0 ]; then
        print_warning "Missing UFW rules for: ${missing_ufw[*]}"
        for port in "${missing_ufw[@]}"; do
            echo "  Suggestion: sudo ufw allow $port/tcp"
        done
    fi
    if [ ${#missing_iptables[@]} -gt 0 ]; then
        print_warning "Missing iptables rules for: ${missing_iptables[*]}"
        for port in "${missing_iptables[@]}"; do
            echo "  Suggestion: sudo iptables -A INPUT -p tcp --dport $port -j ACCEPT"
        done
    fi
    echo ""
}

# Function to test connectivity to other nodes
test_connectivity() {
    print_status "Testing connectivity to other nodes..."
    local current_node=$(get_current_node)
    echo "Current node: $current_node"
    echo ""
    local local_rules_ok=true
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
            echo "    Suggestion: Check GCP firewall, network routing, or if node is offline."
            continue
        fi
        # Test RPC port
        if timeout 5 bash -c "</dev/tcp/$ip/$rpc_port" 2>/dev/null; then
            print_success "  RPC port $rpc_port is open"
        else
            print_error "  RPC port $rpc_port is closed"
            echo "    Suggestion: Ensure firewall allows $rpc_port/tcp and GCP firewall is configured."
            echo "    To allow with ufw: sudo ufw allow $rpc_port/tcp"
            echo "    To allow with iptables: sudo iptables -A INPUT -p tcp --dport $rpc_port -j ACCEPT"
            if [ "$auto_fix" = true ]; then
                echo "    [AUTO-FIX] Applying ufw and iptables rules for RPC..."
                ufw allow $rpc_port/tcp 2>&1 && echo "      [OK] ufw rule applied" || echo "      [FAIL] ufw rule failed"
                iptables -A INPUT -p tcp --dport $rpc_port -j ACCEPT 2>&1 && echo "      [OK] iptables rule applied" || echo "      [FAIL] iptables rule failed"
            else
                read -p "    Attempt to auto-fix firewall for RPC port $rpc_port? [y/N]: " ans
                if [[ "$ans" =~ ^[Yy]$ ]]; then
                    ufw allow $rpc_port/tcp 2>&1 && echo "      [OK] ufw rule applied" || echo "      [FAIL] ufw rule failed"
                    iptables -A INPUT -p tcp --dport $rpc_port -j ACCEPT 2>&1 && echo "      [OK] iptables rule applied" || echo "      [FAIL] iptables rule failed"
                fi
            fi
            local_rules_ok=false
        fi
        # Test P2P port
        if timeout 5 bash -c "</dev/tcp/$ip/$p2p_port" 2>/dev/null; then
            print_success "  P2P port $p2p_port is open"
        else
            print_error "  P2P port $p2p_port is closed"
            echo "    Suggestion: Ensure firewall allows $p2p_port/tcp and GCP firewall is configured."
            echo "    To allow with ufw: sudo ufw allow $p2p_port/tcp"
            echo "    To allow with iptables: sudo iptables -A INPUT -p tcp --dport $p2p_port -j ACCEPT"
            if [ "$auto_fix" = true ]; then
                echo "    [AUTO-FIX] Applying ufw and iptables rules for P2P..."
                ufw allow $p2p_port/tcp 2>&1 && echo "      [OK] ufw rule applied" || echo "      [FAIL] ufw rule failed"
                iptables -A INPUT -p tcp --dport $p2p_port -j ACCEPT 2>&1 && echo "      [OK] iptables rule applied" || echo "      [FAIL] iptables rule failed"
            else
                read -p "    Attempt to auto-fix firewall for P2P port $p2p_port? [y/N]: " ans
                if [[ "$ans" =~ ^[Yy]$ ]]; then
                    ufw allow $p2p_port/tcp 2>&1 && echo "      [OK] ufw rule applied" || echo "      [FAIL] ufw rule failed"
                    iptables -A INPUT -p tcp --dport $p2p_port -j ACCEPT 2>&1 && echo "      [OK] iptables rule applied" || echo "      [FAIL] iptables rule failed"
                fi
            fi
            local_rules_ok=false
        fi
        # Test HTTP RPC endpoint
        if curl -s --max-time 5 "http://$ip:$rpc_port/status" > /dev/null 2>&1; then
            print_success "  HTTP RPC endpoint is responding"
        else
            print_error "  HTTP RPC endpoint is not responding"
            echo "    Suggestion: Check if the service is running and listening on $rpc_port, and firewall rules."
        fi
        echo ""
    done
    if [ "$local_rules_ok" = true ]; then
        print_status "All local firewall rules appear correct. If remote ports are still closed, check Google Cloud firewall configuration."
    fi
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
    echo "   - Ensure rules allow TCP 26656, 26657 from 0.0.0.0/0"
    echo ""
    echo "2. Check node configurations:"
    echo "   - Verify each node is running: sudo systemctl status fluentum-testnet.service (or fluentum-* service)"
    echo "   - If service not found, check installation and service files."
    echo "   - Check logs: sudo journalctl -u fluentum-testnet.service -f (or correct service name)"
    echo "   - Verify ports are listening: netstat -tulpn | grep fluentum"
    echo ""
    echo "3. Test connectivity manually:"
    echo "   - From each node, test: curl http://[NODE_IP]:[RPC_PORT]/status"
    echo "   - Test P2P: telnet [NODE_IP] [P2P_PORT] or nc -vz [NODE_IP] [P2P_PORT]"
    echo ""
    echo "4. Update persistent peers in config:"
    echo "   - Edit /opt/fluentum/config/config.toml"
    echo "   - Update persistent_peers with correct addresses"
    echo ""
    echo "5. If service is missing:"
    echo "   - Check for service files in /etc/systemd/system/ or /lib/systemd/system/"
    echo "   - Reinstall or recreate the service if needed."
    echo ""
    echo "6. If ports are closed remotely but open locally:"
    echo "   - Check Google Cloud firewall and VPC routes."
    echo "   - Ensure no host-based firewall is blocking traffic."
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
    echo "9. Auto-fix firewall rules for all nodes (ufw & iptables)"
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
        9)
            auto_fix_all_firewall
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

auto_fix_all_firewall() {
    print_status "Auto-fixing firewall rules for all nodes (ufw & iptables)..."
    auto_fix=true
    test_connectivity
    print_success "Auto-fix completed."
}

# If --auto-fix is set, skip menu and run auto-fix directly
if [ "$auto_fix" = true ] && [[ "$*" == *--auto-fix* ]]; then
    auto_fix_all_firewall
    exit 0
fi

# Run main function
main "$@"
