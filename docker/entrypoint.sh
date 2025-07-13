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

# Set COOKIE_SECURE based on mode
if [ "$CLOUD_MODE" = "true" ]; then
    # Cloud mode MUST use secure cookies for security
    export COOKIE_SECURE="true"
    echo "Cloud mode: Enforcing COOKIE_SECURE=true"
else
    # Regular mode: respect user settings, default to false if not set
    if [ -z "$COOKIE_SECURE" ]; then
        export COOKIE_SECURE="false"
    fi
fi

# Validate required settings for cloud mode
if [ "$CLOUD_MODE" = "true" ]; then
    missing_credentials=""
    if [ -z "$GOOGLE_CLIENT_ID" ]; then
        missing_credentials="GOOGLE_CLIENT_ID"
    fi
    if [ -z "$GOOGLE_CLIENT_SECRET" ]; then
        if [ -n "$missing_credentials" ]; then
            missing_credentials="$missing_credentials, GOOGLE_CLIENT_SECRET"
        else
            missing_credentials="GOOGLE_CLIENT_SECRET"
        fi
    fi
    if [ -n "$missing_credentials" ]; then
        echo "ERROR: Missing required Google OAuth credentials for cloud mode: $missing_credentials"
        exit 1
    fi
fi

echo "Starting Vega AI..."
echo "Mode: $([ "$CLOUD_MODE" = "true" ] && echo "Cloud" || echo "Self-hosted")"

exec "$@"
