const Minio = require('minio');

/**
 * SCENARIO C: IP-Restricted / Mixed (e.g., Internal LMS)
 * 
 * Goal:
 * 1. Users INSIDE the office/VPN (Allowed IP) can access files directly.
 * 2. Users OUTSIDE (e.g., Remote Workers) need a secure temporary link.
 */

const minioClient = new Minio.Client({
  endPoint: 'minio.yourdomain.com',
  port: 443,
  useSSL: true,
  accessKey: 'admin',
  secretKey: 'SuperStrongPassword123!'
});

const BUCKET_NAME = 'lms-videos';

async function main() {
  const objectName = 'training-session.mp4';

  try {
    console.log(`\n--- Scenario C: IP-Restricted Bucket (${BUCKET_NAME}) ---\n`);

    // STEP 1: Secure Upload (Always needed)
    const uploadUrl = await minioClient.presignedPutObject(BUCKET_NAME, objectName, 3600);
    console.log('1️⃣  UPLOAD (PUT):');
    console.log(`   ${uploadUrl}\n`);

    // STEP 2: Internal Access (Direct)
    // If the user's IP is in the allowed list (e.g., Office Wi-Fi), this works.
    const internalUrl = `https://minio.yourdomain.com/${BUCKET_NAME}/${objectName}`;
    console.log('2️⃣  INTERNAL ACCESS (Office IP):');
    console.log(`   ${internalUrl}`);
    console.log('   (Works automatically if user is on allowed IP)\n');

    // STEP 3: Remote Access (Presigned GET)
    // If user is at home (wrong IP), the direct link gives 403.
    // Generate a signed URL to BYPASS the IP restriction temporarily.
    const remoteUrl = await minioClient.presignedGetObject(BUCKET_NAME, objectName, 3600);
    console.log('3️⃣  EXTERNAL ACCESS (Remote Worker):');
    console.log(`   ${remoteUrl}`);
    console.log('   (Bypasses IP check for 1 hour)');

  } catch (err) {
    console.error('Error:', err.message);
  }
}

main();
