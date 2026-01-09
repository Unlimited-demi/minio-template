#!/bin/sh
set -e

echo "▶ Starting MinIO in background..."
minio server /data --console-address ":9001" &

# Wait for MinIO to be ready
sleep 5

# Choose init method
case "$MINIO_INIT_METHOD" in
  sh)
    echo "▶ Using shell script to configure bucket"
    /init/init.sh
    ;;
  go)
    echo "▶ Using Go script to configure bucket"
    if [ -n "$MINIO_BUCKETS" ]; then
      /setup-bucket \
        -endpoint "127.0.0.1:9000" \
        -accessKey "$MINIO_ROOT_USER" \
        -secretKey "$MINIO_ROOT_PASSWORD" \
        -buckets "$MINIO_BUCKETS" \
        -gen-client="${GENERATE_CLIENT:-false}" \
        -client-lang="${CLIENT_LANG:-node}" \
        -public-url="${MINIO_SERVER_URL:-http://localhost:9000}"
    fi
    ;;
  *)
    echo "⚠️  No valid MINIO_INIT_METHOD set, skipping bucket setup"
    ;;
esac

echo "✅ MinIO ready"
wait
