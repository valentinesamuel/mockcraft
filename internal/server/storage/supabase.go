package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	supa "github.com/supabase-community/supabase-go"
)

// SupabaseStorage handles file operations with Supabase storage
type SupabaseStorage struct {
	client *supa.Client
	bucket string
}

// NewSupabaseStorage creates a new SupabaseStorage instance
func NewSupabaseStorage(url, key, bucket string) (*SupabaseStorage, error) {
	// Ensure we're using the service role key (should start with 'eyJ...')
	if len(key) < 10 || key[:3] != "eyJ" {
		return nil, fmt.Errorf("invalid service role key format")
	}

	client, err := supa.NewClient(url, key, &supa.ClientOptions{
		Headers: map[string]string{
			"Authorization": "Bearer " + key,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	// Verify bucket exists
	_, err = client.Storage.GetBucket(bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket '%s' not found or not accessible: %w", bucket, err)
	}

	return &SupabaseStorage{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadFile uploads a file to Supabase storage and returns its public URL
func (s *SupabaseStorage) UploadFile(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fmt.Println("🗑️ 🗑️ bucket", s.bucket)

	fileName := filepath.Base(filePath)
	fmt.Println("📍📍 fileName", fileName)
	fmt.Println("📧📧 file", file)
	_, err = s.client.Storage.UploadFile(s.bucket, fileName, file)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Get the public URL
	url := s.client.Storage.GetPublicUrl(s.bucket, fileName)
	fmt.Println("🗑️ 🗑️ url", url.SignedURL)
	return url.SignedURL, nil
}

// DownloadFile downloads a file from Supabase storage
func (s *SupabaseStorage) DownloadFile(ctx context.Context, url, filePath string) error {
	// Extract the file name from the URL
	fileName := filepath.Base(url)

	// Download the file
	data, err := s.client.Storage.DownloadFile(s.bucket, fileName)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write the data to the file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeleteFile deletes a file from Supabase storage
func (s *SupabaseStorage) DeleteFile(ctx context.Context, url string) error {
	fileName := filepath.Base(url)
	_, err := s.client.Storage.RemoveFile(s.bucket, []string{fileName})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
