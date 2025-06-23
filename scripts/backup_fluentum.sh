#!/bin/bash
# Fluentum Backup Script
# This script creates backups of critical Fluentum files

set -e

# Configuration
BACKUP_DIR="/backup/fluentum"
DATE=$(date +%Y%m%d_%H%M%S)
FLUENTUM_HOME="$HOME/.fluentum"
LOG_FILE="/var/log/fluentum_backup.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check if Fluentum home directory exists
if [ ! -d "$FLUENTUM_HOME" ]; then
    error_exit "Fluentum home directory not found: $FLUENTUM_HOME"
fi

# Create backup directory
log "${YELLOW}Creating backup directory: $BACKUP_DIR${NC}"
mkdir -p "$BACKUP_DIR" || error_exit "Failed to create backup directory"

log "${GREEN}🔄 Creating Fluentum backup at $DATE${NC}"

# Function to backup file with error handling
backup_file() {
    local source="$1"
    local dest="$2"
    local description="$3"
    
    if [ -f "$source" ]; then
        cp "$source" "$dest" || error_exit "Failed to backup $description"
        log "${GREEN}✅ Backed up $description${NC}"
    else
        log "${YELLOW}⚠️  Warning: $description not found at $source${NC}"
    fi
}

# Backup critical files
backup_file "$FLUENTUM_HOME/config/genesis.json" "$BACKUP_DIR/genesis.json.$DATE" "genesis file"
backup_file "$FLUENTUM_HOME/config/priv_validator_key.json" "$BACKUP_DIR/priv_validator_key.json.$DATE" "private validator key"
backup_file "$FLUENTUM_HOME/config/node_key.json" "$BACKUP_DIR/node_key.json.$DATE" "node key"
backup_file "$FLUENTUM_HOME/data/priv_validator_state.json" "$BACKUP_DIR/priv_validator_state.json.$DATE" "validator state"
backup_file "$FLUENTUM_HOME/config/config.toml" "$BACKUP_DIR/config.toml.$DATE" "configuration file"

# Set proper permissions for sensitive files
log "${YELLOW}Setting proper permissions for sensitive files${NC}"
chmod 600 "$BACKUP_DIR"/*.json.$DATE 2>/dev/null || log "${YELLOW}Warning: Could not set permissions on some files${NC}"

# Create checksums for integrity verification
log "${YELLOW}Creating checksums for integrity verification${NC}"
cd "$BACKUP_DIR" || error_exit "Failed to change to backup directory"
sha256sum *.$DATE > "checksums.$DATE" || error_exit "Failed to create checksums"

# Create backup manifest
cat > "manifest.$DATE" << EOF
Fluentum Backup Manifest
========================
Backup Date: $(date)
Backup Directory: $BACKUP_DIR
Fluentum Home: $FLUENTUM_HOME

Files Backed Up:
$(ls -la *.$DATE | grep -v checksums | grep -v manifest)

Checksums:
$(cat checksums.$DATE)

Backup completed successfully.
EOF

# Clean old backups (keep last 7 days)
log "${YELLOW}Cleaning old backups (keeping last 7 days)${NC}"
find "$BACKUP_DIR" -name "*.$(date -d '7 days ago' +%Y%m%d)*" -delete 2>/dev/null || log "${YELLOW}Warning: Could not clean old backups${NC}"

# Calculate backup size
BACKUP_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/*.$DATE 2>/dev/null | wc -l)

log "${GREEN}✅ Backup completed successfully${NC}"
log "${GREEN}📊 Backup location: $BACKUP_DIR${NC}"
log "${GREEN}📊 Backup size: $BACKUP_SIZE${NC}"
log "${GREEN}📊 Files backed up: $BACKUP_COUNT${NC}"
log "${GREEN}📊 Manifest: $BACKUP_DIR/manifest.$DATE${NC}"
log "${GREEN}📊 Checksums: $BACKUP_DIR/checksums.$DATE${NC}"

# Verify backup integrity
log "${YELLOW}Verifying backup integrity...${NC}"
if sha256sum -c "checksums.$DATE" >/dev/null 2>&1; then
    log "${GREEN}✅ Backup integrity verified${NC}"
else
    log "${RED}❌ Backup integrity check failed${NC}"
    exit 1
fi

log "${GREEN}🎉 Fluentum backup completed successfully!${NC}" 