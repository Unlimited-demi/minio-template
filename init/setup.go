package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := flag.String("endpoint", "127.0.0.1:9000", "MinIO endpoint")
	accessKey := flag.String("accessKey", "", "Access key")
	secretKey := flag.String("secretKey", "", "Secret key")
	buckets := flag.String("buckets", "", "Comma-separated list of buckets")
	genClient := flag.Bool("gen-client", false, "Generate client code")
	clientLang := flag.String("client-lang", "node", "Client language (node, python, go)")
    // Note: We need the EXTERNAL URL for the generated code, not the internal 127.0.0.1
    // We will try to guess or let user pass it, but for now we might have to reuse endpoint if it's not localhost
    // Actually, usually we want the localhost or domain from env. 
    // Let's stick to a placeholder or try to parse it. 
    // In this restricted setup, we'll just use the passed endpoint but replace 127.0.0.1 if needed?
    // User requested "logs location of file as per their url".
    // We'll pass the Public URL as a flag too.
    publicUrl := flag.String("public-url", "http://localhost:9000", "Public URL for the generated code")

	flag.Parse()

	if *buckets == "" {
		log.Println("No buckets to configure.")
		return
	}

	// Initialize MinIO client
	minioClient, err := minio.New(*endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*accessKey, *secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	bucketList := strings.Split(*buckets, ",")
    var firstBucket string

	for _, b := range bucketList {
		parts := strings.Split(b, ":")
		bucketName := strings.TrimSpace(parts[0])
		mode := "private"
		if len(parts) > 1 {
			mode = strings.TrimSpace(parts[1])
		}

		if bucketName == "" {
			continue
		}
        
        if firstBucket == "" {
            firstBucket = bucketName
        }

		// Create bucket if not exists
		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err != nil {
			log.Printf("Error checking bucket %s: %v\n", bucketName, err)
			continue
		}

		if !exists {
			fmt.Printf("Creating bucket: %s\n", bucketName)
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("Error creating bucket %s: %v\n", bucketName, err)
				continue
			}
		}

		// Apply Policy
		var policy string
		switch {
		case mode == "public":
			policy = fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {"AWS": ["*"]},
						"Action": ["s3:GetObject"],
						"Resource": ["arn:aws:s3:::%s/*"]
					}
				]
			}`, bucketName)
		case strings.HasPrefix(mode, "ip="):
			allowedIPs := strings.TrimPrefix(mode, "ip=")
			policy = fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Deny",
						"Principal": {"AWS": ["*"]},
						"Action": "s3:*",
						"Resource": ["arn:aws:s3:::%s/*", "arn:aws:s3:::%s"],
						"Condition": {
							"NotIpAddress": {
								"aws:SourceIp": [%s]
							}
						}
					}
				]
			}`, bucketName, bucketName, formatIPs(allowedIPs))
		case mode == "private":
			policy = ""
		}

		if policy != "" {
			err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
			if err != nil {
				log.Printf("Error setting policy for %s: %v\n", bucketName, err)
			} else {
				fmt.Printf("Applied policy '%s' to bucket: %s\n", mode, bucketName)
			}
		} else {
			_ = minioClient.SetBucketPolicy(ctx, bucketName, "") 
			fmt.Printf("Ensured bucket is private: %s\n", bucketName)
		}
	}

    // Generate Client Code if requested
    if *genClient && firstBucket != "" {
        generateAndUpload(ctx, minioClient, firstBucket, *clientLang, *accessKey, *secretKey, *publicUrl)
    }
}

func formatIPs(ipStr string) string {
	ips := strings.Split(ipStr, ";") 
	var quoted []string
	for _, ip := range ips {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, strings.TrimSpace(ip)))
	}
	return strings.Join(quoted, ",")
}

