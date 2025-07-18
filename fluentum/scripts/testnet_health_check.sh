#!/bin/bash

# Fluentum Testnet Health Check Script
# Version: 3.2.0
# Description: Comprehensive health check with advanced diagnostics and troubleshooting
# Features:
# - Detailed network connectivity checks
# - Node synchronization status
# - Resource utilization monitoring
# - JSON output support
# - Email/Telegram alerts

set -euo pipefail

# Ensure required commands are available
for cmd in curl jq nc; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd is required but not installed." >&2
        exit 1
    fi
done

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
NODE_IP="localhost"
RPC_PORT=26657
API_PORT=1317
GRPC_PORT=9090
CHECK_INTERVAL=60  # seconds
LOG_FILE="/var/log/fluentum_health.log"
ALERT_THRESHOLD=3  # Number of failed attempts before alerting
OUTPUT_FORMAT="text"  # text or json
DIAGNOSE=true      # Run detailed diagnostics
TIMEOUT=5          # Connection timeout in seconds

# Node configuration - can be overridden with --nodes flag
# Format: NODE_ID:IP:RPC_PORT:API_PORT:GRPC_PORT
NODES=(
    "node1:34.30.12.211:26657:1317:9090"
    "node2:35.232.125.109:26657:1318:9091"
)

# Node IDs for peer checking
NODE_IDS=(
    "node1:ddd24452832859f5f60fcdc768526985a3b9acec"
    "node2:7d0d3edf3a91d1d211280803521c0def4ec5c946"
)

# Alert configuration
ALERT_EMAIL=""  # Set this to receive email alerts
ALERT_TELEGRAM_CHAT_ID=""
ALERT_TELEGRAM_BOT_TOKEN=""

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if jq is installed, if not use python
if command_exists jq; then
    JSON_PROCESSOR="jq"
elif command_exists python3; then
    JSON_PROCESSOR="python3 -c \"import sys, json; print(json.load(sys.stdin)""['result'])\""
else
    echo "Error: jq or python3 is required but not installed."
    exit 1
