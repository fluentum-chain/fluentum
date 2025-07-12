#!/bin/bash

# Fluentum Testnet Health Check Script
# Version: 2.0.0
# Description: Monitors the health of Fluentum testnet nodes with quantum signing and AI validation

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
NODE_IP="localhost"
RPC_PORT=26657
API_PORT=1317
GRPC_PORT=9090
CHECK_INTERVAL=60  # seconds
NODE_PREFIX="node"
TOTAL_NODES=5
LOG_FILE="/var/log/fluentum_health.log"
ALERT_THRESHOLD=3  # Number of failed attempts before alerting

# Alert configuration
ALERT_EMAIL=""  # Set this to receive email alerts
ALERT_TELEGRAM_CHAT_ID=""
ALERT_TELEGRAM_BOT_TOKEN=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --ip)
            NODE_IP="$2"
            shift
            shift
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift
            shift
            ;;
        --api-port)
            API_PORT="$2"
            shift
            shift
            ;;
        --nodes)
            TOTAL_NODES="$2"
            shift
            shift
            ;;
        --prefix)
            NODE_PREFIX="$2"
            shift
            shift
            ;;
        --interval)
            CHECK_INTERVAL="$2"
            shift
            shift
            ;;
        --log)
            LOG_FILE="$2"
            shift
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esc
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

# Function to check RPC endpoint
check_rpc() {
    local node_ip=$1
    local node_rpc_port=$2
    local endpoint=$3
    
    local response=$(curl -s "http://${node_ip}:${node_rpc_port}/${endpoint}" || echo "ERROR")
    if [ "$response" = "ERROR" ]; then
        echo "ERROR"
    else
        echo $response | jq -r '.result' 2>/dev/null || echo "ERROR"
    fi
}

# Function to check node health
check_node_health() {
    local node_id=$1
    local node_ip=$2
    local node_rpc_port=$3
    local node_api_port=$4
    
    local health_status=0
    local status_info=""
    
    # Check if node is responding
    local node_status=$(check_rpc "$node_ip" "$node_rpc_port" "status")
    if [ "$node_status" = "ERROR" ]; then
        echo "0,Node not responding"
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

# Main monitoring loop
log "Starting Fluentum Testnet Health Monitor"
log "Monitoring ${TOTAL_NODES} nodes with prefix '${NODE_PREFIX}'"
log "Check interval: ${CHECK_INTERVAL} seconds"
log "Log file: ${LOG_FILE}"

# Initialize alert counters
declare -A alert_counters
for ((i=1; i<=TOTAL_NODES; i++)); do
    alert_counters["${NODE_PREFIX}${i}"]=0
done

while true; do
    log "--- Starting health check at $(date) ---"
    
    for ((i=1; i<=TOTAL_NODES; i++)); do
        node_id="${NODE_PREFIX}${i}"
        node_rpc_port=$((RPC_PORT + (i - 1) * 10))
        node_api_port=$((API_PORT + (i - 1) * 10))
        
        log "Checking node ${node_id}..."
        
        # Check node health
        result=$(check_node_health "$node_id" "$NODE_IP" "$node_rpc_port" "$node_api_port")
        health_status=$(echo "$result" | cut -d',' -f1)
        status_info=$(echo "$result" | cut -d',' -f2-)
        
        # Process health status
        if [ "$health_status" -gt 0 ]; then
            alert_counters["$node_id"]=$((alert_counters["$node_id"] + 1))
            
            if [ "${alert_counters["$node_id"]}" -ge "$ALERT_THRESHOLD" ]; then
                alert "Node ${node_id} issue detected: ${status_info}"
                # Reset counter after alerting
                alert_counters["$node_id"]=0
            fi
            
            log "${RED}Node ${node_id}: ${status_info}${NC}"
        else
            # Reset counter if node is healthy
            alert_counters["$node_id"]=0
            log "${GREEN}Node ${node_id}: ${status_info}${NC}"
        fi
    done
    
    # Wait for the next check
    log "Waiting ${CHECK_INTERVAL} seconds until next check..."
    sleep "$CHECK_INTERVAL"
done
