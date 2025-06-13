package jobs

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/mockcraft/internal/generators/registry"
	"github.com/valentinesamuel/mockcraft/internal/server/output"
	"github.com/valentinesamuel/mockcraft/internal/server/schema"
	"github.com/valentinesamuel/mockcraft/internal/server/storage"
)

// Processor handles the processing of data generation jobs
type Processor struct {
	schemaStorage storage.Storage
	outputStorage storage.Storage
	formatter     *output.Formatter
	outputDir     string
	manager       *Manager
	rdb           *redis.Client
	registry      *registry.IndustryRegistry
	ctx           context.Context
}

// NewProcessor creates a new processor
func NewProcessor(
	schemaStorage storage.Storage,
	outputStorage storage.Storage,
	formatter *output.Formatter,
	outputDir string,
	manager *Manager,
	redisOpt *redis.Options,
) (*Processor, error) {
	rdb := redis.NewClient(redisOpt)

	registry := registry.NewIndustryRegistry()

	return &Processor{
		schemaStorage: schemaStorage,
		outputStorage: outputStorage,
		formatter:     formatter,
		outputDir:     outputDir,
		manager:       manager,
		rdb:           rdb,
		registry:      registry,
		ctx:           context.Background(),
	}, nil
}

// Start begins processing jobs
func (p *Processor) Start() error {
	log.Println("Starting job processor...")

	// Start processing jobs in a loop
	for {
		select {
		case <-p.ctx.Done():
			log.Println("Job processor stopped")
			return nil
		default:
			// Process any pending jobs
			if err := p.processPendingJobs(); err != nil {
				log.Printf("Error processing jobs: %v", err)
			}
			// Sleep briefly to avoid tight loop
			time.Sleep(time.Second)
		}
	}
}

// processPendingJobs processes any pending jobs in the queue
func (p *Processor) processPendingJobs() error {
	// Get pending jobs from the manager
	jobs, err := p.manager.GetPendingJobs(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	for _, job := range jobs {
		log.Printf("Processing job %s", job.ID)

		// Update job status to processing
		if err := p.manager.UpdateJobStatus(p.ctx, job.ID, JobStatusProcessing); err != nil {
			log.Printf("Failed to update job status: %v", err)
			continue
		}

		// Process the job based on its type
		switch job.Type {
		case JobTypeGenerateData:
			if err := p.processGenerateData(p.ctx, job); err != nil {
				log.Printf("Failed to process generate data job: %v", err)
				if err := p.manager.UpdateJobStatus(p.ctx, job.ID, JobStatusFailed); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				continue
			}
		default:
			log.Printf("Unknown job type: %s", job.Type)
			continue
		}

		// Update job status to completed
		if err := p.manager.UpdateJobStatus(p.ctx, job.ID, JobStatusCompleted); err != nil {
			log.Printf("Failed to update job status: %v", err)
		}
	}

	return nil
}

func (p *Processor) processGenerateData(ctx context.Context, job *Job) error {
	// Download schema file
	tempFile, err := os.CreateTemp("", "schema-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if err := p.schemaStorage.DownloadFile(ctx, job.SchemaPath, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to download schema: %w", err)
	}

	// Load schema
	schema, err := schema.LoadSchema(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Validate schema
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Create output directory
	outputDir := filepath.Join(os.TempDir(), fmt.Sprintf("mockcraft_%s", job.ID))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	defer os.RemoveAll(outputDir)

	// Generate data for each column
	data := make(map[string][]interface{})
	for _, col := range schema.Columns {
		// Get generator for the industry and type
		generator, err := p.registry.GetGenerator(col.Industry, col.Type)
		if err != nil {
			return fmt.Errorf("failed to get generator for column %s: %w", col.Name, err)
		}

		// Generate data
		value, err := generator(col.Constraints)
		if err != nil {
			return fmt.Errorf("failed to generate data for column %s: %w", col.Name, err)
		}

		data[col.Name] = []interface{}{value}
	}

	// Create output file
	outputPath := filepath.Join(outputDir, "output.json")
	outputData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output data: %w", err)
	}

	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	// Create zip file
	zipPath := filepath.Join(outputDir, "output.zip")
	if err := p.createZipFile(outputPath, zipPath); err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	// Upload zip file
	outputURL, err := p.outputStorage.UploadFile(ctx, zipPath)
	if err != nil {
		return fmt.Errorf("failed to upload output file: %w", err)
	}

	// Update job status
	job.Status = "completed"
	job.OutputURL = outputURL
	job.UpdatedAt = time.Now()

	if err := p.manager.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (p *Processor) createZipFile(sourcePath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	writer, err := zipWriter.Create(filepath.Base(sourcePath))
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	if _, err := io.Copy(writer, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file to zip: %w", err)
	}

	return nil
}

// sendSuccessEmail sends a success email with the download link
func (p *Processor) sendSuccessEmail(email, downloadURL string) error {
	// TODO: Implement email sending
	// For now, just log the success
	fmt.Printf("Success email would be sent to %s with download link: %s\n", email, downloadURL)
	return nil
}

// sendFailureEmail sends a failure email with the error message
func (p *Processor) sendFailureEmail(email, errorMsg string) error {
	// TODO: Implement email sending
	// For now, just log the failure
	fmt.Printf("Failure email would be sent to %s with error: %s\n", email, errorMsg)
	return nil
}

func (p *Processor) getJobMetadata(jobID string) (*Job, error) {
	// Get job metadata from Redis
	jobData, err := p.rdb.Get(context.Background(), fmt.Sprintf("job:%s", jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal(jobData, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

func (p *Processor) saveJob(job *Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Set different TTLs based on job status
	var ttl time.Duration
	switch job.Status {
	case JobStatusCompleted:
		// Keep completed jobs for 1 hour
		ttl = 1 * time.Hour
	case JobStatusFailed:
		// Keep failed jobs for 1 hour
		ttl = 1 * time.Hour
	default:
		// Keep pending/processing jobs for 24 hours
		ttl = 24 * time.Hour
	}

	// Save job metadata to Redis
	return p.rdb.Set(context.Background(), fmt.Sprintf("job:%s", job.ID), jobData, ttl).Err()
}
