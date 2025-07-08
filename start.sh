#!/bin/bash

# Function to check if port 3000 is ready
wait_for_frontend() {
  while ! nc -z localhost 3000; do   
    sleep 1
  done
}

# Function to open URL in the default browser
open_browser() {
  local url=$1
  case "$(uname -s)" in
    Darwin*)  open "$url";;              # macOS
    Linux*)   xdg-open "$url";;         # Linux
    CYGWIN*|MINGW*|MSYS*)  start "$url";; # Windows
    *)        echo "Unknown OS: Please open $url manually";;
  esac
}

# Function to start a new terminal
start_in_new_terminal() {
  local command=$1
  case "$(uname -s)" in
    Darwin*)  # macOS
      osascript -e "tell application \"Terminal\" to do script \"$command\""
      ;;
    Linux*)   # Linux
      if command -v gnome-terminal &> /dev/null; then
        gnome-terminal -- bash -c "$command"
      elif command -v xterm &> /dev/null; then
        xterm -e "$command" &
      else
        echo "No supported terminal emulator found"
        exit 1
      fi
      ;;
    CYGWIN*|MINGW*|MSYS*)  # Windows
      start cmd /c "$command"
      ;;
    *)
      echo "Unknown OS: Cannot start new terminal"
      exit 1
      ;;
  esac
}

# Start backend server in a new terminal window
start_in_new_terminal "cd '$(pwd)'/backend && go run main.go"

# Start frontend server in a new terminal window
start_in_new_terminal "cd '$(pwd)'/frontend && npm run dev"

# Wait for frontend server to be ready
echo "Waiting for frontend server to start..."
wait_for_frontend

# Open frontend in browser
open_browser "http://localhost:3000"
