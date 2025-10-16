# Happy API - Positive Social Network API

A simple REST API for the Claude Code training course that delivers encouraging messages to students. Built for OpenBSD httpd with CGI support.

## Overview

This API supports training exercises where students build clients that fetch positive messages. The instructor can monitor student activity in real-time to track progress during hands-on sessions.

## Features

- **GET /v1/message** - Retrieve random encouraging messages
- **POST /v1/message** - Send positive messages to other users
- **GET /v1/messages** - Retrieve messages for a recipient
- **GET /v1/status** - Health check endpoint
- **Real-time monitoring** - CLI tool to watch student activity
- **Rate limiting** - 100 requests/minute per IP
- **Activity logging** - Track all API usage with session IDs

## Quick Start

### Build and Deploy

Development happens directly on the OpenBSD server (merah.cassia.ifost.org.au):

```bash
# Build
make build

# Deploy (uses doas to copy to vhost directories)
make deploy
```

The `make deploy` target:
- Copies binaries to vhost directories
- Creates API endpoint symlinks
- Sets www:www ownership
- Sets executable permissions
- Initializes database (first time only)

### Local Testing

For local development/testing, you'll need to adjust the database path in the code or set up the directory structure:

```bash
# Create local directory structure (for testing)
sudo mkdir -p /var/www/vhosts/happy.industrial-linguistics.com/data
sudo chown $USER /var/www/vhosts/happy.industrial-linguistics.com/data

# Build and initialize
make build-local
./bin/init-db-local

# Test manually
echo 'REQUEST_METHOD=GET
PATH_INFO=/message
QUERY_STRING=name=TestUser&session_id=test123
REMOTE_ADDR=127.0.0.1' | ./bin/message-api-local
```

## API Documentation

### GET /v1/message

Retrieve a random positive message.

**Query Parameters:**
- `name` (required): User's name (1-50 chars)
- `session_id` (optional): Training session identifier

**Example:**
```bash
curl "https://happy.industrial-linguistics.com/v1/message?name=Kevin&session_id=session_001"
```

**Response:**
```json
{
  "name": "Kevin",
  "message": "You're doing an amazing job!",
  "timestamp": "2025-10-14T14:30:00Z",
  "message_id": "msg_abc123",
  "sequence": 5
}
```

### POST /v1/message

Send a positive message to another user.

**Request Body:**
```json
{
  "from": "Alice",
  "to": "Bob",
  "message": "Great work on that refactoring!",
  "session_id": "session_001"
}
```

**Example:**
```bash
curl -X POST https://happy.industrial-linguistics.com/v1/message \
  -H "Content-Type: application/json" \
  -d '{"from":"Alice","to":"Bob","message":"Great work!","session_id":"session_001"}'
```

**Response:**
```json
{
  "message_id": "msg_xyz789",
  "timestamp": "2025-10-14T14:31:00Z",
  "status": "delivered"
}
```

### GET /v1/messages

Retrieve messages for a recipient.

**Query Parameters:**
- `recipient` (required): Username to fetch messages for
- `limit` (optional): Max messages (default: 10, max: 50)

**Example:**
```bash
curl "https://happy.industrial-linguistics.com/v1/messages?recipient=Bob&limit=5"
```

**Response:**
```json
{
  "recipient": "Bob",
  "count": 2,
  "messages": [
    {
      "message_id": "msg_xyz789",
      "from": "Alice",
      "message": "Great work!",
      "timestamp": "2025-10-14T14:31:00Z"
    }
  ]
}
```

### GET /v1/status

Health check endpoint.

**Example:**
```bash
curl https://happy.industrial-linguistics.com/v1/status
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "requests_today": 1234
}
```

## Monitoring with happywatch

The `happywatch` CLI tool provides real-time monitoring of student activity.
For a browser-based snapshot of the same information, visit the CGI dashboard
at `/v1/happywatch` on the deployment host. It renders the live activity list,
summary statistics, student progress table, and inactive student report using
the same queries as the CLI.

### Live Mode (default)

Watch activity in real-time:

```bash
happywatch
```

Output:
```
=== Live Activity Monitor ===
Ctrl+C to exit

✓ [14:30:15] Kevin           /message (session: session_001)
✓ [14:30:18] Alice           /message (session: session_001)
✗ [14:30:22] Bob             /message (session: none)
```

### Summary Mode

View statistics for the session:

```bash
happywatch -mode summary
```

Output:
```
=== Activity Summary ===

Total Requests: 156
Active Students: 8

Requests by Endpoint:
  /message     124
  /status      20
  /messages    12

Error Rate: 2.6% (4 errors)
```

### Student Progress

See detailed progress for each student:

```bash
happywatch -mode students
```

Output:
```
=== Student Progress ===

Student              Requests   First Seen   Last Seen   Sessions
-------              --------   ----------   ---------   --------
Kevin                23         09:15        10:42       2
Alice                19         09:18        10:40       1
Bob                  15         09:20        10:30       1

=== Inactive Students (>15 min) ===

  Charlie (last seen 22m ago)
  David (last seen 18m ago)
```

### Export Mode

Export activity as CSV:

```bash
# All activity
happywatch -mode export > session.csv

# Single student
happywatch -mode export -student "Kevin" > kevin.csv

# Since a specific time
happywatch -mode export -since "2025-10-14T09:00:00+08:00" > morning.csv
```

### Monitoring During Training

Recommended setup with multiple terminals:

```bash
# Terminal 1: Live activity
happywatch

# Terminal 2: Student progress (refresh every 30s)
watch -n 30 'happywatch -mode students'

# Terminal 3: Summary stats (refresh every 60s)
watch -n 60 'happywatch -mode summary'
```

