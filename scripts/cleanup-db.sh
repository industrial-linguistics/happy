#!/bin/sh
# Database cleanup script
# Run weekly via cron: 0 2 * * 0 /var/www/vhosts/happy.industrial-linguistics.com/bin/cleanup-db.sh

DB_PATH="/var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db"

echo "Cleaning up database: $DB_PATH"

# Remove old activity logs (keep 30 days)
sqlite3 "$DB_PATH" <<EOF
DELETE FROM activity_log
WHERE timestamp < datetime('now', '-30 days');
EOF

# Remove old rate limit data (keep 24 hours)
sqlite3 "$DB_PATH" <<EOF
DELETE FROM request_stats
WHERE minute_bucket < datetime('now', '-1 day');
EOF

# Vacuum to reclaim space
sqlite3 "$DB_PATH" "VACUUM;"

echo "Database cleanup complete"

# Note: Cleanup is logged to syslog, not a separate file
