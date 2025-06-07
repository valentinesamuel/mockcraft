package jobs

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/valentinesamuel/mockcraft/internal/generators/base"
	"github.com/valentinesamuel/mockcraft/internal/interfaces"
	"github.com/valentinesamuel/mockcraft/internal/server/output"
	"github.com/valentinesamuel/mockcraft/internal/server/schema"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type Job struct {
	ID         string
	Status     JobStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Error      string
	OutputPath string
}

type Manager struct {
	jobs      map[string]*Job
	mu        sync.RWMutex
	outputDir string
	generator interfaces.Generator
	formatter *output.Formatter
}

func NewManager(outputDir string) (*Manager, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Manager{
		jobs:      make(map[string]*Job),
		outputDir: outputDir,
		generator: base.NewBaseGenerator(),
		formatter: output.New(outputDir),
	}, nil
}

func (m *Manager) CreateJob(file *multipart.FileHeader) (string, error) {
	jobID := uuid.New().String()

	job := &Job{
		ID:        jobID,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.mu.Lock()
	m.jobs[jobID] = job
	m.mu.Unlock()

	// Start processing in a goroutine
	go m.processJob(job, file)

	return jobID, nil
}

func (m *Manager) GetJobStatus(jobID string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}

	return job, nil
}

func (m *Manager) GetJobOutput(jobID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return "", fmt.Errorf("job not found")
	}

	if job.Status != StatusCompleted {
		return "", fmt.Errorf("job not completed")
	}

	return job.OutputPath, nil
}

func (m *Manager) processJob(job *Job, file *multipart.FileHeader) {
	// Update status to processing
	m.updateJobStatus(job, StatusProcessing, "")

	// Create a temporary directory for this job
	jobDir := filepath.Join(m.outputDir, job.ID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to create job directory: %v", err))
		return
	}

	// Save the uploaded file
	schemaPath := filepath.Join(jobDir, "schema.yaml")
	if err := saveUploadedFile(file, schemaPath); err != nil {
		m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to save schema file: %v", err))
		return
	}

	// Parse the schema
	schema, err := schema.Parse(schemaPath)
	if err != nil {
		m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to parse schema: %v", err))
		return
	}

	// Generate data for each table
	data := make(map[string][][]interface{})
	for _, table := range schema.Tables {
		rows := make([][]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row := make([]interface{}, len(table.Columns))
			for j, col := range table.Columns {
				value, err := m.generator.GenerateByType(col.Type, nil)
				if err != nil {
					m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to generate data: %v", err))
					return
				}
				row[j] = value
			}
			rows[i] = row
		}
		data[table.Name] = rows
	}

	// Format the data
	outputDir, err := m.formatter.Format(schema, data, "csv") // Default to CSV for now
	if err != nil {
		m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to format data: %v", err))
		return
	}

	// Create a zip file
	zipPath := filepath.Join(jobDir, "output.zip")
	if err := createZipFile(outputDir, zipPath); err != nil {
		m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to create zip file: %v", err))
		return
	}

	// Check zip file size
	info, err := os.Stat(zipPath)
	if err != nil {
		m.updateJobStatus(job, StatusFailed, fmt.Sprintf("failed to get zip file info: %v", err))
		return
	}

	if info.Size() > 50*1024*1024 { // 50MB
		m.updateJobStatus(job, StatusFailed, "output file size exceeds 50MB limit")
		return
	}

	// Update job with output path and status
	job.OutputPath = zipPath
	m.updateJobStatus(job, StatusCompleted, "")
}

func (m *Manager) updateJobStatus(job *Job, status JobStatus, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job.Status = status
	job.Error = errMsg
	job.UpdatedAt = time.Now()
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func createZipFile(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create a new file in the zip
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		zipFile, err := writer.Create(relPath)
		if err != nil {
			return err
		}

		// Open the source file
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		// Copy the file contents to the zip
		_, err = io.Copy(zipFile, sourceFile)
		return err
	})
}
