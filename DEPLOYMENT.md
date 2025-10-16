# Deployment Guide

## Server Architecture

- **Physical server**: `merah.cassia.ifost.org.au` (OpenBSD)
- **Public hostname**: `happy.industrial-linguistics.com` (CDN front-end)
- **SSH/deployment target**: `happy@merah.cassia.ifost.org.au`

The public URL is served through a CDN that fronts the actual OpenBSD server.

## Quick Deployment

```bash
# Build and deploy in one command
make deploy

# This will:
# 1. Build OpenBSD binaries
# 2. Create vhost directories on server
# 3. Copy binaries to final locations
# 4. Create symlinks for API endpoints
# 5. Set permissions
# 6. Initialize database (if needed)
# 7. Reload httpd
```

## Manual Deployment

If you can't use `make deploy` (e.g., no SSH access), use the install script on the server:

### Step 1: Build and Copy

```bash
make build
scp bin/* happy@merah.cassia.ifost.org.au:/tmp/
scp scripts/install.sh happy@merah.cassia.ifost.org.au:/tmp/
```

### Step 2: Run Install Script

```bash
ssh happy@merah.cassia.ifost.org.au
doas /tmp/install.sh
```

The install script handles directory creation, permissions, symlinks, and service reload.

## First-Time Setup

### 1. Configure httpd

On the server, edit `/etc/httpd.conf`:

```nginx
server "happy.industrial-linguistics.com" {
    log style combined
    directory { auto index }
    listen on $listen_addr port 80
    root "/vhosts/happy.industrial-linguistics.com/htdocs"

    location "/v1/*" {
        fastcgi socket "/run/slowcgi.sock"
        root "/vhosts/happy.industrial-linguistics.com/v1"
    }
}
```

**Note:** No TLS in httpd - the CDN (happy.industrial-linguistics.com) handles SSL termination.

The `install.sh` script creates symlinks in `/v1/`:
- `message` → `message-api`
- `messages` → `message-api`
- `status` → `message-api`

So requests to `/v1/message`, `/v1/messages`, and `/v1/status` all execute the same binary.

### 2. SSL Certificates

**Not needed.** The CDN (happy.industrial-linguistics.com) handles SSL termination.

The backend server (merah.cassia.ifost.org.au) receives plain HTTP on port 80.

### 3. Enable Services

```bash
# Enable and start slowcgi (for CGI support)
doas rcctl enable slowcgi
doas rcctl start slowcgi

# Reload httpd
doas rcctl reload httpd
```

### 4. Test Locally

On the server:

```bash
# Test API endpoint
curl http://localhost/v1/status

# Should return:
# {"status":"ok","version":"1.0.0","requests_today":0}
```

### 5. Test Publicly

From anywhere:

```bash
curl https://happy.industrial-linguistics.com/v1/status
```

## Updating After Code Changes

```bash
# Rebuild and deploy
make deploy

# The install.sh script will:
# - Overwrite the message-api binary
# - Restart slowcgi if needed
# - Reload httpd
```

## Database Maintenance

### Initialize Database (First Time)

```bash
ssh happy@merah.cassia.ifost.org.au
doas -u www /var/www/vhosts/happy.industrial-linguistics.com/bin/init-db
```

### Backup Database

```bash
ssh happy@merah.cassia.ifost.org.au
cp /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db \
   ~/backups/positive-social-$(date +%Y%m%d).db
```

### Clean Old Logs

```bash
ssh happy@merah.cassia.ifost.org.au
doas /var/www/vhosts/happy.industrial-linguistics.com/bin/cleanup-db.sh
```

## Monitoring

### From Your Local Machine

After deployment, you can monitor in real-time:

```bash
# SSH and run happywatch (full path)
ssh happy@merah.cassia.ifost.org.au /var/www/vhosts/happy.industrial-linguistics.com/bin/happywatch

# Or specific modes
ssh happy@merah.cassia.ifost.org.au "/var/www/vhosts/happy.industrial-linguistics.com/bin/happywatch -mode students"
ssh happy@merah.cassia.ifost.org.au "/var/www/vhosts/happy.industrial-linguistics.com/bin/happywatch -mode summary"
```

### On the Server

```bash
ssh happy@merah.cassia.ifost.org.au

# Live monitoring
/var/www/vhosts/happy.industrial-linguistics.com/bin/happywatch

# Or add to PATH in ~/.profile:
export PATH=$PATH:/var/www/vhosts/happy.industrial-linguistics.com/bin

# Then you can just use:
happywatch -mode students
happywatch -mode summary
happywatch -mode export -since "2025-10-17T09:00:00+11:00" > session.csv
```

