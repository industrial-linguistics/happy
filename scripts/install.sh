#!/bin/sh
# Installation script for OpenBSD
# Run as the deployment user (e.g., happy)

set -e

REPO_DIR=$(cd "$(dirname "$0")/.." && pwd)

VHOST_DIR="/var/www/vhosts/happy.industrial-linguistics.com"

echo "Installing happy-api to $VHOST_DIR..."

# Create directories
mkdir -p ${VHOST_DIR}/data
mkdir -p ${VHOST_DIR}/v1
mkdir -p ${VHOST_DIR}/bin
mkdir -p ${VHOST_DIR}/htdocs

# Copy binaries
if [ -f /tmp/message-api ]; then
    cp /tmp/message-api ${VHOST_DIR}/v1/
    echo "Installed message-api to ${VHOST_DIR}/v1/"

    # Create symlinks for each endpoint
    cd ${VHOST_DIR}/v1
    ln -sf message-api message
    ln -sf message-api messages
    ln -sf message-api automessage
    ln -sf message-api status
    echo "Created symlinks for API endpoints"
fi

if [ -f /tmp/init-db ]; then
    cp /tmp/init-db ${VHOST_DIR}/bin/
    echo "Installed init-db to ${VHOST_DIR}/bin/"
fi

if [ -f /tmp/happywatch ]; then
    cp /tmp/happywatch ${VHOST_DIR}/bin/
    echo "Installed happywatch to ${VHOST_DIR}/bin/"
fi

if [ -f /tmp/happywatch.cgi ]; then
    cp /tmp/happywatch.cgi ${VHOST_DIR}/v1/happywatch
    echo "Installed happywatch CGI to ${VHOST_DIR}/v1/happywatch"
fi

# Copy static assets
if [ -f "${REPO_DIR}/htdocs/index.html" ]; then
    cp "${REPO_DIR}/htdocs/index.html" "${VHOST_DIR}/htdocs/index.html"
    chmod 644 "${VHOST_DIR}/htdocs/index.html"
    echo "Installed index.html to ${VHOST_DIR}/htdocs/"
fi

# Set executable permissions
chmod 755 ${VHOST_DIR}/v1/* 2>/dev/null || true
chmod 755 ${VHOST_DIR}/bin/* 2>/dev/null || true

# Initialize database if it doesn't exist
if [ ! -f ${VHOST_DIR}/data/positive-social.db ]; then
    echo "Initializing database..."
    ${VHOST_DIR}/bin/init-db
else
    echo "Database already exists, skipping initialization"
fi

echo ""
echo "✓ Installation complete!"
echo ""
echo "Directory structure:"
echo "  ${VHOST_DIR}/htdocs/  - Web pages"
echo "  ${VHOST_DIR}/v1/      - CGI programs (message-api + symlinks)"
echo "  ${VHOST_DIR}/data/    - Database (student monitoring via logwatch)"
echo "  ${VHOST_DIR}/bin/     - Utility programs (init-db)"
echo ""
echo "Test with:"
echo "  curl http://localhost/v1/status"
echo "  curl https://happy.industrial-linguistics.com/v1/status"
echo ""
echo "Monitor with:"
echo "  /var/www/vhosts/happy.industrial-linguistics.com/bin/happywatch -mode live"
