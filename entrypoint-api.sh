#!/bin/sh

echo "Starting API server on $PORT..."

if [ -z "$DATABASE_PRIVATE_URL" ]; then
  echo "DATABASE_PRIVATE_URL is not set. Exiting."
  exit 1
fi

if [ -z "$PORT" ]; then
  echo "PORT is not set. Exiting."
  exit 1
fi

/api --database-url $DATABASE_PRIVATE_URL --port $PORT