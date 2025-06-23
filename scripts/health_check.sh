#!/bin/bash
# Fluentum Health Check Script
# This script verifies the health and status of a Fluentum node

set -e

# Configuration
RPC_ENDPOINT="http://localhost:26657"
LOG_FILE="/var/log/fluentum_health.log"
ALERT_EMAIL="admin@example.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "${RED}ERROR: $1${NC}"
    exit 1
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    error_exit "jq is required but not installed. Please install jq first."
fi

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    error_exit "curl is required but not installed. Please install curl first."
fi

echo "${BLUE}🔍 Fluentum Node Health Check${NC}"
echo "${BLUE}============================${NC}"
log "Starting Fluentum health check"

# Function to make RPC calls with error handling
rpc_call() {
    local endpoint="$1"
    local timeout="${2:-10}"
    
    curl -s --max-time "$timeout" "$RPC_ENDPOINT$endpoint" 2>/dev/null || echo "null"
}

# Function to check if JSON response is valid
is_valid_json() {
    local json="$1"
    echo "$json" | jq . >/dev/null 2>&1
}

# 1. Check if process is running
log "Checking process status..."
if pgrep -x "fluentum" > /dev/null; then
    echo -e "${GREEN}✅ Process Status: Running${NC}"
    log "Process status: Running"
else
    echo -e "${RED}❌ Process Status: Not Running${NC}"
    log "Process status: Not Running"
    exit 1
fi

# 2. Check RPC endpoint accessibility
log "Checking RPC endpoint accessibility..."
RPC_STATUS=$(rpc_call "/status")
if is_valid_json "$RPC_STATUS" && [ "$(echo "$RPC_STATUS" | jq -r '.result // "null"')" != "null" ]; then
    echo -e "${GREEN}✅ RPC Endpoint: Accessible${NC}"
    log "RPC endpoint: Accessible"
else
    echo -e "${RED}❌ RPC Endpoint: Not Accessible${NC}"
    log "RPC endpoint: Not Accessible"
    exit 1
fi

# 3. Check sync status
log "Checking sync status..."
SYNC_INFO=$(echo "$RPC_STATUS" | jq -r '.result.sync_info.catching_up // "unknown"')
if [ "$SYNC_INFO" = "false" ]; then
    echo -e "${GREEN}✅ Sync Status: Caught Up${NC}"
    log "Sync status: Caught Up"
elif [ "$SYNC_INFO" = "true" ]; then
    echo -e "${YELLOW}⚠️  Sync Status: Catching Up${NC}"
    log "Sync status: Catching Up"
else
    echo -e "${RED}❌ Sync Status: Unknown${NC}"
    log "Sync status: Unknown"
fi

# 4. Check latest block
log "Checking latest block..."
LATEST_BLOCK=$(echo "$RPC_STATUS" | jq -r '.result.sync_info.latest_block_height // "N/A"')
echo -e "${GREEN}📦 Latest Block: $LATEST_BLOCK${NC}"
log "Latest block: $LATEST_BLOCK"

# 5. Check network information
log "Checking network information..."
NET_INFO=$(rpc_call "/net_info")
if is_valid_json "$NET_INFO"; then
    PEER_COUNT=$(echo "$NET_INFO" | jq -r '.result.n_peers // 0')
    echo -e "${GREEN}🌐 Connected Peers: $PEER_COUNT${NC}"
    log "Connected peers: $PEER_COUNT"
    
    # Check if we have enough peers
    if [ "$PEER_COUNT" -lt 2 ]; then
        echo -e "${YELLOW}⚠️  Warning: Low peer count (less than 2)${NC}"
        log "Warning: Low peer count ($PEER_COUNT)"
    fi
else
    echo -e "${RED}❌ Network Info: Not Available${NC}"
    log "Network info: Not Available"
fi

# 6. Check validator information
log "Checking validator information..."
VALIDATORS=$(rpc_call "/validators")
if is_valid_json "$VALIDATORS"; then
    VALIDATOR_COUNT=$(echo "$VALIDATORS" | jq -r '.result.validators | length // 0')
    echo -e "${GREEN}👥 Validator Set Size: $VALIDATOR_COUNT${NC}"
    log "Validator set size: $VALIDATOR_COUNT"
else
    echo -e "${RED}❌ Validator Info: Not Available${NC}"
    log "Validator info: Not Available"
fi

