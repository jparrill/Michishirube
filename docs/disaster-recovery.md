# Disaster Recovery - Database Backup and Restore

This document describes the backup and restore procedures for the Michishirube SQLite database to ensure data protection and recovery capabilities.

## Overview

Michishirube provides automated scripts for database backup and restore operations. These scripts handle both compressed and uncompressed backups, perform integrity checks, and provide multiple operational modes for different scenarios.

## Scripts Location

The backup and restore scripts are located in the `scripts/` directory:
- `scripts/backup.sh` - Database backup script
- `scripts/restore.sh` - Database restore script

Both scripts are executable and include comprehensive help documentation.

## Database Backup

### Quick Start

```bash
# Basic backup (recommended for regular use)
./scripts/backup.sh

# Compressed backup (saves space)
./scripts/backup.sh --compress

# View all backup options
./scripts/backup.sh --help
```

### Backup Script Features

- **Automatic timestamping**: Backups are named with `YYYYMMDD_HHMMSS` format
- **Database optimization**: Performs SQLite VACUUM before backup
- **Compression support**: Optional gzip compression to reduce file size
- **Custom paths**: Support for custom database and backup directory paths
- **Integrity verification**: Uses SQLite to ensure database consistency
- **Progress reporting**: Colored output with detailed status information

### Backup Options

| Option | Description | Example |
|--------|-------------|---------|
| `-h, --help` | Show help message | `./scripts/backup.sh --help` |
| `-d, --database PATH` | Custom database file path | `./scripts/backup.sh -d /custom/path/db.sqlite` |
| `-b, --backup-dir DIR` | Custom backup directory | `./scripts/backup.sh -b /backup/location` |
| `-c, --compress` | Compress backup with gzip | `./scripts/backup.sh --compress` |

### Environment Variables

```bash
# Override default database path
export DB_PATH="/custom/path/michishirube.db"

# Override default backup directory
export BACKUP_DIR="/backup/storage"

# Run backup with custom settings
./scripts/backup.sh
```

### Backup Examples

```bash
# Daily backup routine
./scripts/backup.sh --compress

# Backup before major changes
./scripts/backup.sh -b ./pre-migration-backups

# Backup custom database location
./scripts/backup.sh -d /var/data/michishirube.db -b /backups/monthly
```

## Database Restore

### Quick Start

```bash
# Interactive restore (choose from list)
./scripts/restore.sh

# Restore latest backup automatically
./scripts/restore.sh --latest

# List available backups
./scripts/restore.sh --list

# Restore specific backup
./scripts/restore.sh /path/to/backup.db
```

### Restore Script Features

- **Interactive mode**: Browse and select from available backups
- **Latest backup**: Automatically restore the most recent backup
- **Safety backup**: Creates backup of current database before restore
- **Compression support**: Automatically handles compressed backups
- **Integrity verification**: Validates restored database integrity
- **Force mode**: Skip confirmation prompts for automation

### Restore Options

| Option | Description | Example |
|--------|-------------|---------|
| `-h, --help` | Show help message | `./scripts/restore.sh --help` |
| `-d, --database PATH` | Target database file path | `./scripts/restore.sh -d /custom/path/db.sqlite` |
| `-l, --list` | List available backup files | `./scripts/restore.sh --list` |
| `-f, --force` | Force restore without confirmation | `./scripts/restore.sh --force backup.db` |
| `--latest` | Restore from latest backup | `./scripts/restore.sh --latest` |

### Restore Examples

```bash
# Emergency restore from latest backup
./scripts/restore.sh --latest --force

# Restore specific backup with confirmation
./scripts/restore.sh ./backups/michishirube_backup_20240101_120000.db

# Restore to custom location
./scripts/restore.sh -d /new/location/db.sqlite backup.db.gz
```

## Backup Strategy Recommendations

### Daily Backups

Set up automated daily backups using cron:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM (compressed)
0 2 * * * cd /path/to/michishirube && ./scripts/backup.sh --compress

# Add weekly cleanup (keep last 30 backups)
0 3 * * 0 cd /path/to/michishirube && find ./backups -name "*.db*" -mtime +30 -delete
```

### Pre-Migration Backups

Before major updates or migrations:

```bash
# Create specific backup directory
mkdir -p ./backups/pre-migration-$(date +%Y%m%d)

# Create backup
./scripts/backup.sh -b ./backups/pre-migration-$(date +%Y%m%d)
```

### Development Backups

For development environments:

```bash
# Backup development database
DB_PATH="./michishirube_dev_data/michishirube-dev.db" ./scripts/backup.sh

# Restore development database
DB_PATH="./michishirube_dev_data/michishirube-dev.db" ./scripts/restore.sh --latest
```

## File Organization

### Default Directory Structure

```
michishirube/
├── backups/                           # Backup storage directory
│   ├── michishirube_backup_20240101_120000.db
│   ├── michishirube_backup_20240101_120000.db.gz
│   └── ...
├── michishirube_data/                 # Production database
│   ├── michishirube.db
│   └── michishirube.db.backup.TIMESTAMP  # Auto-created safety backups
├── michishirube_dev_data/             # Development database
│   └── michishirube-dev.db
└── scripts/
    ├── backup.sh
    └── restore.sh
```

### Backup Naming Convention

Backup files follow this naming pattern:
- Uncompressed: `michishirube_backup_YYYYMMDD_HHMMSS.db`
- Compressed: `michishirube_backup_YYYYMMDD_HHMMSS.db.gz`

Safety backups (created during restore):
- `michishirube.db.backup.YYYYMMDD_HHMMSS`

## Docker Integration

### Backup Container Database

When using Docker, the scripts automatically work with the mounted volume:

```bash
# Backup containerized database
./scripts/backup.sh

