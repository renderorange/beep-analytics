# CLI Documentation

## Overview

The `beep-analytics` command-line interface provides tools for managing sites, tokens, IP filters, and viewing statistics.

All commands support `--help` (or `-h`) to display usage information.

## Global Options

All commands support these global options, which can appear before or after the subcommand:

- `--server URL`: API server URL (default: `http://localhost:8080`)
- `--token TOKEN`: API token (or set via environment variable)

## Authentication

### Token Sources

1. **Command-line flag**: `--token YOUR_TOKEN`
2. **Environment variable**: `BEEP_TOKEN`
3. **Config file**: `~/.config/beep-analytics/token`

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
./beep-analytics serve [options]
```

**Options:**
- `--port PORT`: Port to listen on (default: 8080)
- `--db PATH`: Path to SQLite database (default: beep-analytics.db)
- `--geoip PATH`: Path to GeoLite2-City directory (optional)

**Example:**
```bash
./beep-analytics serve --port 9000 --db /var/lib/beep/data.db
```

#### version

Display version information.

```bash
./beep-analytics version
```

**Output:**
```
beep-analytics v0.1.0
```

### Site Management

#### add-site

Register a new site for tracking.

```bash
./beep-analytics add-site <domain> [options]
```

**Arguments:**
- `domain`: The domain to track (e.g., `example.com`)

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics add-site example.com
./beep-analytics add-site blog.example.com --server http://analytics.example.com
```

**Output:**
```
Site example.com added
```

#### remove-site

Remove a site from tracking.

```bash
./beep-analytics remove-site <domain> [options]
```

**Arguments:**
- `domain`: The domain to remove

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics remove-site example.com
```

**Output:**
```
Site example.com removed
```

#### list-sites

List all registered sites.

```bash
./beep-analytics list-sites [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics list-sites
```

**Output:**
```
example.com
blog.example.com
```

When no sites are registered, outputs `No sites registered`.

### IP Filtering

#### ignore-ip

Add an IP address to the ignore list.

```bash
./beep-analytics ignore-ip <ip> [options]
```

**Arguments:**
- `ip`: The IP address to ignore (e.g., `192.168.1.1`)

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics ignore-ip 192.168.1.1  # Note: CIDR notation not supported
```

**Output:**
```
IP 192.168.1.1 added to ignore list
```

#### unignore-ip

Remove an IP address from the ignore list.

```bash
./beep-analytics unignore-ip <ip> [options]
```

**Arguments:**
- `ip`: The IP address to remove

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics unignore-ip 192.168.1.1
```

**Output:**
```
IP 192.168.1.1 removed from ignore list
```

#### list-ignored

List all ignored IP addresses.

```bash
./beep-analytics list-ignored [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics list-ignored
```

**Output:**
```
192.168.1.1
10.0.0.1
```

When no IPs are ignored, outputs `No IPs ignored`.

### Token Management

#### generate-token

Generate a new API token.

```bash
./beep-analytics generate-token [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token (optional for initial setup)

**Example:**
```bash
./beep-analytics generate-token
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
./beep-analytics revoke-token <id> [options]
```

**Arguments:**
- `id`: The numeric token ID to revoke

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token

**Example:**
```bash
./beep-analytics revoke-token 1
```

**Output:**
```
Token 1 revoked
```

### Statistics

#### stats

View pageview statistics.

```bash
./beep-analytics stats [options]
```

**Options:**
- `--server URL`: API server URL
- `--token TOKEN`: API token
- `--site DOMAIN`: Filter by site domain
- `--last PERIOD`: Relative time period (24h, 7d, 30d, 1mo, 3mo, 6mo)
- `--from DATE`: Start date (YYYY-MM-DD). Without --to, goes from this date to now.
- `--to DATE`:   End date (YYYY-MM-DD). Without --from, goes from all time to this date.
- `--verbose`, `-v`: Show detailed view

**Examples:**

```bash
# Last 24 hours (default)
./beep-analytics stats

# Last 7 days
./beep-analytics stats --last 7d

# Last 3 months
./beep-analytics stats --last 3mo

# Specific date range
./beep-analytics stats --from 2024-01-01 --to 2024-01-31

# From a date to now
./beep-analytics stats --from 2024-06-01

# All time up to a date
./beep-analytics stats --to 2024-06-01

# Filter by site
./beep-analytics stats --site example.com

# Verbose output
./beep-analytics stats --verbose

# Combine options
./beep-analytics stats --site example.com --last 30d --verbose
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
IP                 Country  Region           City             Browser      OS           Path            Referrer             Time
192.168.1.1       US       California       San Francisco    Chrome 120   Windows 10   /               https://google.com   2024-01-01T12:00:00
10.0.0.1          GB       England          London           Firefox 120  macOS 14     /about          (direct)             2024-01-01T11:30:00
```

## Configuration Examples

### Using Environment Variables

```bash
export BEEP_TOKEN="your-token-here"
export BEEP_SERVER="http://analytics.example.com"

./beep-analytics list-sites
```

### Using Config File

```bash
mkdir -p ~/.config/beep-analytics
echo "your-token-here" > ~/.config/beep-analytics/token
chmod 600 ~/.config/beep-analytics/token

./beep-analytics list-sites
```

### Remote Server

```bash
./beep-analytics add-site example.com --server http://analytics.example.com --token your-token
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
  ./beep-analytics add-site $domain
done

# Ignore multiple IPs
for ip in 192.168.1.1 192.168.1.2 192.168.1.3; do
  ./beep-analytics ignore-ip $ip
done
```
