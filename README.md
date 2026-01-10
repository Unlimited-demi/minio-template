# 🚀 MinIO Production Template

> A secure, production-ready MinIO Object Storage template for Docker Compose.  
> Includes **Auto-SSL**, **Multi-Tenancy**, **IP Whitelisting**, and **Client Code Generation**.

---

## ✨ Features
- **Zero Config Start**: Defaults to localhost for instant local dev.
- **🔐 Multi-Bucket Security**: Configure Public, Private, and IP-Restricted buckets via `.env`.
- **📜 Client Code Generator**: Creates a production-ready `StorageService.js` custom-tailored to your config.
- **🛡️ Auto-SSL (Optional)**: Built-in Traefik integration for automatic Let's Encrypt HTTPS.
- **📂 File Browser (Optional)**: Web UI for non-technical users to manage files.
- **💾 Auto-Backup (Optional)**: Rclone sidecar to sync data to AWS/S3/Backblaze.

---

## 🚀 Quick Start

### 1. Requirements
- Docker & Docker Compose

### 2. Setup
Clone the repo and start:

```bash
# 1. Clone
git clone https://github.com/your-repo/minio-template.git
cd minio-template

# 2. Configure (Optional for local dev)
# Copy example env or just skip to use defaults
cp env-examples/.env.public .env

# 3. Run
docker-compose up -d --build
```

**That's it!**
- **Console**: [http://localhost:9001](http://localhost:9001) (User: `minioadmin`, Pass: `minioadmin123`)
- **API**: [http://localhost:9000](http://localhost:9000)

---

## 🛠️ Client Code Generator (Automated!)

You don't need to write connection code. We generate it for you—automatically on startup.

### Option A: Auto-Generate on Startup (Recommended)
Add this to your `.env` file:
```ini
GENERATE_CLIENT=true
CLIENT_LANG=node   # options: node, python, go
```
**What happens?**
1. When the container starts, it generates a custom `StorageService` class using your exact config.
2. It uploads this file to your first bucket.
3. It prints the URL in the logs.
   
**View the URL:**
```bash
docker-compose logs minio | grep "StorageService"
```

### Option B: Manual Script (Legacy)
If you prefer running it manually:
```bash
cd examples
npm install
node gen-client.js
```

---

## 📖 Configuration Guide

### 1. Bucket Security Modes
Define buckets in `.env` using `MINIO_BUCKETS`. Format: `name:mode`.

| Mode | Syntax | Description |
| :--- | :--- | :--- |
| **Public** | `public` | 🌍 Readable by anyone. Great for website assets/images. |
| **Private** | `private` | 🔒 No public access. authenticated APIs only. |
| **IP Restricted** | `ip=x.x.x.x` | 🏢 Only accessible from specific IPs (e.g., Office VPN). |

**Example `.env`:**
```ini
MINIO_BUCKETS=website-assets:public,finance-docs:private,internal-logs:ip=192.168.1.50
```

### 2. profiles (Advanced Features)
Enable optional features by setting the `COMPOSE_PROFILES` environment variable.

| Profile | Service | Description |
| :--- | :--- | :--- |
| `ssl` | **Traefik** | Automatic HTTPS (Let's Encrypt). Requires Domain + DNS. |
| `explorer` | **FileBrowser** | Web UI manager (Port 7070). |
| `backup` | **Rclone** | Syncs data to remote cloud storage. |

**How to Enable:**
```bash
# Linux / Mac
COMPOSE_PROFILES=ssl,explorer docker-compose up -d

# Windows PowerShell
$env:COMPOSE_PROFILES="ssl,explorer"; docker-compose up -d
```

---

## 🔒 Security Best Practices

### 1. Uploads
**NEVER** allow public write access.
- **Bad**: Allowing anonymous PUTs to your bucket.
- **Good**: Use **Presigned URLs**.
    1. Frontend requests URL from Backend.
    2. Backend generates a signed "Upload URL" (valid for 5 mins).
    3. Frontend uploads file directly to MinIO using that URL.
    *(The Generator script creates this logic for you!)*

### 2. Downloads
- **Public Buckets**: Use direct URLs (`https://minio.site.com/bucket/file.jpg`). Efficient & CDN friendly.
- **Private Buckets**: Use **Presigned GET URLs**. A temporary link that expires (e.g. 1 hour).

---

## ⚡ Examples

Check the `examples/` folder for manual scripts:
- [`1-public-bucket.js`](examples/1-public-bucket.js): Upload & Serve public assets.
- [`2-private-bucket.js`](examples/2-private-bucket.js): Secure private file sharing.
- [`3-mixed-ip-bucket.js`](examples/3-mixed-ip-bucket.js): Complex IP allowlisting.
- [`aws-sdk-example.js`](examples/aws-sdk-example.js): Using standard AWS S3 SDK.

---

## ❓ Troubleshooting

**Q: Connection Refused?**
- Ensure `MINIO_SERVER_URL` in `.env` has the correct protocol (`http://` vs `https://`).
- If running locally, simply unset `MINIO_SERVER_URL` to use the default `http://localhost:9000`.

**Q: "Access Denied" on Public Bucket?**
- Did you rebuild? Run `docker-compose up -d --build`.
- Only *Downloads* are public. Uploads always require keys.

**Q: SSL not working?**
- Did you set `ACME_EMAIL` in `.env`?
- Did you point DNS records (`minio.yourdomain.com`) to the server IP?
- Check Traefik logs: `docker-compose logs -f traefik`.
