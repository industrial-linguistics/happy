#!/bin/sh
# Update from git and deploy if changes were pulled
# Run this script on the OpenBSD server (merah)

set -e

# Get current HEAD
OLD_HEAD=$(git rev-parse HEAD)

# Pull latest changes (silently)
git pull -q

# Get new HEAD
NEW_HEAD=$(git rev-parse HEAD)

# Check if anything changed
if [ "$OLD_HEAD" != "$NEW_HEAD" ]; then
    echo "Changes detected, deploying..."
    echo "  Old: $OLD_HEAD"
    echo "  New: $NEW_HEAD"
    echo ""
    make deploy
fi
