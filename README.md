# beep

A lightweight, self-hosted web analytics tracker designed for simplicity and privacy.

## Features

- **Simple Tracking**: Lightweight JavaScript tracker that captures page views
- **Privacy-Focused**: No cookies, no personal data collection
- **Self-Hosted**: Complete control over your data
- **CLI Management**: Command-line interface for site and token management
- **REST API**: Full API for programmatic access
- **GeoIP Support**: Optional geographic location tracking
- **IP Filtering**: Ability to ignore specific IP addresses
- **Token Authentication**: Secure API access with bearer tokens

## Quick Start

```bash
# Build
go build -o beep ./cmd/beep

# Start server
./beep serve --port 8080 --db beep.db

# Generate API token
./beep generate-token

# Register a site
./beep add-site example.com
```

Add the tracking script to your website:

```html
<script src="https://your-tracker-server/track.js" async></script>
```

## Documentation

- [CLI Reference](docs/CLI.md) - All commands and usage
- [API Reference](docs/API.md) - HTTP API endpoints
- [Deployment Guide](docs/DEPLOYMENT.md) - Production setup

## Architecture

The system consists of:

1. **Tracking Server**: HTTP server that receives pageview data
2. **SQLite Database**: Stores sites, pageviews, tokens, and ignored IPs
3. **JavaScript Tracker**: Client-side script that sends pageview data
4. **CLI Tool**: Command-line interface for management tasks

## Testing

Run all unit tests:

```bash
go test ./...
```

Run a specific package's tests:

```bash
go test ./internal/db -v
```

Run the integration test (starts a real server):

```bash
go test -tags=integration -v ./tests/
```

Run everything (fmt, vet, unit + integration):

```bash
make check
```

See the [Makefile](Makefile) for all available targets (`make`, `make build`, `make test`, `make test-integration`, etc.).

## Copyright and License

`beep` is Copyright (c) 2026 Blaine Motsinger under the MIT license.
