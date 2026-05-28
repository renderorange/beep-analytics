# Deployment Guide

## Prerequisites

- Go 1.26.3 or later
- SQLite support (via modernc.org/sqlite)
- Optional: GeoLite2-City database for geographic data

## Building from Source

### Standard Build

```bash
go build -o beep-analytics ./cmd/beep-analytics
# or
make build
```

### Production Build

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o beep-analytics ./cmd/beep-analytics
# or
make build
```

Differences from standard build:
- `CGO_ENABLED=0` - Creates a statically-linked binary with no C dependencies
- `-ldflags="-s -w"` - Strips debug symbols (`-s`) and DWARF info (`-w`), reducing binary size

### Debian Package

Build a `.deb` package:

```bash
./scripts/build-deb.sh
# or
make deb
```

Produces `build/beep-analytics_<version>_amd64.deb`. Version is taken from the latest git tag, or defaults to `0.1.0`.

Install:

```bash
sudo dpkg -i build/beep-analytics_0.1.0_amd64.deb
sudo systemctl start beep-analytics
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o beep-analytics-linux-amd64 ./cmd/beep

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o beep-analytics-darwin-arm64 ./cmd/beep

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o beep-analytics.exe ./cmd/beep
```

## Systemd Service

### Service File

Create `/etc/systemd/system/beep-analytics.service`:

```ini
[Unit]
Description=beep-analytics server
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=beep-analytics
Group=beep-analytics
ExecStart=/usr/bin/beep-analytics serve \
  --port 8080 \
  --db /var/lib/beep-analytics/beep.db \
  --geoip /usr/share/GeoIP
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ReadWritePaths=/var/lib/beep-analytics
ReadOnlyPaths=/usr/share/GeoIP

[Install]
WantedBy=multi-user.target
```

### Installation Steps

```bash
# Create user and directories
sudo useradd -r -s /bin/false beep-analytics
sudo mkdir -p /var/lib/beep-analytics
sudo chown beep:beep /var/lib/beep-analytics

# Copy binary
sudo cp beep-analytics /usr/bin/

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable beep-analytics
sudo systemctl start beep-analytics

# Check status
sudo systemctl status beep-analytics
```

## SSL with Let's Encrypt

### Install Certbot

```bash
# Debian/Ubuntu
sudo apt install certbot python3-certbot-nginx

# RHEL/CentOS/Fedora
sudo dnf install certbot python3-certbot-nginx
```

### Obtain Certificate

```bash
sudo certbot --nginx -d analytics.example.com
```

Certbot will:
1. Verify domain ownership via HTTP challenge
2. Obtain the certificate
3. Modify your Nginx config automatically

### Manual Certificate (without auto-configure)

```bash
sudo certbot certonly --standalone -d analytics.example.com
```

Certificates are stored in `/etc/letsencrypt/live/analytics.example.com/`:
- `fullchain.pem` - Full certificate chain
- `privkey.pem` - Private key

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

## Apache2 Reverse Proxy

### Enable Required Modules

```bash
sudo a2enmod proxy proxy_http ssl headers rewrite expires
sudo systemctl restart apache2
```

### Virtual Host Configuration

Create `/etc/apache2/sites-available/beep-analytics.conf`:

```apache
<VirtualHost *:80>
    ServerName analytics.example.com
    Redirect permanent / https://analytics.example.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName analytics.example.com

    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/analytics.example.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/analytics.example.com/privkey.pem

    # Security headers
    Header always set X-Frame-Options "SAMEORIGIN"
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-XSS-Protection "1; mode=block"

    # Tracking endpoint
    <Location /collect>
        ProxyPass http://127.0.0.1:8080/collect
        ProxyPassReverse http://127.0.0.1:8080/collect

        # CORS headers
        Header always set Access-Control-Allow-Origin "*"
        Header always set Access-Control-Allow-Methods "POST, OPTIONS"
        Header always set Access-Control-Allow-Headers "Content-Type"

        # Handle OPTIONS preflight
        RewriteEngine On
        RewriteCond %{REQUEST_METHOD} OPTIONS
        RewriteRule ^(.*)$ $1 [R=204,L]
    </Location>

    # Tracking script
    <Location /track.js>
        ProxyPass http://127.0.0.1:8080/track.js
        ProxyPassReverse http://127.0.0.1:8080/track.js

        # Cache the tracking script
        Header set Cache-Control "public, immutable"
        ExpiresActive On
        ExpiresDefault "access plus 1 hour"
    </Location>

    # API endpoints
    <Location /api>
        ProxyPass http://127.0.0.1:8080/api
        ProxyPassReverse http://127.0.0.1:8080/api

        # Restrict API access
        Require ip 127.0.0.1
        Require ip your-admin-ip
    </Location>

    # Deny everything else to API
    <LocationMatch "^/api/">
        Require ip 127.0.0.1
    </LocationMatch>
</VirtualHost>
```

### Enable the Site

```bash
sudo a2ensite beep-analytics.conf
sudo systemctl reload apache2
```

### Rate Limiting

Apache2 has two options for rate limiting:

**Option 1: mod_ratelimit (bandwidth limiting)**

```bash
sudo a2enmod ratelimit
sudo systemctl restart apache2
```

```apache
<Location /collect>
    SetOutputFilter RATE_LIMIT
    SetEnv rate-limit 400
    SetEnv rate-initial-burst 512