## Troubleshooting Deployment

### SSH Issues

```bash
# Test SSH connection
ssh happy@merah.cassia.ifost.org.au echo "Connected"

# If SSH key issues
ssh -v happy@merah.cassia.ifost.org.au
```

### Permission Issues During Install

If `install.sh` fails with permission errors:

```bash
# Check you can use doas
ssh happy@merah.cassia.ifost.org.au doas whoami
# Should output: root

# If not configured, ask sysadmin to add to /etc/doas.conf:
# permit nopass happy as root
```

### Service Not Starting

```bash
# Check slowcgi status
ssh happy@merah.cassia.ifost.org.au rcctl check slowcgi

# Check httpd status
ssh happy@merah.cassia.ifost.org.au rcctl check httpd

# View logs
ssh happy@merah.cassia.ifost.org.au tail -f /var/www/vhosts/happy.industrial-linguistics.com/logs/error.log
```

### Binary Won't Execute

```bash
# Check binary is for correct architecture
ssh happy@merah.cassia.ifost.org.au file /var/www/vhosts/happy.industrial-linguistics.com/v1/message-api
# Should show: ELF 64-bit LSB executable, x86-64 (or arm64 depending on server)

# Rebuild with correct GOARCH if needed
make build GOARCH=amd64  # or arm64
```

### Database Permission Issues

```bash
ssh happy@merah.cassia.ifost.org.au
doas chown www:www /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db
doas chmod 644 /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db
```

## CDN Configuration

Since `happy.industrial-linguistics.com` is a CDN front-end:

1. **Origin Server**: Point CDN to `merah.cassia.ifost.org.au`
2. **SSL**: CDN handles public SSL; backend may use self-signed or CDN origin cert
3. **Headers**: Ensure CDN passes through:
   - `X-Forwarded-For` (for IP logging)
   - `User-Agent`
   - Query parameters

The CGI program reads client IP from `REMOTE_ADDR` environment variable, which httpd sets from the TCP connection. If behind CDN, you may need to configure httpd to read from `X-Forwarded-For` header instead.

## Deployment Checklist

Before going live:

- [ ] Build completes successfully: `make build`
- [ ] Deploy completes: `make deploy`
- [ ] Database initialized: Check with `happywatch -mode summary`
- [ ] Local test passes: `curl http://localhost/v1/status` (on server)
- [ ] Public test passes: `curl https://happy.industrial-linguistics.com/v1/status`
- [ ] Monitoring works: `happywatch` shows activity
- [ ] Logs rotating: Check `/etc/newsyslog.conf`
- [ ] Backup scheduled: Cron job for database backups

## During Training Session

1. **Before class**: Deploy latest version with `make deploy`
2. **Start monitoring**: `ssh happy@merah.cassia.ifost.org.au` and run `happywatch -mode students`
3. **During class**: Watch for inactive students (>15 min without requests)
4. **After class**: Export session data: `happywatch -mode export > session.csv`

## Rollback Procedure

If deployment causes issues:

```bash
# Restore previous binary from backup
ssh happy@merah.cassia.ifost.org.au
doas cp /var/backups/message-api.previous \
        /var/www/vhosts/happy.industrial-linguistics.com/v1/message-api
doas rcctl reload httpd
```

Consider adding to `install.sh`:
```bash
# Backup current binary before overwriting
if [ -f ${VHOST_DIR}/v1/message-api ]; then
    cp ${VHOST_DIR}/v1/message-api /var/backups/message-api.previous
fi
```

## Environment Variables

You can override deployment target:

```bash
# Deploy to different server
make deploy DEPLOY_HOST=user@test-server.example.com

# Build for different architecture
make build GOARCH=arm64

# Both
make deploy GOARCH=arm64 DEPLOY_HOST=user@arm-server.example.com
```

## Log Files

Logs are at:
- Access: `/var/www/vhosts/happy.industrial-linguistics.com/logs/access.log`
- Error: `/var/www/vhosts/happy.industrial-linguistics.com/logs/error.log`
- Cleanup: `/var/www/vhosts/happy.industrial-linguistics.com/logs/cleanup.log`

Tail them with:
```bash
ssh happy@merah.cassia.ifost.org.au \
  tail -f /var/www/vhosts/happy.industrial-linguistics.com/logs/error.log
```
