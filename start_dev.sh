#!/bin/bash

# Start the frontend and backend in development mode

# Check if bun is installed
if ! command -v bun &> /dev/null; then
    echo "Error: bun is not installed. Please install it first."
    echo "Visit https://bun.sh/ for installation instructions."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install it first."
    echo "Visit https://golang.org/doc/install for installation instructions."
    exit 1
fi

# Create necessary directories if they don't exist
mkdir -p data/media data/metadata

# Define the env file for PIDs
PID_ENV_FILE=".dev.env"

# Check if the backend is already running
if [ -f "$PID_ENV_FILE" ]; then
    source "$PID_ENV_FILE"
    
    # Check if the backend process is running
    if ps -p $BACKEND_PID > /dev/null 2>&1; then
        echo "Backend server is already running (PID: $BACKEND_PID). Stopping it..."
        kill $BACKEND_PID
        echo "Backend server stopped. Restarting..."
        # Wait for the process to terminate
        sleep 2
    fi
fi

# Start the frontend in the background
echo "Starting frontend development server..."
cd client
bun run dev --host > ../frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..

# Wait a moment for the frontend to start
sleep 2

# Start the backend in the background
echo "Starting backend server..."
# First, check if anything is using port 8080 and kill it
PORT_PID=$(lsof -ti :8080)
if [ ! -z "$PORT_PID" ]; then
    echo "Found process using port 8080 (PID: $PORT_PID). Killing it..."
    kill -9 $PORT_PID
    sleep 1
fi

cd server
go run main.go transcription.go > ../backend.log 2>&1 &
# Get the actual Go process PID, not the shell PID
sleep 2
BACKEND_PID=$(pgrep -f "go run main.go")
# If pgrep fails, try to find by port
if [ -z "$BACKEND_PID" ]; then
    BACKEND_PID=$(lsof -ti :8080)
fi
cd ..

# Save PIDs to env file
echo "FRONTEND_PID=$FRONTEND_PID" > $PID_ENV_FILE
echo "BACKEND_PID=$BACKEND_PID" >> $PID_ENV_FILE

echo "Development servers started."
echo "Frontend PID: $FRONTEND_PID"
echo "Backend PID: $BACKEND_PID"
echo "PIDs saved to $PID_ENV_FILE"
echo "Use ./stop_dev.sh to stop the servers."
