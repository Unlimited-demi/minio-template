const { S3Client, PutObjectCommand, GetObjectCommand } = require("@aws-sdk/client-s3");
const { getSignedUrl } = require("@aws-sdk/s3-request-presigner");

/**
 * 📦 USING STANDARD AWS SDK (v3)
 * Use this if you are already familiar with AWS S3.
 * MinIO is 100% compatible!
 */

// 1. Initialize Client
const s3 = new S3Client({
  endpoint: "https://minio.yourdomain.com", // Your MinIO URL
  region: "us-east-1", // MinIO requires a region, but ignores the value
  credentials: {
    accessKeyId: "admin",
    secretAccessKey: "SuperStrongPassword123!"
  },
  forcePathStyle: true // ⚠️ CRITICAL: Must be true for MinIO/Self-hosted S3!
});

async function main() {
  const bucketName = 'company-documents';
  const objectName = 'contract-v2.pdf';

  try {
    console.log(`\n--- AWS SDK Example (${bucketName}) ---\n`);

    // SCENARIO: Secure Upload (PUT)
    const putCommand = new PutObjectCommand({
      Bucket: bucketName,
      Key: objectName,
    });
    
    // Generate signed URL for 1 hour (3600 seconds)
    const uploadUrl = await getSignedUrl(s3, putCommand, { expiresIn: 3600 });
    
    console.log('1️⃣  AWS S3 UPLOAD URL:');
    console.log(`   ${uploadUrl}\n`);


    // SCENARIO: Secure Download (GET)
    const getCommand = new GetObjectCommand({
      Bucket: bucketName,
      Key: objectName,
    });
    
    // Generate signed URL for 15 minutes
    const downloadUrl = await getSignedUrl(s3, getCommand, { expiresIn: 900 });

    console.log('2️⃣  AWS S3 DOWNLOAD URL:');
    console.log(`   ${downloadUrl}`);

  } catch (err) {
    console.error('Error:', err);
  }
}

main();
