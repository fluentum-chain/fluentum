#!/bin/bash

echo "=== Checking Genesis File Content ==="
echo "File: /opt/fluentum/config/genesis.json"
echo ""

if [ -f "/opt/fluentum/config/genesis.json" ]; then
    echo "Genesis file exists. Content:"
    cat /opt/fluentum/config/genesis.json | jq .
else
    echo "Genesis file does not exist!"
fi

echo ""
echo "=== Checking for initial_height field ==="
if [ -f "/opt/fluentum/config/genesis.json" ]; then
    echo "initial_height value:"
    cat /opt/fluentum/config/genesis.json | jq -r '.initial_height'
    echo ""
    echo "initial_height type:"
    cat /opt/fluentum/config/genesis.json | jq -r '.initial_height | type'
fi 