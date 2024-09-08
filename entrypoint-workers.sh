#!/bin/sh

echo "Starting Workers server..."

if [ -z "$DATABASE_PRIVATE_URL" ]; then
  echo "DATABASE_PRIVATE_URL is not set. Exiting."
  exit 1
fi

if [ -z "$CF_ACCOUNT_ID" ]; then
  echo "CF_ACCOUNT_ID is not set. Exiting."
  exit 1
fi

if [ -z "$R2_ACCESS_KEY_ID" ]; then
  echo "R2_ACCESS_KEY_ID is not set. Exiting."
  exit 1
fi

if [ -z "$R2_ACCESS_KEY_SECRET" ]; then
  echo "R2_ACCESS_KEY_SECRET is not set. Exiting."
  exit 1
fi

/workers \
  --database-url $DATABASE_PRIVATE_URL \
  --dir /data \
  --object-store-endpoint https://$CF_ACCOUNT_ID.r2.cloudflarestorage.com \
  --object-store-region auto \
  --object-store-access-key-id $R2_ACCESS_KEY_ID \
  --object-store-access-key-secret $R2_ACCESS_KEY_SECRET \
  --cdn-base-url https://images.vscodethemes.com