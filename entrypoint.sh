#!/bin/bash

set -euo pipefail

# Ensure necessary environment variables are set
if [ -z "${LITESTREAM_REPLICA_URI:-}" ]; then
  echo "Error: LITESTREAM_REPLICA_URI is not set."
  exit 1
fi

if [ -z "${DB_PATH:-}" ]; then
  echo "Error: DB_PATH is not set."
  exit 1
fi

# Ensure the directory for the database exists
DB_DIR=$(dirname "$DB_PATH")
mkdir -p "$DB_DIR"

# Attempt to restore the database
echo "Attempting to restore SQLite database from Litestream URI: $LITESTREAM_REPLICA_URI"
if litestream restore -if-replica-exists "$LITESTREAM_REPLICA_URI" "$DB_PATH"; then
  echo "Database restored successfully from Litestream."
fi

# Start the server
echo "Starting the tagging server..."
exec ./tagging-server

