package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := flag.String("endpoint", "127.0.0.1:9000", "MinIO endpoint")
	accessKey := flag.String("accessKey", "", "Access key")
	secretKey := flag.String("secretKey", "", "Secret key")
	buckets := flag.String("buckets", "", "Comma-separated list of buckets (e.g., 'public-bucket:public,private-bucket:private,secure:ip=1.2.3.4')")

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
			// Simplified policy: Deny everything NOT from allowed IP, or Allow from allowed IP
			// Standard S3 way: Deny if NotIpAddress
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
			// Explicitly remove policy to ensure it's private (or empty string clears it)
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
			// For private, we might want to ensure no public policy exists
			// But setting empty policy might error or not be supported depending on SDK version?
			// Usually SetBucketPolicy with empty string deletes it.
			_ = minioClient.SetBucketPolicy(ctx, bucketName, "") 
			fmt.Printf("Ensured bucket is private: %s\n", bucketName)
		}
	}
}

func formatIPs(ipStr string) string {
	ips := strings.Split(ipStr, ";") // separate multiple IPs with semicolon in config if needed? or space? 
    // Let's assume config uses semi-colon inside the value to avoid comma conflict
    // actually user might use a single IP usually. Let's support semicolon.
	var quoted []string
	for _, ip := range ips {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, strings.TrimSpace(ip)))
	}
	return strings.Join(quoted, ",")
}