## Student Exercise Examples

### Exercise 1: Simple Client

```bash
# Students build a simple program that hits the API
curl "https://happy.industrial-linguistics.com/v1/message?name=YourName&session_id=bitmex_java_001"
```

### Exercise 2: Parse JSON Response

```java
// Students write Java code to parse the response
String apiUrl = "https://happy.industrial-linguistics.com/v1/message";
String params = "?name=" + name + "&session_id=bitmex_java_001";
// ... parse JSON and display message
```

### Exercise 3: Web Interface

```html
<!-- Students build a simple HTML form -->
<form id="messageForm">
  <input type="text" name="name" placeholder="Your name">
  <button type="submit">Get Encouragement</button>
</form>
<div id="message"></div>
```

### Exercise 4: Store Messages Locally

Students add SQLite database to store received messages.

### Exercise 5: Send Messages

Students implement POST functionality to send encouraging messages to classmates.

## Configuration

### Environment Variables

For local development, you can override the database path:

```bash
# In the Go code, make dbPath configurable:
dbPath := os.Getenv("DB_PATH")
if dbPath == "" {
    dbPath = "/var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db"
}
```

### httpd Configuration

See `deploy/httpd.conf` for complete configuration. Key settings:

```
location "/v1/*" {
    fastcgi socket "/run/slowcgi.sock"
    root "/vhosts/happy.industrial-linguistics.com/v1/message-api"
}
```

## Maintenance

### Database Cleanup

Run weekly to clean old logs:

```bash
# Manual cleanup
doas /var/www/vhosts/happy.industrial-linguistics.com/bin/cleanup-db.sh

# Or add to crontab
0 2 * * 0 /var/www/vhosts/happy.industrial-linguistics.com/bin/cleanup-db.sh
```

### Log Rotation

Add to `/etc/newsyslog.conf`:

```
/var/www/vhosts/happy.industrial-linguistics.com/logs/access.log  www:www  644  7  *  $D0  Z
```

### Backup

```bash
# Backup database
cp /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db \
   /var/backups/positive-social-$(date +%Y%m%d).db

# Compress old backups
gzip /var/backups/positive-social-*.db
```

## Testing

Run the test suite:

```bash
./scripts/test-api.sh
```

Set custom base URL:

```bash
BASE_URL=http://localhost:8080/v1 ./scripts/test-api.sh
```

## Troubleshooting

### No requests appearing in happywatch

1. Check database permissions:
   ```bash
   ls -l /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db
   # Should be owned by www:www
   ```

2. Check httpd logs:
   ```bash
   tail -f /var/www/vhosts/happy.industrial-linguistics.com/logs/error.log
   ```

3. Verify slowcgi is running:
   ```bash
   rcctl check slowcgi
   ```

### High error rates

Check what errors are occurring:

```bash
sqlite3 /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db \
  "SELECT response_code, COUNT(*) FROM activity_log
   WHERE response_code >= 400
   GROUP BY response_code"
```

### Students can't connect

1. Test from server:
   ```bash
   curl http://localhost/v1/status
   ```

2. Check firewall:
   ```bash
   pfctl -sr | grep 443
   ```

3. Verify DNS:
   ```bash
   nslookup happy.industrial-linguistics.com
   ```

## Architecture

```
┌─────────────┐
│   Student   │
│   Client    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────────────┐
│  OpenBSD httpd      │
│  (TLS termination)  │
└──────┬──────────────┘
       │ CGI via slowcgi
       ▼
┌─────────────────────────────────────┐
│  message-api (Go CGI binary)        │
│  /var/www/vhosts/.../v1/            │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  SQLite Database                    │
│  /var/www/vhosts/.../data/          │
└─────────────────────────────────────┘

         ┌──────────────┐
         │  happywatch    │
         │  (monitoring)│
         └──────┬───────┘
                │
                ▼
         (same database)
```

## Security Considerations

- **Input validation**: All inputs sanitized (name/message length, content)
- **Rate limiting**: 100 req/min per IP
- **SQL injection**: Parameterized queries throughout
- **XSS prevention**: JSON-only responses
- **CORS**: Wide-open for student convenience
- **Chroot**: httpd runs in `/var/www` jail
- **No authentication**: Intentionally open for training exercises

**Note:** This is designed for training environments, not production use.

## License

This is training course material. Use for educational purposes.

## Support

For issues during training:
1. Check `happywatch -mode summary` for error patterns
2. Test with `curl` to isolate client vs server issues
3. Review `/var/www/vhosts/happy.industrial-linguistics.com/logs/error.log` for server errors

## Files

```
happy-api/
├── README.md                 # This file
├── Makefile                  # Build automation
├── cmd/
│   ├── message-api.go       # Main API handler
│   ├── happywatch.go        # Monitoring tool
│   └── init-db.go           # Database initialization
├── scripts/
│   ├── install.sh           # Installation script
│   ├── test-api.sh          # API test suite
│   └── cleanup-db.sh        # Database maintenance
└── deploy/
    └── httpd.conf           # Example httpd configuration
```

## Development

### Dependencies

```bash
go get github.com/mattn/go-sqlite3
```

### Building

```bash
# Local development
make build-local

# OpenBSD deployment
make build
```

### Adding Messages

Edit `cmd/init-db.go` and add to `positiveMessages` array, rebuild, then reinitialize:

```bash
make build
# Copy new init-db to server, then:
doas -u www /var/www/vhosts/happy.industrial-linguistics.com/bin/init-db
```

## Credits

Built for Claude Code training at BitMEX, October 2025.
