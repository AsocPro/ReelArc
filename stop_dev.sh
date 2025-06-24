#!/bin/bash

# Stop the development servers

# Define the env file for PIDs
PID_ENV_FILE=".dev.env"

# Check if the PID file exists
if [ ! -f "$PID_ENV_FILE" ]; then
    echo "Error: PID file $PID_ENV_FILE not found."
    echo "Are the development servers running? If not, use ./start_dev.sh to start them."
else
    # Source the PID file to get the PIDs
    source "$PID_ENV_FILE"
    
    # Check if the frontend process is running
    if ps -p $FRONTEND_PID > /dev/null; then
        echo "Stopping frontend server (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID
        echo "Frontend server stopped."
    else
        echo "Frontend server (PID: $FRONTEND_PID) is not running."
    fi
    
    # Check if the backend process is running
    if ps -p $BACKEND_PID > /dev/null; then
        echo "Stopping backend server (PID: $BACKEND_PID)..."
        kill $BACKEND_PID
        echo "Backend server stopped."
    else
        echo "Backend server (PID: $BACKEND_PID) is not running."
    fi
fi

# Try to find and kill any process using port 8080
PORT_PID=$(lsof -i :8080 | tail -1 | awk '{print $2}')
if [ ! -z "$PORT_PID" ]; then
echo "Found process using port 8080 (PID: $PORT_PID). Killing it..."
kill -9 $PORT_PID
echo "Process on port 8080 stopped."
fi
# Remove the PID file
rm -f "$PID_ENV_FILE"
echo "Development servers stopped and PID file removed."
