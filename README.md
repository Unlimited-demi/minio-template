# MinIO Pro Quickstart Template

A production-ready MinIO template with automatic bucket provisioning and public access configuration, designed for easy deployment on VPS.

## Features
- 🚀 **One-click deployment** with Docker Compose
- 📦 **Automatic Bucket Creation**
- 🔓 **Auto-Public Access** (optional)
- 🔧 **Go-based Initialization** for robustness
- 📝 **Easy Configuration** via `.env` file

## Quick Start

1. **Clone the repository** (or download files to your VPS).
2. **Create a `.env` file** based on the example below.
3. **Build and Run**:
   ```bash
   docker-compose up -d --build
   ```

## Configuration (.env)

Create a `.env` file in the root directory.

**⚠️ IMPORTANT:** `MINIO_SERVER_URL` **MUST** include the protocol (`http://` or `https://`).

```ini
# ======================
# MinIO Root Credentials
# ======================
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=ChangeThisToSomethingStrong123!

# ======================
# Bucket Configuration
# ======================
MINIO_BUCKET_NAME=my-public-bucket
# Set to 'true' to make the bucket readable by anyone (anonymous download)
MINIO_PUBLIC=true

# ======================
# Domain / URL Settings
# ======================
# ⚠️ MUST include http:// or https://
MINIO_SERVER_URL=https://minio.example.com

# Console redirect URL (The UI you access in browser)
MINIO_BROWSER_REDIRECT_URL=https://console.example.com

# ======================
# Init Strategy
# ======================
# Options: go | sh
# 'go' is recommended for better reliability
MINIO_INIT_METHOD=go
```

## Troubleshooting

- **Invalid MINIO_SERVER_URL**: Ensure you included `http://` or `https://` in the `MINIO_SERVER_URL` variable.
- **Endpoint url cannot have fully qualified paths**: This issue has been fixed in the latest version of this template. Ensure you rebuild your image (`docker-compose build --no-cache`).

## Directory Structure
- `docker-compose.yml`: Main service definition.
- `init/`: Initialization scripts (`entrypoint.sh`, `setup.go`).
- `Dockerfile`: Custom image definition.
