#!/bin/sh
# Update from git and deploy if changes were pulled
# Run this script on the OpenBSD server (merah)

set -e

echo "Checking for updates..."

# Get current HEAD
OLD_HEAD=$(git rev-parse HEAD)

# Pull latest changes
git pull

# Get new HEAD
NEW_HEAD=$(git rev-parse HEAD)

# Check if anything changed
if [ "$OLD_HEAD" != "$NEW_HEAD" ]; then
    echo ""
    echo "Changes detected, deploying..."
    echo "  Old: $OLD_HEAD"
    echo "  New: $NEW_HEAD"
    echo ""
    make deploy
else
    echo "Already up to date, nothing to deploy"
fi
