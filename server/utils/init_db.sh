#!/bin/bash

set -e

echo "========================================="
echo " Atoman - Database Initialization"
echo "========================================="
echo ""

# Check if database exists
if [ -f "database.sqlite" ]; then
    echo "⚠️  Existing database found: database.sqlite"
    echo ""
    echo "Options:"
    echo "  1) Backup and delete (create fresh database)"
    echo "  2) Keep existing database"
    echo "  3) Cancel"
    echo ""
    read -p "Choose [1-3]: " choice
    
    case $choice in
        1)
            BACKUP="database.sqlite.backup_$(date +%Y%m%d_%H%M%S)"
            cp database.sqlite "$BACKUP"
            echo "✅ Backed up to: $BACKUP"
            rm database.sqlite
            echo "✅ Deleted old database"
            ;;
        2)
            echo "ℹ️  Keeping existing database"
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo "Cancelled"
            exit 0
            ;;
    esac
fi

echo ""
echo "Creating new database..."
echo ""

# Create database by starting server briefly
timeout 5s go run ./cmd/server 2>&1 | grep -E "Database|Migration|S3" || true

echo ""
echo "Database created! Now creating admin user..."
echo ""

# Create admin user
cd cmd/create_admin
go run main.go <<EOF
admin
admin@atoman.com
admin123
EOF

echo ""
echo "========================================="
echo "✅ Database initialization complete!"
echo "========================================="
echo ""
echo "Admin credentials:"
echo "  Username: admin"
echo "  Email: admin@atoman.com"
echo "  Password: admin123"
echo ""
echo "⚠️  CHANGE THIS PASSWORD IN PRODUCTION!"
echo ""
echo "You can now start the server with: go run ./cmd/server"
echo ""
