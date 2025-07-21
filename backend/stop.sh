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

"$TAILSCALE_PATH" funnel off

PORT=8080
PID=$(sudo lsof -t -i TCP:$PORT)

if [ -z "$PID" ]; then
  echo "No process is using port $PORT"
else
  echo "Killing process $PID using port $PORT..."
  sudo kill -9 $PID
  echo "Process killed."
fi