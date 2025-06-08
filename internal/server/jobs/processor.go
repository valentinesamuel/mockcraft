package jobs

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/mockcraft/internal/registry"
	"github.com/valentinesamuel/mockcraft/internal/server/formatter"
	"github.com/valentinesamuel/mockcraft/internal/server/schema"
	"github.com/valentinesamuel/mockcraft/internal/server/storage"
)

const (
	TypeGenerateData = "generate:data"
)

// Processor handles job processing
type Processor struct {
	client        *asynq.Client
	inspector     *asynq.Inspector
	schemaStorage storage.Storage
	outputStorage storage.Storage
	formatter     formatter.Formatter
	outputDir     string
	manager       *Manager
	rdb           *redis.Client
}

// NewProcessor creates a new Processor instance
func NewProcessor(redisOpt asynq.RedisClientOpt, outputDir, supabaseURL, supabaseKey string) (*Processor, error) {
	// Create storage instances for different buckets
	schemaStorage, err := storage.NewSupabaseStorage(supabaseURL, supabaseKey, "schemas")
	if err != nil {
		return nil, fmt.Errorf("failed to create schema storage: %w", err)
	}

	outputStorage, err := storage.NewSupabaseStorage(supabaseURL, supabaseKey, "output")
	if err != nil {
		return nil, fmt.Errorf("failed to create output storage: %w", err)
	}

	formatter := formatter.NewCSVFormatter()

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Password: redisOpt.Password,
		DB:       redisOpt.DB,
	})

	return &Processor{
		client:        asynq.NewClient(redisOpt),
		inspector:     asynq.NewInspector(redisOpt),
		schemaStorage: schemaStorage,
		outputStorage: outputStorage,
		formatter:     formatter,
		outputDir:     outputDir,
		rdb:           rdb,
	}, nil
}

// SetManager sets the job manager for the processor
func (p *Processor) SetManager(manager *Manager) {
	p.manager = manager
}

func (p *Processor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	switch t.Type() {
	case TypeGenerateData:
		var payload GenerateDataPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		return p.processGenerateData(ctx, &payload)
	default:
		return fmt.Errorf("unknown task type: %s", t.Type())
	}
}

func (p *Processor) processGenerateData(ctx context.Context, payload *GenerateDataPayload) error {
	// Create a temporary file for the schema
	tempFile, err := os.CreateTemp("", "schema-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download the schema from schema storage
	if err := p.schemaStorage.DownloadFile(ctx, payload.SchemaURL, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to download schema: %w", err)
	}

	// Parse the schema
	schema, err := schema.Parse(tempFile.Name())
	if err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Create output directory
	outputDir := filepath.Join(p.outputDir, payload.JobID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate data for each table
	for _, table := range schema.Tables {
		// Generate data
		data := make([][]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row := make([]interface{}, len(table.Columns))
			for j, col := range table.Columns {
				// Get generator for the column
				generator, err := registry.CreateGenerator(col.Generator)
				if err != nil {
					// Send failure email
					if err := p.sendFailureEmail(payload.Email, fmt.Sprintf("No generator found for column %s: %s", col.Name, col.Generator)); err != nil {
						return fmt.Errorf("failed to send failure email: %w", err)
					}
					return fmt.Errorf("no generator found for column %s: %s", col.Name, col.Generator)
				}

				value, err := generator.GenerateByType(col.Type, col.Params)
				if err != nil {
					// Send failure email
					if err := p.sendFailureEmail(payload.Email, fmt.Sprintf("Failed to generate data for column %s: %v", col.Name, err)); err != nil {
						return fmt.Errorf("failed to send failure email: %w", err)
					}
					return fmt.Errorf("failed to generate data for column %s: %w", col.Name, err)
				}
				row[j] = value
			}
			data[i] = row
		}

		// Format and save data
		outputPath, err := p.formatter.Format(schema, map[string][][]interface{}{table.Name: data}, "csv")
		if err != nil {
			// Send failure email
			if err := p.sendFailureEmail(payload.Email, fmt.Sprintf("Failed to format data for table %s: %v", table.Name, err)); err != nil {
				return fmt.Errorf("failed to send failure email: %w", err)
			}
			return fmt.Errorf("failed to format data for table %s: %w", table.Name, err)
		}

		// Move the formatted file to the output directory
		destPath := filepath.Join(outputDir, filepath.Base(outputPath))
		if err := os.Rename(outputPath, destPath); err != nil {
			// Send failure email
			if err := p.sendFailureEmail(payload.Email, fmt.Sprintf("Failed to move output file for table %s: %v", table.Name, err)); err != nil {
				return fmt.Errorf("failed to send failure email: %w", err)
			}
			return fmt.Errorf("failed to move output file for table %s: %w", table.Name, err)
		}
	}

	// Create a zip file of all output files
	zipPath := filepath.Join(outputDir, "output.zip")
	if err := p.createZipFile(outputDir, zipPath); err != nil {
		// Send failure email
		if err := p.sendFailureEmail(payload.Email, "Failed to create zip file of generated data."); err != nil {
			return fmt.Errorf("failed to send failure email: %w", err)
		}
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	// Upload the zip file to output storage
	outputURL, err := p.outputStorage.UploadFile(ctx, zipPath)
	if err != nil {
		// Send failure email
		if err := p.sendFailureEmail(payload.Email, "Failed to upload generated data."); err != nil {
			return fmt.Errorf("failed to send failure email: %w", err)
		}
		return fmt.Errorf("failed to upload output file: %w", err)
	}

	// Update job status
	job, err := p.manager.GetJob(ctx, payload.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	job.Status = StatusCompleted
	job.OutputURL = outputURL
	job.UpdatedAt = time.Now()

	if err := p.manager.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Send success email
	if err := p.sendSuccessEmail(payload.Email, outputURL); err != nil {
		return fmt.Errorf("failed to send success email: %w", err)
	}

	return nil
}

// createZipFile creates a zip file containing all files in the given directory
func (p *Processor) createZipFile(dir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the zip file itself
		if info.IsDir() || path == zipPath {
			return nil
		}

		// Create a new file in the zip
		writer, err := zipWriter.Create(info.Name())
		if err != nil {
			return fmt.Errorf("failed to create file in zip: %w", err)
		}

		// Open the file to be added to the zip
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// Copy the file contents to the zip
		if _, err := io.Copy(writer, file); err != nil {
			return fmt.Errorf("failed to copy file to zip: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
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
	case StatusCompleted:
		// Keep completed jobs for 1 hour
		ttl = 1 * time.Hour
	case StatusFailed:
		// Keep failed jobs for 1 hour
		ttl = 1 * time.Hour
	default:
		// Keep pending/processing jobs for 24 hours
		ttl = 24 * time.Hour
	}

	// Save job metadata to Redis
	return p.rdb.Set(context.Background(), fmt.Sprintf("job:%s", job.ID), jobData, ttl).Err()
}

func (p *Processor) GetTaskInfo(taskID string) (*asynq.TaskInfo, error) {
	return p.inspector.GetTaskInfo("default", taskID)
}

// Start starts processing tasks
func (p *Processor) Start() error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     p.rdb.Options().Addr,
			Password: p.rdb.Options().Password,
			DB:       p.rdb.Options().DB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeGenerateData, p.ProcessTask)

	return srv.Run(mux)
}
