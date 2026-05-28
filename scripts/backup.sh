#!/bin/bash
set -e

BACKUP_DIR="/backup/beep-analytics"
DB_PATH="/var/lib/beep-analytics/beep-analytics.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Create backup
	sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/beep-analytics-$DATE.db'"

# Compress
gzip "$BACKUP_DIR/beep-analytics-$DATE.db"

# Remove backups older than 30 days
find "$BACKUP_DIR" -name "beep-analytics-*.db.gz" -mtime +30 -delete

echo "Backup completed: beep-analytics-$DATE.db.gz"