# 7. Check consensus state
log "Checking consensus state..."
CONSENSUS_STATE=$(rpc_call "/consensus_state")
if is_valid_json "$CONSENSUS_STATE"; then
    ROUND_STATE=$(echo "$CONSENSUS_STATE" | jq -r '.result.round_state.step // "unknown"')
    echo -e "${GREEN}🔄 Consensus Round State: $ROUND_STATE${NC}"
    log "Consensus round state: $ROUND_STATE"
else
    echo -e "${YELLOW}⚠️  Consensus State: Not Available${NC}"
    log "Consensus state: Not Available"
fi

# 8. Check system resources
log "Checking system resources..."
if command -v free &> /dev/null; then
    MEMORY_USAGE=$(free | grep Mem | awk '{printf("%.1f%%", $3/$2 * 100.0)}')
    echo -e "${GREEN}💾 Memory Usage: $MEMORY_USAGE${NC}"
    log "Memory usage: $MEMORY_USAGE"
fi

if command -v df &> /dev/null; then
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}')
    echo -e "${GREEN}💿 Disk Usage: $DISK_USAGE${NC}"
    log "Disk usage: $DISK_USAGE"
fi

# 9. Check for recent errors in logs
log "Checking for recent errors..."
if [ -f "$HOME/.fluentum/logs/tendermint.log" ]; then
    RECENT_ERRORS=$(tail -100 "$HOME/.fluentum/logs/tendermint.log" | grep -i "error" | wc -l)
    if [ "$RECENT_ERRORS" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  Recent Errors: $RECENT_ERRORS in last 100 lines${NC}"
        log "Recent errors: $RECENT_ERRORS in last 100 lines"
    else
        echo -e "${GREEN}✅ Recent Errors: None detected${NC}"
        log "Recent errors: None detected"
    fi
else
    echo -e "${YELLOW}⚠️  Log File: Not found${NC}"
    log "Log file: Not found"
fi

# 10. Check uptime
log "Checking node uptime..."
if pgrep -x "fluentum" > /dev/null; then
    UPTIME=$(ps -o etime= -p $(pgrep -x "fluentum") | xargs)
    echo -e "${GREEN}⏱️  Node Uptime: $UPTIME${NC}"
    log "Node uptime: $UPTIME"
fi

# 11. Performance metrics
log "Checking performance metrics..."
if [ -f "/proc/$(pgrep -x fluentum)/stat" ]; then
    CPU_TIME=$(cat "/proc/$(pgrep -x fluentum)/stat" | awk '{print $14+$15}')
    echo -e "${GREEN}⚡ CPU Time: ${CPU_TIME}ms${NC}"
    log "CPU time: ${CPU_TIME}ms"
fi

# Summary and recommendations
echo ""
echo "${BLUE}============================${NC}"
echo "${BLUE}Health Check Summary${NC}"
echo "${BLUE}============================${NC}"

# Determine overall health status
HEALTH_STATUS="HEALTHY"
if [ "$SYNC_INFO" = "true" ]; then
    HEALTH_STATUS="SYNCING"
fi

if [ "$PEER_COUNT" -lt 2 ]; then
    HEALTH_STATUS="WARNING"
fi

case $HEALTH_STATUS in
    "HEALTHY")
        echo -e "${GREEN}🎉 Overall Status: HEALTHY${NC}"
        log "Overall status: HEALTHY"
        ;;
    "SYNCING")
        echo -e "${YELLOW}⚠️  Overall Status: SYNCING${NC}"
        log "Overall status: SYNCING"
        ;;
    "WARNING")
        echo -e "${YELLOW}⚠️  Overall Status: WARNING${NC}"
        log "Overall status: WARNING"
        ;;
    *)
        echo -e "${RED}❌ Overall Status: UNHEALTHY${NC}"
        log "Overall status: UNHEALTHY"
        ;;
esac

# Recommendations
echo ""
echo "${BLUE}Recommendations:${NC}"
if [ "$SYNC_INFO" = "true" ]; then
    echo -e "${YELLOW}• Node is still syncing, monitor progress${NC}"
fi

if [ "$PEER_COUNT" -lt 2 ]; then
    echo -e "${YELLOW}• Low peer count, check network connectivity${NC}"
fi

if [ "$RECENT_ERRORS" -gt 5 ]; then
    echo -e "${YELLOW}• High error rate, check logs for issues${NC}"
fi

echo ""
log "Health check completed"
echo -e "${GREEN}✅ Health check completed!${NC}" 