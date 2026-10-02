package domain

import (
	"context"
	"io"
)

// MediaRepository defines the data access layer for media files (e.g. S3)
type MediaRepository interface {
	Upload(ctx context.Context, fileName string, reader io.Reader, size int64, contentType string) (string, error)
	GetFile(ctx context.Context, fileName string) (io.ReadCloser, string, error)
}
