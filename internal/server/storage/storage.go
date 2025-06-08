package storage

import "context"

// Storage defines the interface for file storage operations
type Storage interface {
	// UploadFile uploads a file to storage and returns its public URL
	UploadFile(ctx context.Context, filePath string) (string, error)

	// DownloadFile downloads a file from storage to the specified path
	DownloadFile(ctx context.Context, url, filePath string) error

	// DeleteFile deletes a file from storage
	DeleteFile(ctx context.Context, url string) error
}
