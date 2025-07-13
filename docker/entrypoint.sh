#!/bin/sh
set -e

generate_token() {
    head -c 32 /dev/urandom | base64 | tr -d '\n'
}

# Generate TOKEN_SECRET if not provided
if [ -z "$TOKEN_SECRET" ] || [ "$TOKEN_SECRET" = "default-secret-key" ]; then
    if [ "$CLOUD_MODE" = "false" ]; then
        echo "=================================================="
        echo "WARNING: No TOKEN_SECRET provided."
        echo "Generating a random secret..."
        echo ""
        echo "For production use, please set a persistent"
        echo "TOKEN_SECRET environment variable."
        echo "=================================================="
        export TOKEN_SECRET=$(generate_token)
    else
        echo "ERROR: TOKEN_SECRET must be set for cloud deployments"
        exit 1
    fi
fi

# Validate required settings for cloud mode
if [ "$CLOUD_MODE" = "true" ]; then
    if [ -z "$GOOGLE_CLIENT_ID" ] || [ -z "$GOOGLE_CLIENT_SECRET" ]; then
        echo "ERROR: Google OAuth credentials required for cloud mode"
        exit 1
    fi
    # Ensure HTTPS cookies for cloud mode
    export COOKIE_SECURE="true"
fi

echo "Starting Vega AI..."
echo "Mode: $([ "$CLOUD_MODE" = "true" ] && echo "Cloud" || echo "Self-hosted")"

# Create data directory if it doesn't exist
mkdir -p /app/data

exec "$@"