</Location>
```

**Option 2: mod_evasive (request rate limiting, recommended)**

```bash
sudo apt install libapache2-mod-evasive
```

Create `/etc/apache2/mods-available/evasive.conf`:

```apache
<IfModule mod_evasive20.c>
    DOSHashTableSize    3097
    DOSPageCount        10
    DOSSiteCount        50
    DOSPageInterval     1
    DOSSiteInterval     1
    DOSBlockingPeriod   10
    DOSLogDir           "/var/log/apache2/mod_evasive"
</IfModule>
```

```bash
sudo mkdir -p /var/log/apache2/mod_evasive
sudo chown wwwbeep-analytics-data:www-data /var/log/apache2/mod_evasive
sudo a2enmod evasive
sudo systemctl restart apache2
```

### Connection Limiting

Enable `mod_reqtimeout` to prevent slowloris attacks:

```bash
sudo a2enmod reqtimeout
sudo systemctl restart apache2
```

Default settings in `/etc/apache2/mods-available/reqtimeout.conf` are usually sufficient. To customize:

```apache
<IfModule mod_reqtimeout.c>
    RequestReadTimeout header=20-40,MinRate=500 body=20,MinRate=500
</IfModule>
```

## GeoIP Setup

### Download GeoLite2 Database

1. Sign up for a MaxMind account at https://www.maxmind.com/
2. Generate a license key
3. Download the GeoLite2-City CSV format

### Installation and Update Script

The installation and update script is installed at `/usr/lib/beep-analytics/update-geoip.sh`.

```bash
GEOIP_ACCOUNT_ID=12345 GEOIP_LICENSE_KEY=your-key /usr/lib/beep-analytics/update-geoip.sh >> /var/log/beep-analytics-geoip.log 2>&1
```

### Cron Job

Add to crontab for weekly updates:

```bash
0 3 * * 0 GEOIP_ACCOUNT_ID=12345 GEOIP_LICENSE_KEY=your-key /usr/lib/beep-analytics/update-geoip.sh >> /var/log/beep-analytics-geoip.log 2>&1
```

## Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o beep-analytics ./cmd/beep-analytics

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' beep-analytics

COPY --from=builder /app/beep-analytics /usr/bin/

USER beep-analytics
EXPOSE 8080

VOLUME ["/data"]
ENTRYPOINT ["beep-analytics"]
CMD ["serve", "--port", "8080", "--db", "/data/beep-analytics.db"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  beep-analytics:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - beep-analyticsbeep-analytics-data:/data
      - geoipbeep-analytics-data:/usr/share/GeoIP:ro
    command: serve --port 8080 --db /data/beep-analytics.db --geoip /usr/share/GeoIP
    restart: unless-stopped

volumes:
  beepbeep-analytics-data:
  geoipbeep-analytics-data:
```

## Backup and Maintenance

### Database Backup

```bash
# Backup SQLite database
sqlite3 /var/lib/beep-analytics/beep.db ".backup '/backup/beep-analytics-$(date +%Y%m%d).db'"

# Or use the .dump method
sqlite3 /var/lib/beep-analytics/beep.db ".dump" | gzip > "/backup/beep-analytics-$(date +%Y%m%d).sql.gz"
```

### Automated Backup Script

The backup script is installed at `/usr/lib/beep-analytics/backup.sh`:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/beep-analytics"
DB_PATH="/var/lib/beep-analytics/beep.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Create backup
sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/beep-analytics-$DATE.db'"

# Compress
gzip "$BACKUP_DIR/beep-analytics-$DATE.db"

# Remove backups older than 30 days
find "$BACKUP_DIR" -name "beep-analytics-*.db.gz" -mtime +30 -delete

echo "Backup completed: beep-analytics-$DATE.db.gz"
```

### Cron Job for Backups

```bash
0 2 * * * /usr/lib/beep-analytics/backup.sh >> /var/log/beep-analytics-backup.log 2>&1
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
sudo journalctl -u beep-analytics -f
```

## Security Considerations

### Database Security

```bash
# Restrict database file permissions
sudo chown beep:beep /var/lib/beep-analytics/beep.db
sudo chmod 600 /var/lib/beep-analytics/beep.db
```

## Performance Tuning

### SQLite Optimization

The application already sets:
- `PRAGMA foreign_keys = ON`
- `MaxOpenConns = 1` (for SQLite compatibility)

### Database VACUUM

Over time, SQLite database fragmentation can degrade performance. Regular `VACUUM` reclaims space and defragments the database.

**Manual VACUUM:**
```bash
sqlite3 /var/lib/beep-analytics/beep.db "VACUUM;"
```

**Automated weekly VACUUM via cron:**
```bash
0 4 * * 0 sqlite3 /var/lib/beep-analytics/beep.db "VACUUM;" >> /var/log/beep-analytics-vacuum.log 2>&1
```

VACUUM requires free disk space equal to the database size (it rebuilds the file). Run during low-traffic periods.

### System Tuning

### System Tuning

```bash
# Increase file descriptors
echo "beep-analytics soft nofile 65536" >> /etc/security/limits.conf
echo "beep-analytics hard nofile 65536" >> /etc/security/limits.conf
```