fi

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --ip)
            NODE_IP="$2"
            shift 2
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift 2
            ;;
        --api-port)
            API_PORT="$2"
            shift 2
            ;;
        --interval)
            CHECK_INTERVAL="$2"
            shift 2
            ;;
        --log)
            LOG_FILE="$2"
            shift 2
            ;;
        --format)
            OUTPUT_FORMAT="$2"
            shift 2
            [[ "$OUTPUT_FORMAT" != "json" && "$OUTPUT_FORMAT" != "text" ]] && {
                echo "Error: Invalid format. Must be 'text' or 'json'"
                exit 1
            }
            ;;
        --nodes)
            # Override default nodes
            shift
            NODES=()
            while [[ $# -gt 0 && ! $1 =~ ^-- ]]; do
                NODES+=("$1")
                shift
            done
            ;;
        --help)
            echo "Fluentum Testnet Health Check"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --ip IP                Set the node IP (default: localhost)"
            echo "  --rpc-port PORT        Set the RPC port (default: 26657)"
            echo "  --api-port PORT        Set the API port (default: 1317)"
            echo "  --interval SECONDS     Set check interval in seconds (default: 60)"
            echo "  --log FILE             Set log file (default: /var/log/fluentum_health.log)"
            echo "  --format FORMAT        Output format: text or json (default: text)"
            echo "  --nodes NODE1,NODE2    Override default nodes (format: id:ip:rpc_port:api_port:grpc_port)"
            echo "  --help                 Show this help message"
            echo ""
            echo "Example:"
            echo "  $0 --format json"
            echo "  $0 --nodes 'node1:1.2.3.4:26657:1317:9090' 'node2:5.6.7.8:26657:1318:9091'"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create log directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"

# Function to log messages
log() {
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    local log_msg="[$timestamp] $1"
    echo -e "$log_msg" | tee -a "$LOG_FILE"
}

# Function to send alerts
alert() {
    local message="[FLUENTUM ALERT] $1"
    log "ALERT: $1"
    
    # Send email alert if configured
    if [ -n "$ALERT_EMAIL" ]; then
        echo "$message" | mail -s "Fluentum Node Alert" "$ALERT_EMAIL"
    fi
    
    # Send Telegram alert if configured
    if [ -n "$ALERT_TELEGRAM_CHAT_ID" ] && [ -n "$ALERT_TELEGRAM_BOT_TOKEN" ]; then
        curl -s -X POST "https://api.telegram.org/bot${ALERT_TELEGRAM_BOT_TOKEN}/sendMessage" \
            -d "chat_id=${ALERT_TELEGRAM_CHAT_ID}" \
            -d "text=${message}" \
            -d "parse_mode=Markdown"
    fi
}

# Function to parse JSON response
parse_json() {
    local json_data=$1
    local key=$2
    
    if [ "$JSON_PROCESSOR" = "jq" ]; then
        echo "$json_data" | jq -r "$key" 2>/dev/null
    else
        # Simple Python-based JSON parser
        echo "$json_data" | python3 -c "import sys, json; print(json.load(sys.stdin)$key or '')" 2>/dev/null || echo ""
    fi
}

# Function to check RPC endpoint with detailed error reporting
check_rpc() {
    local node_ip=$1
    local node_rpc_port=$2
    local endpoint=$3
    local full_url="http://${node_ip}:${node_rpc_port}/${endpoint}"
    
    # Try to get both the response and HTTP status code
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" --connect-timeout $TIMEOUT --max-time $TIMEOUT "$full_url" 2>/dev/null) || {
        log "[DEBUG] curl failed to $full_url"
        echo "ERROR:CONNECTION_FAILED"
        return 1
    }
    
    # Extract the HTTP status code and response body
    http_code=$(echo "$response" | tail -n1)
    response=$(echo "$response" | sed '$d')
    
    if [[ -z "$response" || "$http_code" == "000" ]]; then
        log "[DEBUG] No response from $full_url (HTTP $http_code)"
        echo "ERROR:NO_RESPONSE"
        return 1
    fi
    
    if [[ "$http_code" != "200" ]]; then
        log "[DEBUG] HTTP $http_code from $full_url"
        echo "ERROR:HTTP_$http_code"
        return 1
    fi
    
    # Check if response is valid JSON
    if ! echo "$response" | jq -e . >/dev/null 2>&1; then
        log "[DEBUG] Invalid JSON response from $full_url"
        echo "ERROR:INVALID_JSON"
        return 1
    fi
    
    # Extract the result field if it exists
    local result
    result=$(echo "$response" | jq -r '.result' 2>/dev/null || echo "$response")
    
    if [ -z "$result" ] || [ "$result" = "null" ]; then
        log "[DEBUG] Empty result from $full_url"
        echo "ERROR:EMPTY_RESULT"
        return 1
    fi
    
    echo "$result"
    return 0
}

