# Virtual Host Directory Structure

## Overview

The Happy API is deployed using OpenBSD's virtual host directory structure under `/var/www/vhosts/`.

## Directory Layout

```
/var/www/vhosts/happy.industrial-linguistics.com/
├── htdocs/                 # Web pages (DocumentRoot)
│   └── index.html         # Optional: landing page
│
├── v1/                    # CGI programs for API v1
│   ├── message-api        # Main API handler (CGI binary)
│   ├── message -> message-api   # Symlink
│   ├── messages -> message-api  # Symlink
│   └── status -> message-api    # Symlink
│
├── data/                  # Persistent data
│   └── positive-social.db # SQLite database
│
└── bin/                   # Utility programs
    ├── init-db            # Database initialization
    ├── happywatch         # Monitoring CLI tool
    └── cleanup-db.sh      # Maintenance script

Logs (shared):
/var/www/logs/
├── access.log             # HTTP access log
└── error.log              # HTTP error log
```

## httpd Configuration

In `/etc/httpd.conf`:

```nginx
server "happy.industrial-linguistics.com" {
    log style combined
    directory { auto index }
    listen on $listen_addr port 80
    root "/vhosts/happy.industrial-linguistics.com/htdocs"

    # CGI handler for API
    # Uses symlinks: /v1/message, /v1/messages, /v1/status all point to message-api
    location "/v1/*" {
        fastcgi socket "/run/slowcgi.sock"
        root "/vhosts/happy.industrial-linguistics.com/v1"
    }
}
```

**Note:** No TLS - the CDN handles SSL termination.

## URL Mapping

| URL Path | Executes | Actual Binary | Type |
|----------|----------|---------------|------|
| `/v1/message` | `/var/www/vhosts/.../v1/message` | `message-api` (via symlink) | API |
| `/v1/messages` | `/var/www/vhosts/.../v1/messages` | `message-api` (via symlink) | API |
| `/v1/status` | `/var/www/vhosts/.../v1/status` | `message-api` (via symlink) | API |
| `/` | `/var/www/vhosts/.../htdocs/` | Static files | Static |

## Permissions

All files under the vhost directory should be owned by `www:www`:

```bash
chown -R www:www /var/www/vhosts/happy.industrial-linguistics.com
chmod 755 /var/www/vhosts/happy.industrial-linguistics.com/v1/*
chmod 755 /var/www/vhosts/happy.industrial-linguistics.com/bin/*
chmod 644 /var/www/vhosts/happy.industrial-linguistics.com/data/*.db
```

## How CGI Works

1. Client requests `https://happy.industrial-linguistics.com/v1/message?name=Alice`
2. CDN forwards to `http://merah.cassia.ifost.org.au/v1/message?name=Alice`
3. httpd receives request on port 80
4. httpd matches `location "/v1/*"` rule
5. httpd looks for CGI script at `root` + stripped path: `/vhosts/.../v1/message`
6. Finds symlink: `message` → `message-api`
7. httpd sets CGI environment variables:
   - `REQUEST_METHOD=GET`
   - `PATH_INFO=/message`
   - `QUERY_STRING=name=Alice`
   - `REMOTE_ADDR=<client-ip>`
8. httpd executes `message-api` via slowcgi
9. `message-api` binary:
   - Parses CGI environment (checks `PATH_INFO` to route)
   - Opens database at `../data/positive-social.db`
   - Processes request based on `PATH_INFO`
   - Outputs HTTP headers and JSON to stdout
10. httpd sends response to client
11. CDN forwards response to client with HTTPS

## Installation

Use the provided scripts:

```bash
# Build and deploy (to default server)
make deploy

# Or deploy to a different server
make deploy DEPLOY_HOST=user@other-server.com

# Or manually on server
doas ./scripts/install.sh
```

## Monitoring

The `happywatch` tool monitors activity by reading the database:

```bash
# Live monitoring
happywatch -mode live

# Student progress
happywatch -mode students

# Summary
happywatch -mode summary
```

## Adding Additional API Versions

To add a v2 API:

```bash
# Create v2 directory
mkdir -p /var/www/vhosts/happy.industrial-linguistics.com/v2

# Copy new API binary
cp message-api-v2 /var/www/vhosts/happy.industrial-linguistics.com/v2/

# Update httpd.conf
location "/v2/*" {
    fastcgi socket "/run/slowcgi.sock"
    root "/vhosts/happy.industrial-linguistics.com/v2/message-api-v2"
}
```

## Troubleshooting

### Database Permission Errors

```bash
# Fix permissions
doas chown www:www /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db
doas chmod 644 /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db
```

### CGI Not Executing

```bash
# Check slowcgi is running
rcctl check slowcgi

# Check binary is executable
ls -l /var/www/vhosts/happy.industrial-linguistics.com/v1/message-api

# Check httpd error log
tail -f /var/www/vhosts/happy.industrial-linguistics.com/logs/error.log
```

### Static Files Not Serving

```bash
# Check htdocs directory exists
ls -ld /var/www/vhosts/happy.industrial-linguistics.com/htdocs

# Check file exists and permissions
ls -l /var/www/vhosts/happy.industrial-linguistics.com/htdocs/index.html

# Should be readable by www
chmod 644 /var/www/vhosts/happy.industrial-linguistics.com/htdocs/*
```

## Migration from Old Structure

If migrating from `/var/www/data` and `/var/www/bin`:

```bash
# Create new structure
doas mkdir -p /var/www/vhosts/happy.industrial-linguistics.com/{htdocs,v1,data,logs,bin}

# Move database
doas mv /var/www/data/positive-social.db \
        /var/www/vhosts/happy.industrial-linguistics.com/data/

# Move binaries
doas mv /var/www/bin/message-api \
        /var/www/vhosts/happy.industrial-linguistics.com/v1/
doas mv /var/www/bin/init-db \
        /var/www/vhosts/happy.industrial-linguistics.com/bin/

# Fix permissions
doas chown -R www:www /var/www/vhosts/happy.industrial-linguistics.com

# Update httpd.conf
# (see above for new configuration)

# Reload httpd
doas rcctl reload httpd
```

## Benefits of Virtual Host Structure

1. **Isolation**: Each domain/service has its own directory
2. **Clarity**: Clear separation of static files, CGI, data, logs
3. **Security**: httpd chroot doesn't need global access to /var/www
4. **Scalability**: Easy to add more virtual hosts
5. **Maintenance**: Easier to backup/restore individual sites
6. **Standards**: Follows common web server conventions

## Notes

- The `/var/www` chroot is still in effect
- Paths in httpd.conf are relative to `/var/www/`
- CGI programs run as user `www`
- Database must be in a location writable by `www`
- Logs should rotate via `newsyslog(8)`