# Restore containerized database (stops container temporarily)
docker-compose down
./scripts/restore.sh --latest
docker-compose up -d
```

### Container Volume Backup

For complete container volume backup:

```bash
# Create volume backup
docker run --rm \
  -v michishirube_data:/data:ro \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/volume_backup_$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
```

## Troubleshooting

### Common Issues

#### "Database file not found"
```bash
# Check database location
ls -la ./michishirube_data/michishirube.db

# Start application to create database
docker-compose up -d
```

#### "Permission denied"
```bash
# Make scripts executable
chmod +x scripts/backup.sh scripts/restore.sh

# Check directory permissions
ls -la scripts/
```

#### "SQLite command not found"
```bash
# Install SQLite (macOS)
brew install sqlite

# Install SQLite (Ubuntu/Debian)
sudo apt-get install sqlite3

# Install SQLite (CentOS/RHEL)
sudo yum install sqlite
```

#### "Backup directory not accessible"
```bash
# Create backup directory
mkdir -p ./backups

# Check permissions
ls -la ./backups/
```

### Recovery Scenarios

#### Complete Database Loss

1. **Stop the application**:
   ```bash
   docker-compose down
   ```

2. **Restore from latest backup**:
   ```bash
   ./scripts/restore.sh --latest --force
   ```

3. **Restart application**:
   ```bash
   docker-compose up -d
   ```

#### Corrupted Database

1. **Verify corruption**:
   ```bash
   sqlite3 ./michishirube_data/michishirube.db "PRAGMA integrity_check;"
   ```

2. **Restore from backup**:
   ```bash
   ./scripts/restore.sh --latest
   ```

#### Accidental Data Deletion

1. **Stop application immediately**:
   ```bash
   docker-compose down
   ```

2. **Restore from recent backup**:
   ```bash
   ./scripts/restore.sh --list
   ./scripts/restore.sh [selected_backup]
   ```

## Security Considerations

### Backup Security

- **Encryption**: Consider encrypting backups for sensitive data
- **Access Control**: Restrict backup directory permissions
- **Retention**: Implement backup retention policies
- **Storage**: Store backups in multiple locations

### Example Encrypted Backup

```bash
# Create encrypted backup
./scripts/backup.sh --compress
gpg --symmetric --cipher-algo AES256 ./backups/latest_backup.db.gz

# Restore encrypted backup
gpg --decrypt backup.db.gz.gpg | gunzip > ./michishirube_data/michishirube.db
```

## Automation Scripts

### Backup Automation Example

Create `scripts/automated-backup.sh`:

```bash
#!/bin/bash
# Automated backup with retention

BACKUP_RETENTION_DAYS=30
PROJECT_DIR="/path/to/michishirube"

cd "$PROJECT_DIR"

# Create backup
./scripts/backup.sh --compress

# Clean old backups
find ./backups -name "michishirube_backup_*.db*" -mtime +$BACKUP_RETENTION_DAYS -delete

# Log backup status
echo "$(date): Backup completed, old backups cleaned" >> ./logs/backup.log
```

### Health Check Script

Create `scripts/db-health-check.sh`:

```bash
#!/bin/bash
# Database health check

DB_PATH="./michishirube_data/michishirube.db"

if sqlite3 "$DB_PATH" "PRAGMA integrity_check;" | grep -q "ok"; then
    echo "Database integrity: OK"
    exit 0
else
    echo "Database integrity: FAILED"
    echo "Consider restoring from backup"
    exit 1
fi
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Database Backup
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Create Backup
        run: |
          ./scripts/backup.sh --compress
      - name: Upload Backup
        uses: actions/upload-artifact@v3
        with:
          name: database-backup
          path: backups/
```

## Monitoring and Alerting

### Backup Monitoring

Monitor backup operations by checking:
- Backup file creation timestamps
- Backup file sizes
- Backup directory disk usage
- Script execution logs

### Example Monitoring Script

```bash
#!/bin/bash
# Check last backup age

LAST_BACKUP=$(find ./backups -name "michishirube_backup_*.db*" -type f | sort | tail -1)
BACKUP_AGE=$(( ($(date +%s) - $(stat -f %m "$LAST_BACKUP")) / 86400 ))

if [ $BACKUP_AGE -gt 1 ]; then
    echo "WARNING: Last backup is $BACKUP_AGE days old"
    exit 1
else
    echo "OK: Recent backup found ($BACKUP_AGE days old)"
    exit 0
fi
```

## Best Practices

1. **Regular Testing**: Periodically test restore procedures
2. **Multiple Locations**: Store backups in different physical locations
3. **Automated Verification**: Implement automated backup integrity checks
4. **Documentation**: Keep recovery procedures documented and accessible
5. **Monitoring**: Set up alerts for backup failures
6. **Retention Policy**: Implement appropriate backup retention periods
7. **Version Control**: Keep backup scripts in version control
8. **Access Control**: Limit access to backup files and scripts

## Contact and Support

For issues related to backup and restore operations:
1. Check this documentation for common solutions
2. Review script help output (`--help` option)
3. Check application logs for error details
4. Verify file permissions and disk space
5. Test scripts in development environment first

Remember: Regular backups are essential, but regular restore testing is equally important to ensure your disaster recovery procedures work when needed.