func generateAndUpload(ctx context.Context, client *minio.Client, bucket, lang, user, pass, pubUrl string) {
    fmt.Printf("\n🔮 Generating %s client code...\n", strings.ToUpper(lang))
    
    // Parse Public URL for templates
    useSSL := strings.HasPrefix(pubUrl, "https")
    endpoint := strings.TrimPrefix(pubUrl, "https://")
    endpoint = strings.TrimPrefix(endpoint, "http://")
    // Remove port if present to get hostname? No, keep it for now or split.
    // Simplifying parsing for template
    parts := strings.Split(endpoint, ":")
    host := parts[0]
    port := "80"
    if useSSL { port = "443" }
    if len(parts) > 1 {
        port = parts[1]
        // remove trailing slash
        port = strings.TrimSuffix(port, "/")
    }

    var content, filename string

    switch lang {
    case "python":
        filename = "storage_client.py"
        content = fmt.Sprintf(`import boto3
from botocore.client import Config

class StorageService:
    def __init__(self):
        self.s3 = boto3.client('s3',
            endpoint_url='%s',
            aws_access_key_id='%s',
            aws_secret_access_key='%s',
            config=Config(signature_version='s3v4'),
            region_name='us-east-1')

    # 1. Get Presigned Upload URL (Frontend -> MinIO)
    def get_upload_url(self, bucket, filename):
        return self.s3.generate_presigned_url('put_object', 
            Params={'Bucket': bucket, 'Key': filename}, ExpiresIn=3600)

    # 2. Upload File (Backend -> MinIO)
    def upload_file(self, bucket, filename, file_path):
        self.s3.upload_file(file_path, bucket, filename)

    # 3. Get Download URL (Public or Presigned)
    def get_file_url(self, bucket, filename):
        return self.s3.generate_presigned_url('get_object', 
            Params={'Bucket': bucket, 'Key': filename}, ExpiresIn=3600)

# Usage
storage = StorageService()
print("✅ Storage Service Initialized")
`, pubUrl, user, pass)

    case "go":
        filename = "main.go"
        content = fmt.Sprintf(`package main
import (
    "context"
    "log"
    "time"
    "net/url"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
    Client *minio.Client
}

func NewStorage() *StorageService {
    minioClient, err := minio.New("%s", &minio.Options{
        Creds:  credentials.NewStaticV4("%s", "%s", ""),
        Secure: %v,
    })
    if err != nil { log.Fatalln(err) }
    return &StorageService{Client: minioClient}
}

// 1. Get Presigned Upload URL
func (s *StorageService) GetUploadUrl(bucket, filename string) (string, error) {
    expiry := time.Hour * 1
    return s.Client.PresignedPutObject(context.Background(), bucket, filename, expiry)
}

// 2. Upload File
func (s *StorageService) UploadFile(bucket, filename, filepath string) (minio.UploadInfo, error) {
    return s.Client.FPutObject(context.Background(), bucket, filename, filepath, minio.PutObjectOptions{})
}

// 3. Get Download URL
func (s *StorageService) GetFileUrl(bucket, filename string) (string, error) {
    expiry := time.Hour * 1
    u, err := s.Client.PresignedGetObject(context.Background(), bucket, filename, expiry, nil)
    if err != nil { return "", err }
    return u.String(), nil
}

func main() {
    svc := NewStorage()
    log.Println("✅ Storage Service Initialized")
}`, endpoint, user, pass, useSSL)

    default: // node
        filename = "StorageService.js"
        content = fmt.Sprintf(`const Minio = require('minio');

class StorageService {
  constructor() {
    this.client = new Minio.Client({
      endPoint: '%s',
      port: %s,
      useSSL: %v,
      accessKey: '%s',
      secretKey: '%s'
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
module.exports = new StorageService();`, host, port, useSSL, user, pass)
    }

    reader := strings.NewReader(content)
    _, err := client.PutObject(ctx, bucket, filename, reader, int64(len(content)), minio.PutObjectOptions{
        ContentType: "application/javascript",
    })
    
    if err != nil {
        log.Printf("❌ Failed to upload client code: %v\n", err)
    } else {
        // Generate a PRESIGNED URL for the client code itself.
        // CRITICAL: We must use a client initialized with the PUBLIC URL to ensure
        // the signature matches the Host header the user will send.
        
        signingClient, err := minio.New(endpoint, &minio.Options{
            Creds:  credentials.NewStaticV4(user, pass, ""),
            Secure: useSSL,
        })
        
        var displayUrl string
        if err != nil {
             log.Printf("⚠️ Could not create signing client: %v\n", err)
             displayUrl = "Error generating link"
        } else {
             expiry := time.Hour * 24 
             presignedUrl, err := signingClient.PresignedGetObject(ctx, bucket, filename, expiry, nil)
             if err != nil {
                 log.Printf("⚠️ Could not presign client URL: %v\n", err)
                 displayUrl = "Error generating link"
             } else {
                 displayUrl = presignedUrl.String()
             }
        }

        fmt.Println("==================================================")
        fmt.Println("✅ CLIENT CODE GENERATED & UPLOADED")
        fmt.Println("👉 DOWNLOAD URL (Valid for 24h):")
        fmt.Printf("%s\n", displayUrl)
        fmt.Println("==================================================")
    }
}
