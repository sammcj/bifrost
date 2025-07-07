#!/bin/sh
set -e

# Function to fix permissions on mounted volumes
fix_permissions() {
    # Check if /app/data exists and fix ownership if needed
    if [ -d "/app/data" ]; then
        # Get current user info
        CURRENT_UID=$(id -u)
        CURRENT_GID=$(id -g)
        
        # Get directory ownership
        DATA_UID=$(stat -c %u /app/data 2>/dev/null || echo "0")
        DATA_GID=$(stat -c %g /app/data 2>/dev/null || echo "0")
        
        # If ownership doesn't match current user, try to fix it
        if [ "$DATA_UID" != "$CURRENT_UID" ] || [ "$DATA_GID" != "$CURRENT_GID" ]; then
            echo "Fixing permissions on /app/data (was $DATA_UID:$DATA_GID, setting to $CURRENT_UID:$CURRENT_GID)"
            
            # Try to change ownership (will work if running as root or if user has permission)
            if chown -R "$CURRENT_UID:$CURRENT_GID" /app/data 2>/dev/null; then
                echo "Successfully updated permissions on /app/data"
            else
                echo "Warning: Could not change ownership of /app/data. You may need to run:"
                echo "  docker run --user \$(id -u):\$(id -g) ..."
                echo "  or ensure the host directory is owned by UID:GID $CURRENT_UID:$CURRENT_GID"
            fi
        fi
        
        # Ensure logs subdirectory exists with correct permissions
        mkdir -p /app/data/logs
        chmod 755 /app/data/logs 2>/dev/null || true
    fi
}

# Fix permissions before starting the application
fix_permissions

# Execute the main application with all passed arguments
exec /app/main "$@" 