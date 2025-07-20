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

# Check if tailscale funnel is running, if not start it
TAILSCALE_PATH="/Applications/Tailscale.app/Contents/MacOS/Tailscale"
if ! "$TAILSCALE_PATH" funnel status | grep '(Funnel on)'; then
    echo "Starting tailscale funnel on port 8080..."
    "$TAILSCALE_PATH" funnel --bg 8080
else
    echo "Tailscale funnel is already running."
fi