# Function to diagnose connection issues
diagnose_connection() {
    local node_ip=$1
    local port=$2
    local service=$3
    
    log "[DIAGNOSE] Checking $service connection to $node_ip:$port"
    
    # Check if port is open
    if ! nc -z -w $TIMEOUT "$node_ip" "$port" 2>/dev/null; then
        echo "  - ❌ Port $port is not open"
        return 1
    fi
    
    echo "  - ✅ Port $port is open"
    
    # Check system resources if local node
    if [ "$node_id" = "$LOCAL_NODE_ID" ]; then
        echo -e "\n${BLUE}=== System Resources ===${NC}"
        
        # Disk usage
        echo -n "  • Disk usage: "
        local disk_pct
        disk_pct=$(df -h / | awk 'NR==2 {print $5}' | tr -d '%')
        local disk_free
        disk_free=$(df -h / | awk 'NR==2 {print $4}')
        
        if [ "$disk_pct" -gt 90 ]; then
            echo -e "${RED}❌ Critical: ${disk_pct}% used (${disk_free}B free)${NC}"
            health_status=$((health_status + 1))
        elif [ "$disk_pct" -gt 80 ]; then
            echo -e "${YELLOW}⚠️  Warning: ${disk_pct}% used (${disk_free}B free)${NC}"
            health_status=$((health_status + 1))
        else
            echo -e "${GREEN}✅ ${disk_pct}% used (${disk_free}B free)${NC}"
        fi
        
        # Memory usage
        echo -n "  • Memory usage: "
        local mem_total
        mem_total=$(free -m | awk '/Mem:/ {print $2}')
        local mem_used
        mem_used=$(free -m | awk '/Mem:/ {print $3}')
        local mem_pct
        mem_pct=$((mem_used * 100 / mem_total))
        
        if [ "$mem_pct" -gt 90 ]; then
            echo -e "${RED}❌ Critical: ${mem_pct}% used (${mem_used}M/${mem_total}M)${NC}"
            health_status=$((health_status + 1))
        elif [ "$mem_pct" -gt 80 ]; then
            echo -e "${YELLOW}⚠️  Warning: ${mem_pct}% used (${mem_used}M/${mem_total}M)${NC}"
            health_status=$((health_status + 1))
        else
            echo -e "${GREEN}✅ ${mem_pct}% used (${mem_used}M/${mem_total}M)${NC}"
        fi
        
        # Check if fluentumd is running
        echo -n "  • Fluentum service: "
        if systemctl is-active --quiet fluentumd; then
            echo -e "${GREEN}✅ Running${NC}"
        else
            echo -e "${RED}❌ Not running${NC}"
            health_status=$((health_status + 1))
        fi
    fi
    
    return 0
}

