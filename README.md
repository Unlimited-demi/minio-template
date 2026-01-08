version: "3.8"

services:
  minio:
    build: .
    container_name: mini_minio
    volumes:
      - minio-data:/data
    environment:
      # ROOT CREDS
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123

      # BUCKET CONFIG
      MINIO_BUCKET_NAME: anivault
      MINIO_PUBLIC: "true"

      # DOMAIN CONFIG (REQUIRED)
      MINIO_SERVER_URL: https://storage.getmusterup.com
      MINIO_BROWSER_REDIRECT_URL: https://console.getmusterup.com
    command: server /data --console-address ":9001"
    networks:
      - minio_net
    restart: unless-stopped

volumes:
  minio-data:

networks:
  minio_net:
