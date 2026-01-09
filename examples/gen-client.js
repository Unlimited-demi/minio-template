const fs = require('fs');
const path = require('path');
const Minio = require('minio');
require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

// Configuration
const rootUser = process.env.MINIO_ROOT_USER || 'minioadmin';
const rootPass = process.env.MINIO_ROOT_PASSWORD || 'minioadmin123';
const serverUrl = process.env.MINIO_SERVER_URL || 'http://localhost:9000';
const targetLang = process.env.GEN_LANG || 'node'; // node | python | go

// Parse URL
let useSSL = false;
let endpoint = 'localhost';
let port = 9000;
try {
  const url = new URL(serverUrl);
  useSSL = url.protocol === 'https:';
  endpoint = url.hostname;
  port = url.port ? parseInt(url.port) : (useSSL ? 443 : 80);
} catch (e) {
  console.error("❌ Invalid MINIO_SERVER_URL");
}

// Initialize Client
const minioClient = new Minio.Client({
  endPoint: endpoint,
  port: port,
  useSSL: useSSL,
  accessKey: rootUser,
  secretKey: rootPass
});

async function main() {
  // 1. GENERATE CODE STRING
  let codeContent = '';
  let fileName = 'StorageService-Example.js';

  if (targetLang === 'python') {
    codeContent = getPythonCode(endpoint, useSSL, rootUser, rootPass);
    fileName = 'storage_example.py';
  } else if (targetLang === 'go') {
    codeContent = getGoCode(endpoint, useSSL, rootUser, rootPass);
    fileName = 'main.go';
  } else {
    codeContent = getNodeCode(endpoint, port, useSSL, rootUser, rootPass);
  }

  // 2. TEST CONNECTION & UPLOAD GENERATED CODE
  console.log(`\n🔮 MinIO Client Generator (Lang: ${targetLang.toUpperCase()})\n`);

  try {
    process.stdout.write(`1️⃣  Connecting to ${endpoint}... `);
    const buckets = await minioClient.listBuckets();
    console.log('✅ OK');
    
    if (buckets.length > 0) {
      const testBucket = buckets[0].name;
      
      process.stdout.write(`2️⃣  Uploading generated code to '${testBucket}/${fileName}'... `);
      await minioClient.putObject(testBucket, fileName, codeContent);
      console.log('✅ OK');
      
      const fileUrl = `${serverUrl.replace(/\/$/, '')}/${testBucket}/${fileName}`;
      console.log(`\n    👉 VIEW & COPY CODE HERE: ${fileUrl}`);
      console.log(`       (This verifies your server is working AND gives you the code!)`);

    } else {
      console.log('⚠️  No buckets found. Create one in .env to test uploads.');
    }

  } catch (err) {
    console.error(`\n❌ Error: ${err.message}`);
    console.log('   Check your .env credentials and ensure Docker is running.');
  }

  // 3. ALSO PRINT TO CONSOLE
  console.log('\n3️⃣  Console Output:');
  console.log('----------------------------------------------------');
  console.log(codeContent);
  console.log('----------------------------------------------------');
}

