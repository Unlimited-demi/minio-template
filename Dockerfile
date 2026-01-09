FROM golang:1.24-alpine as builder

WORKDIR /app

# Copy Go files
COPY init/setup.go .
COPY init/go.mod .
COPY init/go.sum .

# Build the Go binary
RUN go build -o /setup-bucket setup.go

# Final image: MinIO + setup scripts
FROM minio/minio:latest

COPY --from=builder /setup-bucket /setup-bucket
COPY init/ /init/

RUN chmod +x /init/*.sh /setup-bucket

ENTRYPOINT ["/init/entrypoint.sh"]