# Function to check node health
check_node_health() {
    local node_id=$1
    local node_ip=$2
    local node_rpc_port=$3
    local node_api_port=$4
    
    local health_status=0
    local status_info=""
    local node_info_json=""
    local connection_ok=true
    
    # Run diagnostics if enabled
    if [ "$DIAGNOSE" = true ]; then
        echo -e "\n${BLUE}=== Diagnosing $node_id ($node_ip) ===${NC}"
        
        # Check RPC connection
        if ! diagnose_connection "$node_ip" "$node_rpc_port" "RPC"; then
            status_info="${status_info}RPC connection failed. "
            health_status=$((health_status + 1))
            connection_ok=false
        fi
        
        # Check P2P connection
        if ! diagnose_connection "$node_ip" "$((node_rpc_port - 1))" "P2P"; then
            status_info="${status_info}P2P connection failed. "
            health_status=$((health_status + 1))
            connection_ok=false
        fi
        
        # Skip further checks if basic connections failed
        if [ "$connection_ok" = false ]; then
            echo -e "${RED}❌ Basic connectivity issues detected.${NC}\n"
            if [ "$OUTPUT_FORMAT" = "json" ]; then
                echo "{\"node_id\":\"$node_id\",\"ip\":\"$node_ip\",\"rpc_port\":$node_rpc_port,\"api_port\":$node_api_port,\"health_status\":$health_status,\"status\":\"${status_info}\"}"
            else
                echo -e "${RED}❌ $node_id is not healthy: $status_info${NC}\n"
            fi
            return 1
        fi
    fi
    
    # Check if node is responding
    local node_status
    node_status=$(check_rpc "$node_ip" "$node_rpc_port" "status")
    
    # Handle different error cases
    if [[ "$node_status" == ERROR:* ]]; then
        local error_type=${node_status#ERROR:}
        case $error_type in
            CONNECTION_FAILED)
                status_info="Connection failed. Check if node is running and accessible."
                ;;
            NO_RESPONSE)
                status_info="No response from node. Check if RPC port is open and node is running."
                ;;
            HTTP_*)
                status_info="HTTP ${error_type#HTTP_} error. Check node configuration."
                ;;
            INVALID_JSON)
                status_info="Invalid response format. Node might be initializing."
                ;;
            EMPTY_RESULT)
                status_info="Empty response from node. Check node logs for errors."
                ;;
            *)
                status_info="Unknown error: $error_type"
                ;;
        esac
        
        if [ "$OUTPUT_FORMAT" = "json" ]; then
            echo "{\"node_id\":\"$node_id\",\"ip\":\"$node_ip\",\"rpc_port\":$node_rpc_port,\"api_port\":$node_api_port,\"health_status\":1,\"status\":\"$status_info\"}"
        else
            echo -e "${RED}❌ $node_id: $status_info${NC}\n"
        fi
        return 1
    fi
    
    # Get node info
    local node_info=$(echo "$node_status" | jq -r '.node_info' 2>/dev/null)
    local sync_info=$(echo "$node_status" | jq -r '.sync_info' 2>/dev/null)
    
    # Check node version
    local node_version=$(echo "$node_info" | jq -r '.version')
    if [[ ! "$node_version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+ ]]; then
        status_info="${status_info}Invalid version: ${node_version}. "
        health_status=$((health_status + 1))
    fi
    
    # Check sync status
    local catching_up=$(echo "$sync_info" | jq -r '.catching_up')
    if [ "$catching_up" = "true" ]; then
        status_info="${status_info}Node is catching up. "
        health_status=$((health_status + 1))
    fi
    
    # Check latest block height
    local latest_block=$(echo "$sync_info" | jq -r '.latest_block_height')
    if [ "$latest_block" = "0" ] || [ "$latest_block" = "null" ]; then
        status_info="${status_info}Invalid block height. "
        health_status=$((health_status + 1))
    fi
    
    # Check quantum signing status
    local quantum_status=$(check_rpc "$node_ip" "$node_api_port" "fluentum/features/quantum_signing/status")
    if [ "$quantum_status" = "ERROR" ]; then
        status_info="${status_status}Quantum signing not active. "
        health_status=$((health_status + 1))
    else
        local quantum_enabled=$(echo "$quantum_status" | jq -r '.enabled')
        if [ "$quantum_enabled" != "true" ]; then
            status_info="${status_info}Quantum signing disabled. "
            health_status=$((health_status + 1))
        fi
    fi
    
    # Check AI validation status
    local ai_status=$(check_rpc "$node_ip" "$node_api_port" "fluentum/features/ai_validation/status")
    if [ "$ai_status" = "ERROR" ]; then
        status_info="${status_status}AI validation not active. "
        health_status=$((health_status + 1))
    else
        local ai_enabled=$(echo "$ai_status" | jq -r '.enabled')
        if [ "$ai_enabled" != "true" ]; then
            status_info="${status_info}AI validation disabled. "
            health_status=$((health_status + 1))
        fi
    fi
    
    # Check peer count
    local peers=$(check_rpc "$node_ip" "$node_rpc_port" "net_info" | jq -r '.n_peers' 2>/dev/null || echo "0")
    if [ "$peers" -lt 2 ]; then
        status_info="${status_info}Low peer count: ${peers}. "
        health_status=$((health_status + 1))
    fi
    
    # If no issues found
    if [ -z "$status_info" ]; then
        status_info="Node is healthy"
    fi
    
    echo "${health_status},${status_info}"
}

# Function to check connected peers via RPC
check_connected_peers() {
    local NODE_NAME=$1
    local NODE_IP=$2
    local RPC_PORT=$3
    local NODE_ID=$4
    local EXPECTED_IDS=()
    for entry in "${NODE_IDS[@]}"; do
        IFS=":" read -r PEER_NAME PEER_ID <<< "$entry"
        if [[ $PEER_NAME != $NODE_NAME ]]; then
            EXPECTED_IDS+=("$PEER_ID")
        fi
    done
    local PEERS=$(curl -s --max-time $TIMEOUT http://$NODE_IP:$RPC_PORT/net_info | jq -r '.result.peers[].node_info.id')
    for expected in "${EXPECTED_IDS[@]}"; do
        if echo "$PEERS" | grep -q "$expected"; then
            echo -e "$GREEN[SUCCESS]$NC Connected to peer $expected"
        else
            echo -e "$RED[ERROR]$NC Not connected to expected peer $expected"
        fi
    done
}

# Function to check for stuck node
check_stuck_node() {
    local NODE_NAME=$1
    local BLOCK_HEIGHT=$2
    local TMP_DIR="/tmp/fluentum_health"
    mkdir -p "$TMP_DIR"
    local TMP_FILE="$TMP_DIR/${NODE_NAME}_block_height.tmp"
    local PREV_HEIGHT=0
    if [ -f "$TMP_FILE" ]; then
        PREV_HEIGHT=$(cat "$TMP_FILE")
    fi
    echo "$BLOCK_HEIGHT" > "$TMP_FILE"
    if [ "$BLOCK_HEIGHT" = "$PREV_HEIGHT" ]; then
        echo -e "$YELLOW[WARNING]$NC Node $NODE_NAME block height stuck at $BLOCK_HEIGHT!"
    else
        echo -e "$GREEN[SUCCESS]$NC Node $NODE_NAME block height is advancing: $BLOCK_HEIGHT"
    fi
}

# Function to print node status in text format
print_node_status_text() {
    local status=$1
    local node_id=$(parse_json "$status" '["node_id"]')
    local ip=$(parse_json "$status" '["ip"]')
    local version=$(parse_json "$status" '["version"]')
    local block_height=$(parse_json "$status" '["block_height"]')
    local catching_up=$(parse_json "$status" '["catching_up"]')
    local is_validator=$(parse_json "$status" '["is_validator"]')
    local peers=$(parse_json "$status" '["peers"]')
    local disk_usage=$(parse_json "$status" '["disk_usage"]')
    local memory_usage=$(parse_json "$status" '["memory_usage"]')
    local health_status=$(parse_json "$status" '["health_status"]')
    local status_info=$(parse_json "$status" '["status"]')
    
    local status_color=$GREEN
    if [ "$health_status" -gt 0 ]; then
        status_color=$RED
    elif [ "$catching_up" = "true" ]; then
        status_color=$YELLOW
    fi
    
    echo -e "${BLUE}=== Node: $node_id ($ip) ===${NC}"
    echo -e "Status: ${status_color}${status_info}${NC}"
    echo -e "Version: $version | Block: $block_height | Catching Up: $catching_up"
    echo -e "Validator: $is_validator | Peers: $peers | Disk: ${disk_usage}% | Memory: ${memory_usage}%"
    echo -e "----------------------------------------\n"
}

# Arrays to hold summary info
SUMMARY_NODE_NAMES=()
SUMMARY_NODE_IPS=()
SUMMARY_BLOCK_HEIGHTS=()
SUMMARY_CATCHING_UPS=()
SUMMARY_PEER_STATUSES=()
SUMMARY_STUCK_STATUSES=()
SUMMARY_VERSIONS=()
SUMMARY_VALIDATOR_STATUSES=()
SUMMARY_DISK_USAGES=()
SUMMARY_MEMORY_USAGES=()
SUMMARY_HEALTH_STATUSES=()
SUMMARY_STATUS_INFOS=()

# Main monitoring loop
log "Starting Fluentum Testnet Health Monitor"
log "Monitoring ${#NODES[@]} nodes"
log "Check interval: ${CHECK_INTERVAL} seconds"
log "Log file: ${LOG_FILE}"
log "Output format: ${OUTPUT_FORMAT}"

# Initialize alert counters
declare -A alert_counters
for node in "${NODES[@]}"; do
    node_id=$(echo "$node" | cut -d':' -f1)
    alert_counters["$node_id"]=0
done

while true; do
    log "=== Starting health check at $(date) ==="
    
    # Initialize JSON output array if in JSON mode
    if [ "$OUTPUT_FORMAT" = "json" ]; then
        json_output="["
        first_node=true
    fi
    
    for node in "${NODES[@]}"; do
        IFS=':' read -r node_id node_ip node_rpc_port node_api_port node_grpc_port <<< "$node"
        
        log "Checking node ${node_id} (${node_ip}:${node_rpc_port})..."
        
        # Check node health
        result=$(check_node_health "$node_id" "$node_ip" "$node_rpc_port" "$node_api_port")
        
        # Extract summary info
        version=$(curl -s --max-time $TIMEOUT http://$node_ip:$node_rpc_port/status | jq -r '.result.node_info.version' 2>/dev/null)
        block_height=$(curl -s --max-time $TIMEOUT http://$node_ip:$node_rpc_port/status | jq -r '.result.sync_info.latest_block_height' 2>/dev/null)
        catching_up=$(curl -s --max-time $TIMEOUT http://$node_ip:$node_rpc_port/status | jq -r '.result.sync_info.catching_up' 2>/dev/null)
        is_validator=$(curl -s --max-time $TIMEOUT http://$node_ip:$node_rpc_port/status | jq -r '.result.validator_info.voting_power' 2>/dev/null)
        disk_usage=$(df -h / | awk 'NR==2 {print $5}')
        memory_usage=$(free -m | awk '/Mem:/ { printf("%d", $3*100/$2) }')
        health_status=$(echo "$result" | cut -d',' -f1)
        status_info=$(echo "$result" | cut -d',' -f2-)
        # Peer status
        peer_status="$(check_connected_peers "$node_id" "$node_ip" "$node_rpc_port" "$node_id" | grep -Eo '\[SUCCESS\]|\[ERROR\]')"
        # Stuck status
        stuck_status="$(check_stuck_node "$node_id" "$block_height" | grep -Eo '\[SUCCESS\]|\[WARNING\]')"

        SUMMARY_NODE_NAMES+=("$node_id")
        SUMMARY_NODE_IPS+=("$node_ip")
        SUMMARY_BLOCK_HEIGHTS+=("$block_height")
        SUMMARY_CATCHING_UPS+=("$catching_up")
        SUMMARY_VERSIONS+=("$version")
        SUMMARY_VALIDATOR_STATUSES+=("$is_validator")
        SUMMARY_DISK_USAGES+=("$disk_usage")
        SUMMARY_MEMORY_USAGES+=("$memory_usage")
        SUMMARY_HEALTH_STATUSES+=("$health_status")
        SUMMARY_STATUS_INFOS+=("$status_info")
        SUMMARY_PEER_STATUSES+=("$peer_status")
        SUMMARY_STUCK_STATUSES+=("$stuck_status")

        if [ "$OUTPUT_FORMAT" = "json" ]; then
            # Add to JSON array
            if [ "$first_node" = true ]; then
                json_output="${json_output}${result}"
                first_node=false
            else
                json_output="${json_output},${result}"
            fi
            
            # Extract health status for logging
            health_status=$(parse_json "$result" '["health_status"]')
            status_info=$(parse_json "$result" '["status"]')
        else
            # Text format - already processed in the function
            health_status=$(echo "$result" | cut -d',' -f1)
            status_info=$(echo "$result" | cut -d',' -f2)
            node_json=$(echo "$result" | cut -d',' -f3-)
            print_node_status_text "$node_json"
        fi
        
        # Process health status for alerts
        if [ "$health_status" -gt 0 ]; then
            alert_counters["$node_id"]=$((alert_counters["$node_id"] + 1))
            
            if [ "${alert_counters["$node_id"]}" -ge "$ALERT_THRESHOLD" ]; then
                alert "Node ${node_id} issue detected: ${status_info}"
                # Reset counter after alerting
                alert_counters["$node_id"]=0
            fi
            
            log "${RED}Node ${node_id} has issues: ${status_info}${NC}"
        else
            # Reset counter if node is healthy
            alert_counters["$node_id"]=0
            log "${GREEN}Node ${node_id} is healthy${NC}"
        fi
    done
    
    # Finalize JSON output
    if [ "$OUTPUT_FORMAT" = "json" ]; then
        json_output="${json_output}]"
        if [ "$health_status" -gt 0 ]; then
            echo -e "${RED}${json_output}${NC}"
        else
            echo -e "${GREEN}${json_output}${NC}"
        fi
    fi
    
    # Wait for the next check if not in single-run mode
    if [ "$CHECK_INTERVAL" -gt 0 ]; then
        log "Waiting ${CHECK_INTERVAL} seconds until next check..."
        sleep "$CHECK_INTERVAL"
    else
        # Single run mode
        break
    fi
done
