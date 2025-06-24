#!/bin/bash

# Build the frontend and start the backend in production mode

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

# Build the frontend
echo "Building frontend..."
cd client && bun run build
if [ $? -ne 0 ]; then
    echo "Frontend build failed!"
    exit 1
fi
echo "Frontend build completed successfully."

# Build the backend
echo "Building backend..."
cd ../server
go build -o timelineviewer
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi
echo "Backend build completed successfully."

# Start the backend
echo "Starting server in production mode..."
./timelineviewer