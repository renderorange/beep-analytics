#!/bin/bash
set -e

BACKUP_DIR="/backup/beep"
DB_PATH="/var/lib/beep/beep.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Create backup
sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/beep-$DATE.db'"

# Compress
gzip "$BACKUP_DIR/beep-$DATE.db"

# Remove backups older than 30 days
find "$BACKUP_DIR" -name "beep-*.db.gz" -mtime +30 -delete

echo "Backup completed: beep-$DATE.db.gz"
