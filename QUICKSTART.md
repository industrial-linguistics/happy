# Quick Start Guide

## For Students

### Testing the API

```bash
# Get a message
curl "https://happy.industrial-linguistics.com/v1/message?name=YourName&session_id=bitmex_java_20251014"

# Get status
curl "https://happy.industrial-linguistics.com/v1/status"
```

### First Exercise

Ask Claude:
```
Create a Java program that makes an HTTP GET request to
https://happy.industrial-linguistics.com/v1/message
with my name and session_id=bitmex_java_20251014,
parses the JSON, and prints the message.
```

See [EXERCISES.md](EXERCISES.md) for the full exercise sequence.

---

## For Instructor

### Deploy to Server

```bash
# Build and deploy in one command (to happy@merah.cassia.ifost.org.au)
make deploy

# Or deploy to a different server
make deploy DEPLOY_HOST=user@other-server.com
```

Deployment is automated - no manual steps needed on the server.

### Monitor Students

```bash
# Live activity
happywatch

# Student progress
happywatch -mode students

# Summary stats
happywatch -mode summary
```

### Before Class

1. Deploy API to server
2. Test endpoint: `curl https://happy.industrial-linguistics.com/v1/status`
3. Start monitoring: `happywatch -mode students`
4. Share base URL with students

### During Class

Terminal 1:
```bash
happywatch
```

Terminal 2:
```bash
watch -n 30 'happywatch -mode students'
```

You'll see:
- Who's making requests (actively working)
- Who's stuck (no activity >15 min)
- Error patterns (if many students hitting same error)

### After Class

Export session data:
```bash
happywatch -mode export -since "2025-10-14T09:00:00+08:00" > session-data.csv
```

---

## Project Structure

```
happy-api/
├── cmd/                    # Go source files
│   ├── message-api.go     # Main API handler (CGI)
│   ├── happywatch.go      # Monitoring CLI tool
│   └── init-db.go         # Database initialization
├── scripts/               # Shell scripts
│   ├── install.sh         # Server installation
│   ├── test-api.sh        # API testing
│   └── cleanup-db.sh      # Database maintenance
├── deploy/
│   └── httpd.conf         # Sample OpenBSD httpd config
├── README.md              # Full documentation
├── EXERCISES.md           # Student exercise guide
├── QUICKSTART.md          # This file
├── Makefile               # Build automation
└── go.mod                 # Go dependencies
```

---

## API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/message` | GET | Get random positive message |
| `/v1/message` | POST | Send message to another user |
| `/v1/messages` | GET | Retrieve messages for user |
| `/v1/status` | GET | Health check |

---

## Common Tasks

### Add More Messages

Edit `cmd/init-db.go`, add to `positiveMessages` array, rebuild, then:
```bash
doas /var/www/vhosts/happy.industrial-linguistics.com/bin/init-db
```

### View Logs

```bash
# API errors
tail -f /var/www/vhosts/happy.industrial-linguistics.com/logs/error.log

# Access log
tail -f /var/www/vhosts/happy.industrial-linguistics.com/logs/access.log

# Activity via database
happywatch -mode summary
```

### Backup Database

```bash
cp /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db \
   backup-$(date +%Y%m%d).db
```

### Clear Old Data

```bash
doas /var/www/vhosts/happy.industrial-linguistics.com/bin/cleanup-db.sh
```

---

## Troubleshooting

**Students can't connect**
```bash
# Test locally on server
curl http://localhost/v1/status

# Check httpd running
rcctl check httpd

# Check slowcgi running
rcctl check slowcgi

# View errors
tail /var/www/vhosts/happy.industrial-linguistics.com/logs/error.log
```

**Not seeing activity in happywatch**
```bash
# Check database permissions
ls -l /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db
# Should be: www:www

# Check if database exists
sqlite3 /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db \
  "SELECT COUNT(*) FROM activity_log"
```

**High error rates**
```bash
# See what errors
happywatch -mode summary

# View specific errors
sqlite3 /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db \
  "SELECT response_code, COUNT(*) FROM activity_log
   WHERE response_code >= 400 GROUP BY response_code"
```

---

## Support

- Full docs: [README.md](README.md)
- Exercises: [EXERCISES.md](EXERCISES.md)
- API design: See design document in project root
