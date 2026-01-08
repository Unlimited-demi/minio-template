#!/bin/sh
set -e

echo "▶ Configuring bucket via shell script..."

# Create alias
mc alias set local http://127.0.0.1:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

# Create bucket if missing
if [ -n "$MINIO_BUCKET_NAME" ]; then
  echo "▶ Creating bucket: $MINIO_BUCKET_NAME"
  mc mb --ignore-existing local/$MINIO_BUCKET_NAME
fi

# Public access
if [ "$MINIO_PUBLIC" = "true" ]; then
  echo "▶ Setting bucket public (anonymous read)"
  mc anonymous set download local/$MINIO_BUCKET_NAME
fi
