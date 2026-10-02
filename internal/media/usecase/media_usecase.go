package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/media/domain"
)

type mediaUseCase struct {
	repo domain.MediaRepository
}

func NewMediaUseCase(repo domain.MediaRepository) domain.MediaUseCase {
	return &mediaUseCase{
		repo: repo,
	}
}

func (uc *mediaUseCase) UploadFile(ctx context.Context, reader io.Reader, size int64, originalFilename string) (*domain.Media, error) {
	// 1. Validate Size (e.g. max 5MB)
	const maxFileSize = 5 * 1024 * 1024
	if size > maxFileSize {
		return nil, errors.New("file size exceeds maximum allowed limit of 5MB")
	}

	// 2. Validate Extension / Determine Content Type
	ext := strings.ToLower(filepath.Ext(originalFilename))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".pdf":
		contentType = "application/pdf"
	default:
		return nil, errors.New("unsupported file format. Allowed: jpg, jpeg, png, pdf")
	}

	// Read first 512 bytes for Magic Bytes validation
	buffer := make([]byte, 512)
	n, err := io.ReadFull(reader, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	detectedContentType := http.DetectContentType(buffer[:n])
	if !strings.HasPrefix(detectedContentType, "image/jpeg") &&
		!strings.HasPrefix(detectedContentType, "image/png") &&
		!strings.HasPrefix(detectedContentType, "application/pdf") {
		return nil, errors.New("invalid file content: malicious or unsupported file detected")
	}

	// Reconstruct the reader
	reader = io.MultiReader(bytes.NewReader(buffer[:n]), reader)

	// 3. Generate Unique Filename
	fileID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", fileID, ext)

	// 4. Upload to Repository (S3)
	_, err = uc.repo.Upload(ctx, fileName, reader, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}

	// 5. Construct Gateway URL
	baseURL := os.Getenv("PUBLIC_API_URL")
	if baseURL == "" {
		baseURL = "https://api.yolyolakay.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	url := fmt.Sprintf("%s/api/v1/media/%s", baseURL, fileName)

	// 6. Return Media Object
	return &domain.Media{
		ID:          fileID,
		URL:         url,
		Size:        size,
		ContentType: contentType,
	}, nil
}

// GetFile retrieves a file from the repository
func (uc *mediaUseCase) GetFile(ctx context.Context, fileName string) (io.ReadCloser, string, error) {
	return uc.repo.GetFile(ctx, fileName)
}
