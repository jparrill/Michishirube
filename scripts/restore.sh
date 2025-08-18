#!/bin/bash

# Michishirube Database Restore Script
# Restores SQLite database from backup files

set -e

# Configuration
DB_PATH="${DB_PATH:-./michishirube_data/michishirube.db}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_prompt() {
    echo -e "${BLUE}[PROMPT]${NC} $1"
}

# Help function
show_help() {
    echo "Michishirube Database Restore Script"
    echo ""
    echo "Usage: $0 [OPTIONS] [BACKUP_FILE]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -d, --database PATH Target database file path (default: $DB_PATH)"
    echo "  -l, --list          List available backup files"
    echo "  -f, --force         Force restore without confirmation"
    echo "  --latest            Restore from the latest backup"
    echo ""
    echo "Arguments:"
    echo "  BACKUP_FILE         Path to backup file to restore from"
    echo ""
    echo "Environment variables:"
    echo "  DB_PATH             Target database file path"
    echo "  BACKUP_DIR          Backup directory path"
    echo ""
    echo "Examples:"
    echo "  $0                                           # Interactive mode - choose from list"
    echo "  $0 --latest                                  # Restore latest backup"
    echo "  $0 --list                                    # List available backups"
    echo "  $0 backup_file.db                           # Restore specific backup"
    echo "  $0 backup_file.db.gz                        # Restore compressed backup"
    echo "  $0 --force backup_file.db                   # Force restore without confirmation"
}

# Function to list available backups
list_backups() {
    print_info "Available backup files in $BACKUP_DIR:"
    echo ""
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        print_warning "Backup directory does not exist: $BACKUP_DIR"
        return 1
    fi
    
    local backups=($(find "$BACKUP_DIR" -name "michishirube_backup_*.db*" -type f 2>/dev/null | sort -r))
    
    if [[ ${#backups[@]} -eq 0 ]]; then
        print_warning "No backup files found in $BACKUP_DIR"
        return 1
    fi
    
    for i in "${!backups[@]}"; do
        local backup="${backups[$i]}"
        local size=$(du -h "$backup" | cut -f1)
        local date=$(date -r "$backup" "+%Y-%m-%d %H:%M:%S" 2>/dev/null || stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$backup" 2>/dev/null || echo "Unknown")
        printf "%2d) %s (%s, %s)\n" $((i+1)) "$(basename "$backup")" "$size" "$date"
    done
    
    echo ""
    return 0
}

# Function to get latest backup
get_latest_backup() {
    if [[ ! -d "$BACKUP_DIR" ]]; then
        print_error "Backup directory does not exist: $BACKUP_DIR"
        return 1
    fi
    
    local latest=$(find "$BACKUP_DIR" -name "michishirube_backup_*.db*" -type f 2>/dev/null | sort -r | head -n1)
    
    if [[ -z "$latest" ]]; then
        print_error "No backup files found in $BACKUP_DIR"
        return 1
    fi
    
    echo "$latest"
}

# Function to create database backup before restore
create_current_backup() {
    if [[ -f "$DB_PATH" ]]; then
        local current_backup="${DB_PATH}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Creating backup of current database: $current_backup"
        if cp "$DB_PATH" "$current_backup"; then
            print_info "Current database backed up to: $current_backup"
        else
            print_error "Failed to backup current database"
            return 1
        fi
    fi
}

# Function to restore database
restore_database() {
    local backup_file="$1"
    local target_db="$2"
    
    print_info "Restoring database from: $backup_file"
    print_info "Target database: $target_db"
    
    # Create target directory if it doesn't exist
    local target_dir=$(dirname "$target_db")
    if [[ ! -d "$target_dir" ]]; then
        print_info "Creating target directory: $target_dir"
        mkdir -p "$target_dir"
    fi
    
    # Check if backup file is compressed
    if [[ "$backup_file" == *.gz ]]; then
        print_info "Decompressing and restoring compressed backup..."
        if gunzip -c "$backup_file" > "$target_db"; then
            print_info "Compressed backup restored successfully"
        else
            print_error "Failed to restore compressed backup"
            return 1
        fi
    else
        print_info "Restoring uncompressed backup..."
        if cp "$backup_file" "$target_db"; then
            print_info "Backup restored successfully"
        else
            print_error "Failed to restore backup"
            return 1
        fi
    fi
    
    # Verify restored database
    if command -v sqlite3 >/dev/null 2>&1; then
        print_info "Verifying restored database integrity..."
        if sqlite3 "$target_db" "PRAGMA integrity_check;" | grep -q "ok"; then
            print_info "Database integrity check passed"
        else
            print_error "Database integrity check failed"
            return 1
        fi
    else
        print_warning "sqlite3 command not found, skipping integrity check"
    fi
}

# Parse command line arguments
FORCE=false
LIST_ONLY=false
USE_LATEST=false
BACKUP_FILE=""

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
        -l|--list)
            LIST_ONLY=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        --latest)
            USE_LATEST=true
            shift
            ;;
        -*)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
        *)
            BACKUP_FILE="$1"
            shift
            ;;
    esac
done

# Handle list-only mode
if [[ "$LIST_ONLY" == true ]]; then
    list_backups
    exit $?
fi

# Handle latest backup mode
if [[ "$USE_LATEST" == true ]]; then
    BACKUP_FILE=$(get_latest_backup)
    if [[ $? -ne 0 ]]; then
        exit 1
    fi
    print_info "Selected latest backup: $(basename "$BACKUP_FILE")"
fi

# Interactive mode if no backup file specified
if [[ -z "$BACKUP_FILE" ]]; then
    print_info "Interactive restore mode"
    
    if ! list_backups; then
        exit 1
    fi
    
    local backups=($(find "$BACKUP_DIR" -name "michishirube_backup_*.db*" -type f 2>/dev/null | sort -r))
    
    while true; do
        print_prompt "Select backup to restore (1-${#backups[@]}, or 'q' to quit): "
        read -r selection
        
        if [[ "$selection" == "q" ]]; then
            print_info "Restore cancelled"
            exit 0
        fi
        
        if [[ "$selection" =~ ^[0-9]+$ ]] && [[ "$selection" -ge 1 ]] && [[ "$selection" -le ${#backups[@]} ]]; then
            BACKUP_FILE="${backups[$((selection-1))]}"
            break
        else
            print_error "Invalid selection. Please enter a number between 1 and ${#backups[@]}, or 'q' to quit."
        fi
    done
fi

# Validate backup file exists
if [[ ! -f "$BACKUP_FILE" ]]; then
    print_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Confirmation prompt (unless forced)
if [[ "$FORCE" != true ]]; then
    echo ""
    print_warning "This will replace the current database with the backup."
    if [[ -f "$DB_PATH" ]]; then
        print_warning "Current database will be backed up automatically."
    fi
    print_prompt "Do you want to continue? (y/N): "
    read -r confirm
    
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        print_info "Restore cancelled"
        exit 0
    fi
fi

# Create backup of current database if it exists
if [[ -f "$DB_PATH" ]]; then
    create_current_backup
    if [[ $? -ne 0 ]]; then
        exit 1
    fi
fi

# Perform restore
restore_database "$BACKUP_FILE" "$DB_PATH"
if [[ $? -eq 0 ]]; then
    print_info "Database restore completed successfully!"
    print_info "Restored from: $(basename "$BACKUP_FILE")"
    print_info "Database location: $DB_PATH"
else
    print_error "Database restore failed!"
    exit 1
fi