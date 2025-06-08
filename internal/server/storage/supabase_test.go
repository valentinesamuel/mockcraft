package storage

import (
	"context"
	"os"
	"testing"
)

func TestSupabaseStorage(t *testing.T) {
	// Skip if Supabase credentials are not set
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	if url == "" || key == "" {
		t.Skip("Skipping test: SUPABASE_URL and SUPABASE_KEY environment variables are required")
	}

	// Create a test file
	testContent := []byte("test content")
	testFile := "test.txt"
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Initialize storage
	storage, err := NewSupabaseStorage(url, key, "test-bucket")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test upload
	ctx := context.Background()
	uploadedURL, err := storage.UploadFile(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}
	t.Logf("File uploaded successfully: %s", uploadedURL)

	// Test download
	downloadedFile := "downloaded.txt"
	if err := storage.DownloadFile(ctx, uploadedURL, downloadedFile); err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}
	defer os.Remove(downloadedFile)

	// Verify downloaded content
	content, err := os.ReadFile(downloadedFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("Downloaded content doesn't match original content")
	}

	// Test delete
	if err := storage.DeleteFile(ctx, uploadedURL); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}
	t.Log("File deleted successfully")
}
