package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3MediaRepository struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	secure     bool
}

// NewS3MediaRepository initializes a new MinIO/S3 client
func NewS3MediaRepository(endpoint, accessKey, secretKey, bucketName string) (*S3MediaRepository, error) {
	// Initialize minio client object.
	// For local development with MinIO, useSSL is false. For production AWS S3, it should be true.
	useSSL := false
	if endpoint != "minio:9000" && endpoint != "localhost:9000" && endpoint != "api.yolyolakay.com" && endpoint != "api.yolyolakay.xyz" {
		useSSL = true
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	return &S3MediaRepository{
		client:     minioClient,
		bucketName: bucketName,
		endpoint:   endpoint,
		secure:     useSSL,
	}, nil
}

// Upload streams the file to S3 and returns the public URL
func (r *S3MediaRepository) Upload(ctx context.Context, fileName string, reader io.Reader, size int64, contentType string) (string, error) {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := r.client.PutObject(ctx, r.bucketName, fileName, reader, size, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload object to s3: %w", err)
	}

	// Assuming the bucket is public, we can construct the URL directly
	// For public URLs we use api.yolyolakay.com or configured public domain
	publicEndpoint := os.Getenv("S3_PUBLIC_DOMAIN")
	if publicEndpoint == "" {
		if r.endpoint == "minio:9000" || r.endpoint == "localhost:9000" {
			publicEndpoint = "localhost:9000"
		} else {
			publicEndpoint = r.endpoint
		}
	}

	scheme := "http"
	if r.secure || (publicEndpoint != "localhost:9000" && publicEndpoint != "minio:9000" && !strings.HasPrefix(publicEndpoint, "127.0.0.1")) {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, publicEndpoint, r.bucketName, fileName), nil
}

// GetFile retrieves an object from S3
func (r *S3MediaRepository) GetFile(ctx context.Context, fileName string) (io.ReadCloser, string, error) {
	obj, err := r.client.GetObject(ctx, r.bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from s3: %w", err)
	}
	stat, err := obj.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("failed to stat object: %w", err)
	}
	return obj, stat.ContentType, nil
}
