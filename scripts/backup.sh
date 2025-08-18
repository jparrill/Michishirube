#!/bin/bash

# Michishirube Database Backup Script
# Creates timestamped backups of the SQLite database

set -e

# Configuration
DB_PATH="${DB_PATH:-./michishirube_data/michishirube.db}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/michishirube_backup_${TIMESTAMP}.db"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    echo "Michishirube Database Backup Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -d, --database PATH Path to database file (default: $DB_PATH)"
    echo "  -b, --backup-dir DIR Backup directory (default: $BACKUP_DIR)"
    echo "  -c, --compress      Compress the backup with gzip"
    echo ""
    echo "Environment variables:"
    echo "  DB_PATH             Database file path"
    echo "  BACKUP_DIR          Backup directory path"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Basic backup"
    echo "  $0 --compress                        # Compressed backup"
    echo "  $0 -d ./custom/path/db.sqlite        # Custom database path"
    echo "  $0 -b ./custom/backups               # Custom backup directory"
}

# Parse command line arguments
COMPRESS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--database)
            DB_PATH="$2"
            shift 2
            ;;
        -b|--backup-dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        -c|--compress)
            COMPRESS=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate database file exists
if [[ ! -f "$DB_PATH" ]]; then
    print_error "Database file not found: $DB_PATH"
    print_info "Make sure the database exists or use docker-compose to start the application first"
    exit 1
fi

# Create backup directory if it doesn't exist
if [[ ! -d "$BACKUP_DIR" ]]; then
    print_info "Creating backup directory: $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
fi

# Adjust backup file name if compression is enabled
if [[ "$COMPRESS" == true ]]; then
    BACKUP_FILE="${BACKUP_FILE}.gz"
fi

print_info "Starting backup process..."
print_info "Source database: $DB_PATH"
print_info "Backup file: $BACKUP_FILE"

# Check if SQLite is available for VACUUM operation
if command -v sqlite3 >/dev/null 2>&1; then
    print_info "Performing database vacuum before backup..."
    sqlite3 "$DB_PATH" "VACUUM;"
else
    print_warning "sqlite3 command not found, skipping vacuum operation"
fi

# Create backup
if [[ "$COMPRESS" == true ]]; then
    print_info "Creating compressed backup..."
    if gzip -c "$DB_PATH" > "$BACKUP_FILE"; then
        print_info "Compressed backup created successfully: $BACKUP_FILE"
    else
        print_error "Failed to create compressed backup"
        exit 1
    fi
else
    print_info "Creating backup..."
    if cp "$DB_PATH" "$BACKUP_FILE"; then
        print_info "Backup created successfully: $BACKUP_FILE"
    else
        print_error "Failed to create backup"
        exit 1
    fi
fi

# Show backup file size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
print_info "Backup size: $BACKUP_SIZE"

# Show number of existing backups
BACKUP_COUNT=$(ls -1 "${BACKUP_DIR}"/michishirube_backup_*.db* 2>/dev/null | wc -l)
print_info "Total backups in directory: $BACKUP_COUNT"

print_info "Backup completed successfully!"
print_info "To restore this backup, use: ./scripts/restore.sh \"$BACKUP_FILE\""