#!/bin/bash

# Fix line endings for deployment scripts
echo "Fixing line endings for deployment scripts..."

# Fix the main deployment script
if [ -f "scripts/deploy_all_nodes.sh" ]; then
    echo "Fixing deploy_all_nodes.sh..."
    dos2unix scripts/deploy_all_nodes.sh 2>/dev/null || sed -i 's/\r$//' scripts/deploy_all_nodes.sh
    chmod +x scripts/deploy_all_nodes.sh
    echo "Fixed deploy_all_nodes.sh"
fi

# Fix other scripts
for script in scripts/*.sh; do
    if [ -f "$script" ]; then
        echo "Fixing $script..."
        dos2unix "$script" 2>/dev/null || sed -i 's/\r$//' "$script"
        chmod +x "$script"
    fi
done

echo "All scripts fixed!" 