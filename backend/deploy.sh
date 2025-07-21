#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Building the application..."
if ! go build -o async-multiplayer; then
    echo "Go build failed. Aborting deployment."
    exit 1
fi

echo "Build successful. Starting the application..."
LOG_FILE="deployment_$(date +%Y%m%d_%H%M%S).log"
./async-multiplayer > "$LOG_FILE" 2>&1 &

echo "Application started. Logs are being written to $LOG_FILE"

# Find tailscale executable
TAILSCALE_PATH=""
if command -v tailscale &> /dev/null; then
    TAILSCALE_PATH="$(command -v tailscale)"
elif [ -f "/Applications/Tailscale.app/Contents/MacOS/Tailscale" ]; then
    TAILSCALE_PATH="/Applications/Tailscale.app/Contents/MacOS/Tailscale"
fi

if [ -z "$TAILSCALE_PATH" ]; then
    echo "tailscale command not found. Please ensure it's installed and in your PATH."
    exit 1
fi

# Check if tailscale funnel is running, if not start it
echo "Checking Tailscale funnel status..."
STATUS_OUTPUT="$("$TAILSCALE_PATH" funnel status)"
echo "Tailscale status output: $STATUS_OUTPUT"

if ! echo "$STATUS_OUTPUT" | grep -q '(Funnel on)'; then
    echo "Starting tailscale funnel on port 8080..."
    if "$TAILSCALE_PATH" funnel --bg 8080; then
        echo "Tailscale funnel started successfully."
    else
        echo "Failed to start Tailscale funnel."
        exit 1
    fi
else
    echo "Tailscale funnel is already running."
fi
