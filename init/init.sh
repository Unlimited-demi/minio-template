#!/bin/sh
set -e

echo "▶ Configuring MinIO via Shell Script..."

# 1. Alias
mc alias set local http://127.0.0.1:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

# 2. Parse Buckets (comma separated)
# Format: "bucket1:mode,bucket2:mode"
if [ -n "$MINIO_BUCKETS" ]; then
  # split by comma
  old_ifs="$IFS"
  IFS=','
  first_bucket=""
  
  for bucket_conf in $MINIO_BUCKETS; do
    # split by colon
    bucket_name=$(echo "$bucket_conf" | cut -d':' -f1 | xargs)
    mode=$(echo "$bucket_conf" | cut -d':' -f2 | xargs)

    if [ -z "$bucket_name" ]; then continue; fi
    if [ -z "$first_bucket" ]; then first_bucket="$bucket_name"; fi

    echo "▶ Setting up bucket: $bucket_name (Mode: ${mode:-private})"
    
    # Create if not exists
    mc mb --ignore-existing local/"$bucket_name"

    # Policy
    if [ "$mode" = "public" ]; then
      mc anonymous set download local/"$bucket_name"
    else
      # Private (clear public access)
      mc anonymous set none local/"$bucket_name"
    fi
    
    # Note: Complex IP policies are hard in pure shell without JSON tools. 
    # skipping advanced IP logic for shell script version for simplicity, 
    # or user can use 'go' method for that.
  done
  IFS="$old_ifs"

  # 3. Client Generation
  if [ "$GENERATE_CLIENT" = "true" ] && [ -n "$first_bucket" ]; then
    echo "▶ Generating Client Code ($CLIENT_LANG)..."
    
    SERVER_URL="${MINIO_SERVER_URL:-http://localhost:9000}"
    # Strip protocol for JS/Go templates that need host/port split
    CLEAN_URL=$(echo "$SERVER_URL" | sed 's~http[s]*://~~g') 
    HOST=$(echo "$CLEAN_URL" | cut -d':' -f1)
    PORT=$(echo "$CLEAN_URL" | cut -d':' -f2)
    # If no port in string, infer from protocol
    if [ "$PORT" = "$HOST" ]; then
       if echo "$SERVER_URL" | grep -q "https"; then PORT=443; else PORT=80; fi
    fi
    SSL="false"
    if echo "$SERVER_URL" | grep -q "https"; then SSL="true"; fi

    FILE_NAME="StorageService.js"
    CONTENT=""

    if [ "$CLIENT_LANG" = "python" ]; then
      FILE_NAME="storage_client.py"
      CONTENT="import boto3
from botocore.client import Config

class StorageService:
    def __init__(self):
        self.s3 = boto3.client('s3',
            endpoint_url='$SERVER_URL',
            aws_access_key_id='$MINIO_ROOT_USER',
            aws_secret_access_key='$MINIO_ROOT_PASSWORD',
            config=Config(signature_version='s3v4'),
            region_name='us-east-1')

    # 1. Get Presigned Upload URL
    def get_upload_url(self, bucket, filename):
        return self.s3.generate_presigned_url('put_object', 
            Params={'Bucket': bucket, 'Key': filename}, ExpiresIn=3600)

    # 2. Upload File
    def upload_file(self, bucket, filename, file_path):
        self.s3.upload_file(file_path, bucket, filename)

    # 3. Get Download URL
    def get_file_url(self, bucket, filename):
        return self.s3.generate_presigned_url('get_object', 
            Params={'Bucket': bucket, 'Key': filename}, ExpiresIn=3600)

storage = StorageService()
print(\"✅ Storage Service Initialized\")"

    elif [ "$CLIENT_LANG" = "go" ]; then
      FILE_NAME="main.go"
      CONTENT="package main
import (
	\"context\"
	\"log\"
	\"time\"
	\"github.com/minio/minio-go/v7\"
	\"github.com/minio/minio-go/v7/pkg/credentials\"
)

type StorageService struct {
    Client *minio.Client
}

func NewStorage() *StorageService {
	minioClient, err := minio.New(\"$CLEAN_URL\", &minio.Options{
		Creds:  credentials.NewStaticV4(\"$MINIO_ROOT_USER\", \"$MINIO_ROOT_PASSWORD\", \"\"),
		Secure: $SSL,
	})
	if err != nil { log.Fatalln(err) }
    return &StorageService{Client: minioClient}
}

func (s *StorageService) GetUploadUrl(bucket, filename string) (string, error) {
    expiry := time.Hour * 1
    return s.Client.PresignedPutObject(context.Background(), bucket, filename, expiry)
}

func (s *StorageService) UploadFile(bucket, filename, filepath string) (minio.UploadInfo, error) {
    return s.Client.FPutObject(context.Background(), bucket, filename, filepath, minio.PutObjectOptions{})
}

func (s *StorageService) GetFileUrl(bucket, filename string) (string, error) {
    expiry := time.Hour * 1
    u, err := s.Client.PresignedGetObject(context.Background(), bucket, filename, expiry, nil)
    if err != nil { return \"\", err }
    return u.String(), nil
}

func main() {
    NewStorage()
    log.Println(\"✅ Storage Service Initialized\")
}"

    else # node
      CONTENT="const Minio = require('minio');

class StorageService {
  constructor() {
    this.client = new Minio.Client({
      endPoint: '$HOST',
      port: $PORT,
      useSSL: $SSL,
      accessKey: '$MINIO_ROOT_USER',
      secretKey: '$MINIO_ROOT_PASSWORD'
    });
  }

  // 1. Get Presigned Upload URL (Frontend -> MinIO)
  async getUploadUrl(bucket, filename) {
    return await this.client.presignedPutObject(bucket, filename, 3600);
  }

  // 2. Upload File (Backend -> MinIO)
  async uploadFile(bucket, filename, fileStream, metaData = {}) {
    return await this.client.putObject(bucket, filename, fileStream, null, metaData);
  }

  // 3. Get Download URL (Public or Presigned)
  async getFileUrl(bucket, filename) {
    // Basic logic: return signed URL for safety
    return await this.client.presignedGetObject(bucket, filename, 3600);
  }
}
module.exports = new StorageService();"
    fi

    # Write & Upload
    echo "$CONTENT" > /tmp/$FILE_NAME
    mc cp /tmp/$FILE_NAME local/"$first_bucket"/$FILE_NAME
    
    echo "=================================================="
    echo "✅ CLIENT CODE GENERATED & UPLOADED"
    echo "👉 URL: $SERVER_URL/$first_bucket/$FILE_NAME"
    echo "=================================================="
  fi

else 
  echo "⚠️ MINIO_BUCKETS not defined"
fi
