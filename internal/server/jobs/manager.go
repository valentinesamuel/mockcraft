package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/mockcraft/internal/server/output"
	"github.com/valentinesamuel/mockcraft/internal/server/schema"
	"github.com/valentinesamuel/mockcraft/internal/server/storage"
)

// GenerateDataPayload represents the payload for data generation tasks
type GenerateDataPayload struct {
	JobID        string `json:"job_id"`
	SchemaURL    string `json:"schema_url"`
	Email        string `json:"email"`
	OutputFormat string `json:"output_format"`
}

type Manager struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	rdb       *redis.Client
	outputDir string
	formatter *output.Formatter
	storage   *storage.SupabaseStorage
}

// generateID generates a unique ID for jobs
func generateID() string {
	return uuid.New().String()
}

func NewManager(redisOpt asynq.RedisClientOpt, outputDir string, supabaseURL, supabaseKey string) (*Manager, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Password: redisOpt.Password,
		DB:       redisOpt.DB,
	})

	// Initialize Supabase storage
	supabaseStorage, err := storage.NewSupabaseStorage(supabaseURL, supabaseKey, "schemas")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase storage: %w", err)
	}

	return &Manager{
		client:    asynq.NewClient(redisOpt),
		inspector: asynq.NewInspector(redisOpt),
		rdb:       rdb,
		outputDir: outputDir,
		formatter: output.New(outputDir),
		storage:   supabaseStorage,
	}, nil
}

func (m *Manager) CreateJob(ctx context.Context, schemaFile io.Reader, email string, outputFormat string) (*Job, error) {
	// Create a temporary file to store the uploaded schema
	tempFile, err := os.CreateTemp("", "schema-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Copy the uploaded file to the temporary file
	if _, err := io.Copy(tempFile, schemaFile); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}
	tempFile.Close()

	// Validate the schema
	if _, err := schema.LoadSchema(tempFile.Name()); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	// Create a unique filename for the schema using email and timestamp
	emailPrefix := strings.ReplaceAll(email, "@", "_at_")
	emailPrefix = strings.ReplaceAll(emailPrefix, ".", "_dot_")
	timestamp := time.Now().Format("20060102_150405")
	schemaFileName := fmt.Sprintf("%s_%s_schema.yaml", emailPrefix, timestamp)

	// Upload the schema file to Supabase storage with the new filename
	storageURL, err := m.storage.UploadFile(ctx, tempFile.Name(), schemaFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to upload schema to storage: %w", err)
	}

	// Create a new job
	job := &Job{
		ID:           generateID(),
		Type:         JobTypeGenerateData,
		Status:       JobStatusPending,
		Email:        email,
		SchemaURL:    storageURL,
		OutputFormat: outputFormat,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save job metadata to Redis
	if err := m.saveJob(job); err != nil {
		return nil, fmt.Errorf("failed to save job metadata: %w", err)
	}

	// Create task payload
	payload := &GenerateDataPayload{
		JobID:        job.ID,
		SchemaURL:    storageURL,
		Email:        email,
		OutputFormat: outputFormat,
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Enqueue the task
	task := asynq.NewTask(string(JobTypeGenerateData), payloadBytes)
	if _, err := m.client.Enqueue(task); err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return job, nil
}

func (m *Manager) GetJobStatus(jobID string) (*Job, error) {
	// Get job metadata from Redis
	jobData, err := m.inspector.GetTaskInfo("default", jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job info: %w", err)
	}

	// Get job metadata
	job, err := m.getJobMetadata(jobID)
	if err != nil {
		return nil, err
	}

	// Update job status based on task status
	switch jobData.State {
	case asynq.TaskStatePending:
		job.Status = JobStatusPending
	case asynq.TaskStateActive:
		job.Status = JobStatusProcessing
	case asynq.TaskStateCompleted:
		job.Status = JobStatusCompleted
	case asynq.TaskStateRetry, asynq.TaskStateArchived:
		job.Status = JobStatusFailed
		job.Error = jobData.LastErr
	}

	job.UpdatedAt = time.Now()
	if err := m.saveJob(job); err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	return job, nil
}

func (m *Manager) GetJobOutput(jobID string) (string, error) {
	job, err := m.getJobMetadata(jobID)
	if err != nil {
		return "", err
	}

	if job.Status != JobStatusCompleted {
		return "", fmt.Errorf("job not completed")
	}

	return job.OutputPath, nil
}

func (m *Manager) saveJob(job *Job) error {
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
	return m.rdb.Set(context.Background(), fmt.Sprintf("job:%s", job.ID), jobData, ttl).Err()
}

// CleanupCompletedJobs removes all completed and failed jobs from Redis
func (m *Manager) CleanupCompletedJobs() error {
	// Get all job keys
	keys, err := m.rdb.Keys(context.Background(), "job:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get job keys: %w", err)
	}

	// Check each job's status and delete if completed or failed
	for _, key := range keys {
		jobData, err := m.rdb.Get(context.Background(), key).Bytes()
		if err != nil {
			continue // Skip if job not found
		}

		var job Job
		if err := json.Unmarshal(jobData, &job); err != nil {
			continue // Skip if job data is invalid
		}

		if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
			if err := m.rdb.Del(context.Background(), key).Err(); err != nil {
				return fmt.Errorf("failed to delete job %s: %w", job.ID, err)
			}
		}
	}

	return nil
}

func (m *Manager) getJobMetadata(jobID string) (*Job, error) {
	// Get job metadata from Redis
	jobData, err := m.rdb.Get(context.Background(), fmt.Sprintf("job:%s", jobID)).Bytes()
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

// Helper function to save uploaded file
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

// GetJob retrieves a job by its ID
func (m *Manager) GetJob(ctx context.Context, jobID string) (*Job, error) {
	// Get job metadata from Redis
	jobData, err := m.rdb.Get(ctx, fmt.Sprintf("job:%s", jobID)).Bytes()
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

// UpdateJob updates a job's metadata
func (m *Manager) UpdateJob(ctx context.Context, job *Job) error {
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
	return m.rdb.Set(ctx, fmt.Sprintf("job:%s", job.ID), jobData, ttl).Err()
}

// GetPendingJobs retrieves all pending jobs from the queue
func (m *Manager) GetPendingJobs(ctx context.Context) ([]*Job, error) {
	// Get all jobs from Redis
	keys, err := m.rdb.Keys(ctx, "job:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job keys: %w", err)
	}

	var jobs []*Job
	for _, key := range keys {
		jobData, err := m.rdb.Get(ctx, key).Result()
		if err != nil {
			log.Printf("Failed to get job data for key %s: %v", key, err)
			continue
		}

		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			log.Printf("Failed to unmarshal job data: %v", err)
			continue
		}

		// Only include pending jobs
		if job.Status == JobStatusPending {
			jobs = append(jobs, &job)
		}
	}

	return jobs, nil
}

// UpdateJobStatus updates the status of a job
func (m *Manager) UpdateJobStatus(ctx context.Context, jobID string, status JobStatus) error {
	key := fmt.Sprintf("job:%s", jobID)

	// Get current job data
	jobData, err := m.rdb.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	// Update status and completed time if needed
	job.Status = status
	if status == JobStatusCompleted || status == JobStatusFailed {
		job.CompletedAt = time.Now()
	}

	// Save updated job
	updatedData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	if err := m.rdb.Set(ctx, key, updatedData, 0).Err(); err != nil {
		return fmt.Errorf("failed to save job data: %w", err)
	}

	return nil
}
