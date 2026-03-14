#!/bin/bash

# Exit on error
set -e

echo "Building Atoman locally..."

# Build Backend
echo "Building backend..."
cd server
# Cross-compiling for Linux Arm64 (Alpine compatible)
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0
go mod tidy
go build -o main ./cmd/start_server
cd ..

# Build Frontend
echo "Building frontend..."
cd web
# Using npm or bun depending on what you have installed
if command -v bun &> /dev/null; then
    bun install
    bun run build
else
    npm install
    npm run build
fi
cd ..

echo "Build completed locally!"
echo "Now you can run: docker-compose up --build -d"
