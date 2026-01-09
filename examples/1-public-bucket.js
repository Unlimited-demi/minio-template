const Minio = require('minio');

/**
 * SCENARIO A: Public Bucket (e.g., Website Assets)
 * 
 * Goal: 
 * 1. Allow users to upload files securely (via Presigned URL).
 * 2. Serve files directly via public URL (efficient, CDN-friendly).
 */

// 1. Initialize Client
const minioClient = new Minio.Client({
  endPoint: 'minio.yourdomain.com', // Replace with your domain
  port: 443,
  useSSL: true,
  accessKey: 'admin',      // ⚠️ Use environment variables in production!
  secretKey: 'SuperStrongPassword123!'
});

const BUCKET_NAME = 'public-assets';

async function main() {
  const objectName = 'user-avatar-123.jpg';
  
  try {
    console.log(`\n--- Scenario A: Public Bucket (${BUCKET_NAME}) ---\n`);

    // STEP 1: Generate Upload URL (Presigned PUT)
    // This allows the frontend to upload 'user-avatar-123.jpg' directly to the bucket.
    // The link is valid for 1 hour.
    const uploadUrl = await minioClient.presignedPutObject(BUCKET_NAME, objectName, 60 * 60);
    
    console.log('1️⃣  UPLOAD: Frontend uses this URL to PUT the file:');
    console.log(`   ${uploadUrl}`);
    console.log(`   Command: curl -X PUT -T ./my-image.jpg "${uploadUrl}"\n`);

    // STEP 2: Public Download URL
    // Since the bucket is PUBLIC, you don't need to sign the URL.
    // You just construct it manually.
    const publicUrl = `https://minio.yourdomain.com/${BUCKET_NAME}/${objectName}`;
    
    console.log('2️⃣  DOWNLOAD: Publicly accessible URL (Permanent):');
    console.log(`   ${publicUrl}`);

  } catch (err) {
    console.error('Error:', err.message);
  }
}

main();
