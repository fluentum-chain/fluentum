#!/bin/bash
# Deploy the canonical genesis.json to all node config directories

[ "$DEBUG" = "1" ] && set -x
set -e

now() { date '+%Y-%m-%d %H:%M:%S'; }
print_status() { echo -e "$(now) [INFO] $1"; }
print_error() { echo -e "$(now) [ERROR] $1" >&2; }

source "$(dirname "$0")/nodes.conf"

GENESIS_SRC="config/genesis.json"
FAILED=()

for node_name in "${VALID_NODES[@]}"; do
    for base in "/opt/fluentum" "/tmp"; do
        config_dir="$base/$node_name/config"
        dest="$config_dir/genesis.json"
        if mkdir -p "$config_dir" && cp "$GENESIS_SRC" "$dest"; then
            print_status "Copied genesis.json to $dest"
        else
            print_error "Failed to copy genesis.json to $dest"
            FAILED+=("$dest")
        fi
    done
done

if [ ${#FAILED[@]} -ne 0 ]; then
    print_error "Failed to copy genesis.json to the following locations: ${FAILED[*]}"
else
    print_status "genesis.json deployed to all node config directories successfully."
fi 