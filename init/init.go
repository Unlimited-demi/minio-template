package main

import (
    "flag"
    "fmt"
    "log"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "context"
)

func main() {
    endpoint := flag.String("endpoint", "127.0.0.1:9000", "MinIO endpoint")
    accessKey := flag.String("accessKey", "", "Access key")
    secretKey := flag.String("secretKey", "", "Secret key")
    bucket := flag.String("bucket", "", "Bucket name")

    flag.Parse()

    if *bucket == "" {
        log.Fatal("Bucket name is required")
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

    // Create bucket if not exists
    exists, err := minioClient.BucketExists(ctx, *bucket)
    if err != nil {
        log.Fatalln(err)
    }

    if !exists {
        fmt.Println("Creating bucket:", *bucket)
        err = minioClient.MakeBucket(ctx, *bucket, minio.MakeBucketOptions{})
        if err != nil {
            log.Fatalln(err)
        }
    }

    // Set public read policy
    policy := `{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": {"AWS": ["*"]},
          "Action": ["s3:GetObject"],
          "Resource": ["arn:aws:s3:::` + *bucket + `/*"]
        }
      ]
    }`

    err = minioClient.SetBucketPolicy(ctx, *bucket, policy)
    if err != nil {
        log.Fatalln(err)
    }

    fmt.Println("Bucket is now public:", *bucket)
}