function getNodeCode(ep, port, ssl, user, pass) {
  return `/**
 * 📦 StorageService.js
 * 
 * A Production-Ready wrapper for MinIO.
 * Copy this file to your 'services/' or 'lib/' folder.
 */

const Minio = require('minio');

class StorageService {
  constructor() {
    this.client = new Minio.Client({
      endPoint: '${ep}',
      port: ${port},
      useSSL: ${ssl},
      accessKey: '${user}',
      secretKey: '${pass}'
    });
  }

  /**
   * 1. Get a Presigned URL for Frontend Uploads (Recommended)
   * Use this to let your specific frontend (React, Vue, HTML Form) upload DIRECTLY to MinIO.
   * Prevents loading your backend server with large file data.
   */
  async getUploadUrl(bucketName, filename, expiry = 3600) {
    // Generates a temporary URL valid for 1 hour
    return await this.client.presignedPutObject(bucketName, filename, expiry);
  }

  /**
   * 2. Upload a File directly from Backend
   * Use this if you are processing the file in the backend (e.g. from multer).
   * @param {string} bucketName 
   * @param {string} filename 
   * @param {Buffer|Stream} fileStream 
   * @param {number} size - Optional but recommended
   * @param {string} contentType - e.g. 'image/png'
   */
  async uploadFile(bucketName, filename, fileStream, size, contentType = 'application/octet-stream') {
    const metaData = {
        'Content-Type': contentType
    };
    return await this.client.putObject(bucketName, filename, fileStream, size, metaData);
  }

  /**
   * 3. Get a Download URL
   * - If bucket is PUBLIC: Returns the direct public link (CDN friendly).
   * - If bucket is PRIVATE: Returns a temporary secure link (Presigned GET).
   */
  async getFileUrl(bucketName, filename, isPublic = false) {
    if (isPublic) {
      const protocol = ${ssl} ? 'https' : 'http';
      const portStr = ${ssl} || (${port} === 80) ? '' : ':${port}';
      return \`\${protocol}://${ep}\${portStr}/\${bucketName}/\${filename}\`;
    }
    // Private: Generate signed link (valid for 1 hour)
    return await this.client.presignedGetObject(bucketName, filename, 3600);
  }

  /**
   * 4. List Files
   */
  async listFiles(bucketName) {
    const stream = this.client.listObjectsV2(bucketName);
    const files = [];
    for await (const obj of stream) {
      files.push(obj);
    }
    return files;
  }
}

// Export a singleton instance
module.exports = new StorageService();

/* ============================================================
   👇 EXAMPLE USAGE (Put this in your Controller / API Route)
   ============================================================

   const storage = require('./StorageService');

   // A. Frontend asks for upload URL
   app.get('/api/upload-url', async (req, res) => {
      const url = await storage.getUploadUrl('my-bucket', 'user-uploads/profile.jpg');
      res.json({ url }); 
      // Frontend then does: PUT {url} with the file body
   });

   // B. Backend uploads directly (e.g. from Multer)
   app.post('/api/upload', upload.single('file'), async (req, res) => {
      await storage.uploadFile('my-bucket', req.file.originalname, req.file.buffer);
      res.send('Uploaded!');
   });

   ============================================================ */
`;
}

function getPythonCode(ep, ssl, user, pass) {
  const protocol = ssl ? 'https' : 'http';
  const url = `${protocol}://${ep}`;
  return `import boto3
from botocore.client import Config

# 📦 MinIO Service (Python / Boto3)

class StorageService:
    def __init__(self):
        self.s3 = boto3.resource('s3',
                    endpoint_url='${url}',
                    aws_access_key_id='${user}',
                    aws_secret_access_key='${pass}',
                    config=Config(signature_version='s3v4'),
                    region_name='us-east-1')
        self.client = self.s3.meta.client

    def get_upload_url(self, bucket_name, object_name, expiry=3600):
        """Generate a presigned URL to upload a file"""
        return self.client.generate_presigned_url('put_object',
                                                  Params={'Bucket': bucket_name,
                                                          'Key': object_name},
                                                  ExpiresIn=expiry)

    def list_buckets(self):
        return [b.name for b in self.s3.buckets.all()]

# Usage
# service = StorageService()
# print(service.list_buckets())
`;
}

function getGoCode(ep, ssl, user, pass) {
  return `package main

import (
	"log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// 📦 MinIO Connection Setup

func Connect() (*minio.Client, error) {
	endpoint := "${ep}"
	accessKeyID := "${user}"
	secretAccessKey := "${pass}"
	useSSL := ${ssl}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	return minioClient, err
}

func main() {
    client, err := Connect()
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("✅ Connected to MinIO")
}
`;
}

main();
