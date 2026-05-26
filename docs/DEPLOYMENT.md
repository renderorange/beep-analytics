# Deployment Guide

## Prerequisites

- Go 1.26.3 or later
- SQLite support (via modernc.org/sqlite)
- Optional: GeoLite2-City database for geographic data

## Building from Source

### Standard Build

```bash
go build -o beep ./cmd/beep
```

### Production Build

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o beep ./cmd/beep
```

This creates a statically-linked binary suitable for minimal containers.

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o beep-linux-amd64 ./cmd/beep

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o beep-darwin-arm64 ./cmd/beep

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o beep.exe ./cmd/beep
```

## Systemd Service

### Service File

Create `/etc/systemd/system/beep.service`:

```ini
[Unit]
Description=beep analytics server
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=beep
Group=beep
WorkingDirectory=/opt/beep
ExecStart=/opt/beep/beep serve \
  --port 8080 \
  --db /var/lib/beep/tracker.db \
  --geoip /var/lib/beep/geoip
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ReadWritePaths=/var/lib/beep
ReadOnlyPaths=/opt/beep

[Install]
WantedBy=multi-user.target
```

### Installation Steps

```bash
# Create user and directories
sudo useradd -r -s /bin/false beep
sudo mkdir -p /opt/beep /var/lib/beep
sudo chown beep:beep /var/lib/beep

# Copy binary
sudo cp beep /opt/beep/
sudo chown beep:beep /opt/beep/beep

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable beep
sudo systemctl start beep

# Check status
sudo systemctl status beep
```

## Nginx Reverse Proxy

### Basic Configuration

```nginx
server {
    listen 80;
    server_name analytics.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name analytics.example.com;

    ssl_certificate /etc/letsencrypt/live/analytics.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/analytics.example.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Tracking endpoint
    location /collect {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # CORS headers for tracking script
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "POST, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type";

        if ($request_method = 'OPTIONS') {
            return 204;
        }
    }

    # Tracking script
    location /track.js {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Cache the tracking script
        expires 1h;
        add_header Cache-Control "public, immutable";
    }

    # API endpoints
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Restrict API access
        allow 127.0.0.1;
        allow your-admin-ip;
        deny all;
    }
}
```

### Rate Limiting

Add rate limiting to prevent abuse:

```nginx
http {
    limit_req_zone $binary_remote_addr zone=tracking:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=api:10m rate=1r/s;

    server {
        location /collect {
            limit_req zone=tracking burst=20 nodelay;
            # ... other config
        }

        location /api/ {
            limit_req zone=api burst=5 nodelay;
            # ... other config
        }
    }
}
```

## GeoIP Setup

### Download GeoLite2 Database

1. Sign up for a MaxMind account at https://www.maxmind.com/
2. Generate a license key
3. Download the GeoLite2-City CSV format

### Installation

```bash
# Create directory
sudo mkdir -p /var/lib/beep/geoip

# Download and extract (replace with your download method)
cd /tmp
wget "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=YOUR_KEY&suffix=zip" -O geolite2.zip
unzip geolite2.zip
sudo cp GeoLite2-City_*/GeoLite2-City-Blocks.csv /var/lib/beep/geoip/
sudo cp GeoLite2-City_*/GeoLite2-City-Locations-en.csv /var/lib/beep/geoip/

# Set permissions
sudo chown -R beep:beep /var/lib/beep/geoip
```

### Update Script

Create `/opt/beep/update-geoip.sh`:

```bash
#!/bin/bash
set -e

LICENSE_KEY="YOUR_LICENSE_KEY"
DOWNLOAD_URL="https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=${LICENSE_KEY}&suffix=zip"
TMP_DIR=$(mktemp -d)

# Download
wget "$DOWNLOAD_URL" -O "$TMP_DIR/geolite2.zip"

# Extract
unzip "$TMP_DIR/geolite2.zip" -d "$TMP_DIR"

# Find the extracted directory
EXTRACTED_DIR=$(find "$TMP_DIR" -name "GeoLite2-City_*" -type d | head -1)

# Stop service
sudo systemctl stop beep

# Update files
sudo cp "$EXTRACTED_DIR/GeoLite2-City-Blocks.csv" /var/lib/beep/geoip/
sudo cp "$EXTRACTED_DIR/GeoLite2-City-Locations-en.csv" /var/lib/beep/geoip/

# Start service
sudo systemctl start beep

# Cleanup
rm -rf "$TMP_DIR"

echo "GeoIP database updated successfully"
```

