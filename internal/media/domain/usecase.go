package domain

import (
	"context"
	"io"
)

// MediaUseCase defines the business logic for media files
type MediaUseCase interface {
	UploadFile(ctx context.Context, reader io.Reader, size int64, contentType string) (*Media, error)
	GetFile(ctx context.Context, fileName string) (io.ReadCloser, string, error)
}
