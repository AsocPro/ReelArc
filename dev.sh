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

# Start the frontend in the background
echo "Starting frontend development server..."
cd client && bun run dev &
FRONTEND_PID=$!

# Wait a moment for the frontend to start
sleep 2

# Start the backend
echo "Starting backend server..."
cd server && go run main.go

# If the backend exits, kill the frontend
kill $FRONTEND_PID