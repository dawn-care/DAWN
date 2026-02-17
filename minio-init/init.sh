#!/bin/sh

# Wait for MinIO to be ready
sleep 10

# Configure MinIO client
mc alias set minio http://minio:9000 localuser localpass123

# Create the bucket
mc mb minio/media

# Upload files
mc cp --recursive /files/ minio/media/

echo "MinIO initialization complete!"
