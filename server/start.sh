#!/bin/bash

# All Kanye Backend - Quick Start Script

echo "======================================"
echo "All Kanye Backend Server"
echo "======================================"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found"
    exit 1
fi

echo "✅ Environment file found"
echo ""

# Check database connection
echo "🔍 Testing database connection..."
DB_HOST=$(grep DB_HOST .env | cut -d '=' -f2)
DB_PORT=$(grep DB_PORT .env | cut -d '=' -f2)
DB_USER=$(grep DB_USER .env | cut -d '=' -f2)
DB_PASSWORD=$(grep DB_PASSWORD .env | cut -d '=' -f2)
DB_NAME=$(grep DB_NAME .env | cut -d '=' -f2)

if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME" 2>/dev/null; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    echo "Please check your database configuration"
    exit 1
fi

echo ""
echo "🚀 Starting server..."
echo "Press Ctrl+C to stop"
echo ""

# Start the server
go run ./cmd/server