Make executable:
```bash
chmod +x /opt/beep/update-geoip.sh
```

### Cron Job

Add to crontab for weekly updates:

```bash
0 3 * * 0 /opt/beep/update-geoip.sh >> /var/log/beep-geoip.log 2>&1
```

## Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o beep ./cmd/beep

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' beep

WORKDIR /app
COPY --from=builder /app/beep .

USER beep
EXPOSE 8080

VOLUME ["/data"]
ENTRYPOINT ["./beep"]
CMD ["serve", "--port", "8080", "--db", "/data/tracker.db"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  beep:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - tracker-data:/data
      - ./geoip:/geoip:ro
    command: serve --port 8080 --db /data/tracker.db --geoip /geoip
    restart: unless-stopped

volumes:
  tracker-data:
```

## Backup and Maintenance

### Database Backup

```bash
# Backup SQLite database
sqlite3 /var/lib/beep/tracker.db ".backup '/backup/tracker-$(date +%Y%m%d).db'"

# Or use the .dump method
sqlite3 /var/lib/beep/tracker.db ".dump" | gzip > "/backup/tracker-$(date +%Y%m%d).sql.gz"
```

### Automated Backup Script

Create `/opt/beep/backup.sh`:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/beep"
DB_PATH="/var/lib/beep/tracker.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Create backup
sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/tracker-$DATE.db'"

# Compress
gzip "$BACKUP_DIR/tracker-$DATE.db"

# Remove backups older than 30 days
find "$BACKUP_DIR" -name "tracker-*.db.gz" -mtime +30 -delete

echo "Backup completed: tracker-$DATE.db.gz"
```

### Cron Job for Backups

```bash
0 2 * * * /opt/beep/backup.sh >> /var/log/beep-backup.log 2>&1
```

## Monitoring

### Health Check Endpoint

The server doesn't have a dedicated health check endpoint, but you can monitor the tracking script endpoint:

```bash
curl -f http://localhost:8080/track.js || echo "Service down"
```

### Log Monitoring

Monitor systemd journal:

```bash
sudo journalctl -u beep -f
```

### Resource Monitoring

Monitor disk space for the database:

```bash
df -h /var/lib/beep
```

## Security Considerations

### Firewall

```bash
# Allow only necessary ports
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### Token Security

- Store tokens securely
- Rotate tokens regularly
- Use different tokens for different environments
- Monitor token usage

### Database Security

```bash
# Restrict database file permissions
sudo chown beep:beep /var/lib/beep/tracker.db
sudo chmod 600 /var/lib/beep/tracker.db
```

## Performance Tuning

### SQLite Optimization

The application already sets:
- `PRAGMA foreign_keys = ON`
- `MaxOpenConns = 1` (for SQLite compatibility)

For high traffic, consider:
- Regular `VACUUM` operations
- Monitoring database size
- Implementing data retention policies

### System Tuning

```bash
# Increase file descriptors
echo "beep soft nofile 65536" >> /etc/security/limits.conf
echo "beep hard nofile 65536" >> /etc/security/limits.conf
```

## Troubleshooting

### Common Issues

1. **Permission denied on database**
   ```bash
   sudo chown beep:beep /var/lib/beep/tracker.db
   ```

2. **Service won't start**
   ```bash
   sudo journalctl -u beep -xe
   ```

3. **GeoIP not working**
   - Check file permissions
   - Verify CSV file format
   - Check logs for loading errors

### Debug Mode

Currently, no debug mode is implemented. Check systemd logs for error messages.

## Scaling Considerations

For high-traffic deployments:

1. **Load Balancing**: Use multiple instances behind a load balancer
2. **Database**: Consider migrating to PostgreSQL for concurrent writes
3. **Caching**: Add Redis for session/token caching
4. **CDN**: Serve tracking script via CDN
5. **Analytics Processing**: Implement background processing for statistics