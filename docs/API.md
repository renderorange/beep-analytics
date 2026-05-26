# API Documentation

## Base URL

All API endpoints are relative to the server base URL (e.g., `http://localhost:8080`).

## Authentication

Most endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer YOUR_TOKEN_HERE
```

### Token Bootstrap

The token generation endpoint (`POST /api/tokens/generate`) allows unauthenticated access when no tokens exist in the system. This enables initial setup.

## Endpoints

### Tracking

#### POST /collect

Collects a pageview event. This endpoint is called by the tracking script.

**Headers:**
- `Origin` or `Referer` (required): The domain making the request
- `Content-Type: application/json`

**Request Body:**
```json
{
  "path": "/page-path",
  "referrer": "https://example.com",
  "screen": "1920x1080"
}
```

**Response:**
- `204 No Content`: Pageview recorded successfully
- `400 Bad Request`: Missing origin or invalid JSON
- `204 No Content`: Site not registered (silently ignored)

**Notes:**
- The server extracts the domain from the Origin/Referer header
- IP filtering is applied automatically
- User agent is parsed from the request headers

#### GET /track.js

Returns the tracking JavaScript file.

**Response:**
- `200 OK`: JavaScript tracking code
- Content-Type: `application/javascript`

### Site Management

#### POST /api/sites

Register a new site for tracking.

**Headers:**
- `Authorization: Bearer TOKEN`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "domain": "example.com"
}
```

**Response:**
```json
{
  "id": 1,
  "domain": "example.com"
}
```

**Status Codes:**
- `200 OK`: Site created
- `400 Bad Request`: Missing domain
- `401 Unauthorized`: Invalid or missing token

#### DELETE /api/sites/{domain}

Remove a site from tracking.

**Headers:**
- `Authorization: Bearer TOKEN`

**URL Parameters:**
- `domain`: The domain to remove

**Response:**
- `204 No Content`: Site removed
- `401 Unauthorized`: Invalid or missing token
- `500 Internal Server Error`: Failed to remove site

#### GET /api/sites

List all registered sites.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "domain": "example.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

**Status Codes:**
- `200 OK`: List of sites
- `401 Unauthorized`: Invalid or missing token

### IP Filtering

#### POST /api/ignore

Add an IP address to the ignore list.

**Headers:**
- `Authorization: Bearer TOKEN`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "ip": "192.168.1.1"
}
```

**Response:**
- `201 Created`: IP added to ignore list
- `400 Bad Request`: Missing IP
- `401 Unauthorized`: Invalid or missing token

#### DELETE /api/ignore/{ip}

Remove an IP address from the ignore list.

**Headers:**
- `Authorization: Bearer TOKEN`

**URL Parameters:**
- `ip`: The IP address to remove

**Response:**
- `204 No Content`: IP removed
- `401 Unauthorized`: Invalid or missing token
- `500 Internal Server Error`: Failed to remove IP

#### GET /api/ignore

List all ignored IP addresses.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  "192.168.1.1",
  "10.0.0.1"
]
```

**Status Codes:**
- `200 OK`: List of ignored IPs
- `401 Unauthorized`: Invalid or missing token

### Token Management

#### POST /api/tokens/generate

Generate a new API token.

**Headers:**
- `Authorization: Bearer TOKEN` (optional if no tokens exist)

**Response:**
```json
{
  "token": "abc123...",
  "id": 1
}
```

**Status Codes:**
- `200 OK`: Token generated
- `401 Unauthorized`: Invalid token (when tokens exist)

**Notes:**
- This endpoint allows unauthenticated access when no tokens exist in the system
- This enables initial setup without pre-shared credentials
- Save the returned token immediately - it cannot be retrieved later

#### DELETE /api/tokens/{id}

Revoke an API token.

**Headers:**
- `Authorization: Bearer TOKEN`

**URL Parameters:**
- `id`: The numeric token ID to revoke

**Response:**
- `204 No Content`: Token revoked
- `400 Bad Request`: Invalid token ID
- `401 Unauthorized`: Invalid or missing token

#### GET /api/tokens

List all active tokens.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

**Status Codes:**
- `200 OK`: List of tokens (without token values)
- `401 Unauthorized`: Invalid or missing token

### Statistics

#### GET /api/stats

Retrieve pageview statistics.

**Headers:**
- `Authorization: Bearer TOKEN`

**Query Parameters:**
- `site` (optional): Filter by site domain
- `from` (optional): Start date (YYYY-MM-DD format)
- `to` (optional): End date (YYYY-MM-DD format)
- `last` (optional): Relative time period (24h, 7d, 30d)
- `verbose` (optional): Set to "true" for detailed view

**Response (Aggregate Mode):**
```json
[
  {
    "site": "example.com",
    "ip": "192.168.1.1",
    "path": "/page",
    "count": 42
  }
]
```

**Response (Verbose Mode):**
```json
[
  {
    "site": "example.com",
    "ip": "192.168.1.1",
    "country": "US",
    "browser": "Chrome",
    "os": "Windows",
    "path": "/page",
    "referrer": "https://google.com",
    "time": "2024-01-01T12:00:00Z"
  }
]
```

**Default Behavior:**
- If no time parameters are provided, defaults to last 24 hours
- If `from` and `to` are provided, uses that date range
- If `last` is provided, uses relative time from now

**Status Codes:**
- `200 OK`: Statistics data
- `400 Bad Request`: Invalid date format or parameters
- `401 Unauthorized`: Invalid or missing token
- `404 Not Found`: Site not found (when filtering by site)

## Error Responses

All endpoints may return these error formats:

**400 Bad Request:**
```json
{
  "error": "invalid json"
}
```

**401 Unauthorized:**
```json
{
  "error": "unauthorized"
}
```

**500 Internal Server Error:**
```json
{
  "error": "internal error"
}
```

## Rate Limiting

Currently, no rate limiting is implemented. Consider implementing rate limiting at the reverse proxy level for production deployments.

## CORS

The tracking endpoint (`/collect`) expects requests from browsers. Ensure your server configuration includes appropriate CORS headers if the tracker is served from a different domain.