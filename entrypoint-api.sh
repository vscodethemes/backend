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

if [ -z "$PUBLIC_KEY" ]; then
  echo "PUBLIC_KEY is not set. Exiting."
  exit 1
fi

if [ -z "$ISSUER" ]; then
  echo "ISSUER is not set. Exiting."
  exit 1
fi

# Write the public key to a file
echo "$PUBLIC_KEY" > ./key.rsa.pub

/api \
  --database-url $DATABASE_PRIVATE_URL \
  --port $PORT \
  --public-key-path ./key.rsa.pub \
  --issuer $ISSUER