const Minio = require('minio');

/**
 * SCENARIO B: Private Bucket (e.g., Company Documents, Invoices)
 * 
 * Goal: 
 * 1. Uploads are secure.
 * 2. Downloads LIMIT access to authorized users only (Temporary URLs).
 */

const minioClient = new Minio.Client({
  endPoint: 'minio.yourdomain.com',
  port: 443,
  useSSL: true,
  accessKey: 'admin',
  secretKey: 'SuperStrongPassword123!'
});

const BUCKET_NAME = 'company-documents';

async function main() {
  const objectName = 'confidential-contract.pdf';

  try {
    console.log(`\n--- Scenario B: Private Bucket (${BUCKET_NAME}) ---\n`);

    // STEP 1: Secure Upload
    const uploadUrl = await minioClient.presignedPutObject(BUCKET_NAME, objectName, 60 * 60);
    console.log('1️⃣  UPLOAD (PUT):');
    console.log(`   ${uploadUrl}\n`);

    // STEP 2: Secure Download (Presigned GET)
    // The public URL `https://.../bucket/file` will return 403 Forbidden.
    // We must generate a signed URL. specific expiry (e.g., 15 minutes).
    const downloadUrl = await minioClient.presignedGetObject(BUCKET_NAME, objectName, 15 * 60);
    
    console.log('2️⃣  DOWNLOAD (GET): Temporary Access Link (15 mins):');
    console.log(`   ${downloadUrl}`);
    console.log('   (Share this link only with authorized users)');

  } catch (err) {
    console.error('Error:', err.message);
  }
}

main();
