package utils

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// setupMockDB initializes a mock database
func SetupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

// SetupMockMinIO initializes a mock MinIO BLOB database
func SetupMockMinIO(t *testing.T) *minio.Client {
	endpoint := "localhost:9000"
	accessKey := "localuser"
	secretKey := "localpass123"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		t.Fatalf("Failed to initialize Local MinIO client: %v", err)
	}

	testAudioBucket := "test-audio"
	testAudioExists, err := minioClient.BucketExists(context.Background(), testAudioBucket)
	if err != nil {
		t.Fatalf("Failed to check if bucket exists: %v", err)
	}

	if !testAudioExists {
		err = minioClient.MakeBucket(context.Background(), testAudioBucket, minio.MakeBucketOptions{})
		if err != nil {
			t.Fatalf("Failed to create bucket '%s': %v", testAudioBucket, err)
		}
	}

	testVideoBucket := "test-video"
	testVideoExists, err := minioClient.BucketExists(context.Background(), testVideoBucket)
	if err != nil {
		t.Fatalf("Failed to check if bucket exists: %v", err)
	}

	if !testVideoExists {
		err = minioClient.MakeBucket(context.Background(), testVideoBucket, minio.MakeBucketOptions{})
		if err != nil {
			t.Fatalf("Failed to create bucket '%s': %v", testVideoBucket, err)
		}
	}

	testPhotoBucket := "test-photo"
	testPhotoExists, err := minioClient.BucketExists(context.Background(), testPhotoBucket)
	if err != nil {
		t.Fatalf("Failed to check if bucket exists: %v", err)
	}

	if !testPhotoExists {
		err = minioClient.MakeBucket(context.Background(), testPhotoBucket, minio.MakeBucketOptions{})
		if err != nil {
			t.Fatalf("Failed to create bucket '%s': %v", testPhotoBucket, err)
		}
	}
	return minioClient
}
