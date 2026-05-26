# CLI Documentation

## Overview

The `beep` command-line interface provides tools for managing sites, tokens, IP filters, and viewing statistics.

## Global Options

All commands support these global options:

- `--server URL`: API server URL (default: `http://localhost:8080`)
- `--token TOKEN`: API token (or set via environment variable)

## Authentication

### Token Sources

1. **Command-line flag**: `--token YOUR_TOKEN`
2. **Environment variable**: `BEEP_TOKEN`
3. **Config file**: `~/.config/beep/token`

Priority order: command-line flag > environment variable > config file.

### Server Location

The server URL can be set via:
- `--server URL`: Command-line flag
- `BEEP_SERVER`: Environment variable

Default: `http://localhost:8080`

### Initial Setup

When no tokens exist in the system, the `generate-token` command works without authentication. This enables initial setup.

## Commands

### Server Management

#### serve

Start the tracking server.

```bash
./beep serve [options]
```

**Options:**
- `--port PORT`: Port to listen on (default: 8080)
- `--db PATH`: Path to SQLite database (default: beep.db)
- `--geoip PATH`: Path to GeoLite2-City directory (optional)

**Example:**
```bash
./beep serve --port 9000 --db /var/lib/beep/data.db
```

#### version

Display version information.

```bash
./beep version
```

**Output:**
```
beep v0.1.0
```

### Site Management

#### add-site

Register a new site for tracking.

```bash
./beep add-site <domain> [options]
```

**Arguments:**
- `domain`: The domain to track (e.g., `example.com`)

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep add-site example.com
./beep add-site blog.example.com --server http://analytics.example.com
```

**Output:**
```
Site example.com added
```

#### remove-site

Remove a site from tracking.

```bash
./beep remove-site <domain> [options]
```

**Arguments:**
- `domain`: The domain to remove

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep remove-site example.com
```

**Output:**
```
Site example.com removed
```

#### list-sites

List all registered sites.

```bash
./beep list-sites [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep list-sites
```

**Output:**
```
example.com
blog.example.com
```

### IP Filtering

#### ignore-ip

Add an IP address to the ignore list.

```bash
./beep ignore-ip <ip> [options]
```

**Arguments:**
- `ip`: The IP address to ignore (e.g., `192.168.1.1`)

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep ignore-ip 192.168.1.1  # Note: CIDR notation not supported
```

**Output:**
```
IP 192.168.1.1 added to ignore list
```

#### unignore-ip

Remove an IP address from the ignore list.

```bash
./beep unignore-ip <ip> [options]
```

**Arguments:**
- `ip`: The IP address to remove

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep unignore-ip 192.168.1.1
```

**Output:**
```
IP 192.168.1.1 removed from ignore list
```

#### list-ignored

List all ignored IP addresses.

```bash
./beep list-ignored [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep list-ignored
```

**Output:**
```
192.168.1.1
10.0.0.1
```

### Token Management

#### generate-token

Generate a new API token.

```bash
./beep generate-token [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token (optional for initial setup)

**Example:**
```bash
./beep generate-token
```

**Output:**
```
Token ID: 1
Token: abc123...

Save this token securely. It cannot be retrieved again.
```

**Notes:**
- Works without authentication if no tokens exist
- Store the token immediately - it cannot be retrieved later

#### revoke-token

Revoke an API token.

```bash
./beep revoke-token <id> [options]
```

**Arguments:**
- `id`: The numeric token ID to revoke

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep revoke-token 1
```

**Output:**
```
Token 1 revoked
```

### Statistics

#### stats

View pageview statistics.

```bash
./beep stats [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token
- `--site DOMAIN`: Filter by site domain
- `--last PERIOD`: Relative time period (24h, 7d, 30d)
- `--from DATE`: Start date (YYYY-MM-DD)
- `--to DATE`: End date (YYYY-MM-DD)
- `--verbose`, `-v`: Show detailed view

**Examples:**

```bash
# Last 24 hours (default)
./beep stats

# Last 7 days
./beep stats --last 7d

# Specific date range
./beep stats --from 2024-01-01 --to 2024-01-31

# Filter by site
./beep stats --site example.com

# Verbose output
./beep stats --verbose

# Combine options
./beep stats --site example.com --last 30d --verbose
```

**Aggregate Output (default):**
```
=== example.com ===
IP                   Path                 Count
192.168.1.1         /                    42
10.0.0.1            /about               15
```

**Verbose Output:**
```
=== example.com ===
IP                 Country  Region           City             Browser    OS         Path            Referrer             Time
192.168.1.1       US       California       San Francisco    Chrome     Windows    /               https://google.com   2024-01-01T12:00:00
10.0.0.1          GB       England          London           Firefox    macOS      /about          (direct)             2024-01-01T11:30:00
```

## Configuration Examples

### Using Environment Variables

```bash
export BEEP_TOKEN="your-token-here"
export BEEP_SERVER="http://analytics.example.com"

./beep list-sites
```

### Using Config File

```bash
mkdir -p ~/.config/beep
echo "your-token-here" > ~/.config/beep/token
chmod 600 ~/.config/beep/token

./beep list-sites
```

### Remote Server

```bash
./beep add-site example.com --server http://analytics.example.com --token your-token
```

## Error Handling

All commands exit with:
- `0`: Success
- `1`: Error (with error message to stderr)

Common errors:
- Missing required arguments
- Invalid token or authentication failure
- Network connection issues
- Server errors

## Shell Completion

Currently, shell completion is not implemented. Consider creating wrapper scripts or aliases for frequently used commands.

## Batch Operations

For batch operations, you can use shell scripting:

```bash
# Add multiple sites
for domain in example.com blog.example.com shop.example.com; do
  ./beep add-site $domain
done

# Ignore multiple IPs
for ip in 192.168.1.1 192.168.1.2 192.168.1.3; do
  ./beep ignore-ip $ip
done
```
