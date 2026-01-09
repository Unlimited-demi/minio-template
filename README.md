# MinIO Pro Quickstart Template

A production-ready MinIO template designed for **Multi-Tenancy** and **Security**. Deploy on any VPS in seconds.

## Features
- 🚀 **One-click deployment** with Docker Compose
- 📦 **Multi-Bucket Support** (Public, Private, IP-Restricted)
- 🔒 **Security First**: Easy IP allowlisting for sensitive buckets
- 🔧 **Go-based Initialization** for robustness
- 📝 **Easy Configuration** via `.env` file

## Quick Start
1. **Clone/Download** this repo.
2. **Configure `.env`** (see below).
3. **Run**: `docker-compose up -d --build`

---

## 📖 Configuration Guide & Scenarios

### 1. Base Configuration (Required for all)
Create a `.env` file. These settings apply to the server itself.

```ini
# Credentials
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=SuperStrongPassword123!

# Domains (⚠️ MUST include http:// or https://)
MINIO_SERVER_URL=https://minio.yourdomain.com
MINIO_BROWSER_REDIRECT_URL=https://console.yourdomain.com
```

### 2. Bucket Strategy (New!)
Define your buckets and their security mode in `MINIO_BUCKETS`.
Format: `bucket-name:mode` separated by commas.

| Mode | Syntax | Description |
| :--- | :--- | :--- |
| **Public** | `public` | Readable by anyone on the internet. |
| **Private** | `private` | No public access. Keys/Presigned URLs only. |
| **IP Restricted** | `ip=1.2.3.4` | Only allowed from specific IPs. |
| **Multi-IP** | `ip=1.2.3.4;5.6.7.8` | **Separate multiple IPs with a semicolon (;)**. |

**Example:**
```ini
# A complicated setup:
# 1. public-assets: Public
# 2. company-files: Private
# 3. office-share: Allowed from 2 different office IPs
MINIO_BUCKETS=public-assets:public,company-files:private,office-share:ip=203.0.113.5;198.51.100.2
```

👉 **See [`env-examples/`](env-examples/) for ready-to-copy `.env` files.**

#### Scenario A: Simple Public Storage (Website Assets)
**Goal**: Store images/videos that anyone can view via a URL.
- **Mode**: `public`
- **Uploads**: Secure (via Admin user or Presigned URL).
- **Downloads**: Public (Open URL).
```ini
MINIO_BUCKETS=website-assets:public
```
*Access*: `https://minio.yourdomain.com/website-assets/image.jpg`

#### Scenario B: Secure Company Files
**Goal**: Internal documents that should NEVER be public.
- **Mode**: `private`
- **Uploads**: Authenticated only.
- **Downloads**: Authenticated only (or via short-lived Presigned URL).
```ini
MINIO_BUCKETS=company-documents:private
```

#### Scenario C: Internal LMS / Office Only
**Goal**: Videos/Files accessible ONLY from your office VPN or LMS Server.
- **Mode**: `ip=SEARCH_IP`
- **Uploads**: Authenticated.
- **Downloads**: Authenticated OR restricted by IP.
```ini
MINIO_BUCKETS=lms-videos:ip=203.0.113.5
```

#### Scenario D: Mixed (The "Hybrid" App)
**Goal**: A mix of all the above.
```ini
MINIO_BUCKETS=public-assets:public,user-data:private,internal-logs:ip=10.0.0.5
```

---

## 🔒 Security & Presigned URLs

### "Does a Public Bucket need Presigned URLs?"
- **For Reading (Downloading)**: **NO**. You can just use the direct link: `https://domain.com/bucket/file`.
- **For Writing (Uploading)**: **YES**. You should **never** allow anonymous uploads. Your backend should generate a **Presigned PUT URL** so the frontend can upload directly to MinIO safely.

### Example Usage
See [`examples/`](examples/) for full Node.js code samples.

| Scenario | File | Description |
| :--- | :--- | :--- |
| **A: Public** | [`1-public-bucket.js`](examples/1-public-bucket.js) | Presigned Upload + Public URL |
| **B: Private** | [`2-private-bucket.js`](examples/2-private-bucket.js) | Presigned Upload + Presigned Download |
| **C: IP / Mixed** | [`3-mixed-ip-bucket.js`](examples/3-mixed-ip-bucket.js) | Hybrid Access (Direct vs Remote) |
| **AWS SDK** | [`aws-sdk-example.js`](examples/aws-sdk-example.js) | **Using Standard AWS SDK v3** |

#### Running the Examples
```bash
cd examples
npm install
node aws-sdk-example.js
```

> **Note on AWS SDK**: You absolutely CAN use the standard `aws-sdk`. Just ensure you set `forcePathStyle: true` in the client config. See `aws-sdk-example.js` for details.

## Troubleshooting
- **URL Scheme Error**: Ensure `MINIO_SERVER_URL` starts with `http://` or `https://`.
- **Public Access Denied**: Check if you rebuilt the container (`docker-compose up -d --build`) after changing `.env`.
