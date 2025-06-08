package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/mockcraft/internal/server/output"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type Job struct {
	ID         string                 `json:"id"`
	Status     JobStatus              `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Error      string                 `json:"error,omitempty"`
	OutputPath string                 `json:"output_path,omitempty"`
	SchemaPath string                 `json:"schema_path,omitempty"`
	TaskID     string                 `json:"task_id,omitempty"`
	Progress   int                    `json:"progress"`
	TotalSteps int                    `json:"total_steps"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Manager struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	rdb       *redis.Client
	outputDir string
	formatter *output.Formatter
}

func NewManager(redisOpt asynq.RedisClientOpt, outputDir string) (*Manager, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Password: redisOpt.Password,
		DB:       redisOpt.DB,
	})

	return &Manager{
		client:    client,
		inspector: inspector,
		rdb:       rdb,
		outputDir: outputDir,
		formatter: output.New(outputDir),
	}, nil
}

func (m *Manager) CreateJob(file *multipart.FileHeader) (string, error) {
	jobID := uuid.New().String()
	now := time.Now()

	// Create job directory
	jobDir := filepath.Join(m.outputDir, jobID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create job directory: %w", err)
	}

	// Save the uploaded file
	schemaPath := filepath.Join(jobDir, "schema.yaml")
	if err := saveUploadedFile(file, schemaPath); err != nil {
		return "", fmt.Errorf("failed to save schema file: %w", err)
	}

	// Create task payload
	payload := GenerateDataPayload{
		JobID:      jobID,
		SchemaPath: schemaPath,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create task
	task := asynq.NewTask(TypeGenerateData, payloadBytes, asynq.MaxRetry(4))

	// Enqueue task
	info, err := m.client.Enqueue(task)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Create job record
	job := &Job{
		ID:         jobID,
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		SchemaPath: schemaPath,
		TaskID:     info.ID,
		Progress:   0,
		TotalSteps: 0,
	}

	// Save job metadata to Redis
	if err := m.saveJob(job); err != nil {
		return "", fmt.Errorf("failed to save job: %w", err)
	}

	return jobID, nil
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
		job.Status = StatusPending
	case asynq.TaskStateActive:
		job.Status = StatusProcessing
	case asynq.TaskStateCompleted:
		job.Status = StatusCompleted
	case asynq.TaskStateRetry, asynq.TaskStateArchived:
		job.Status = StatusFailed
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

	if job.Status != StatusCompleted {
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

		if job.Status == StatusCompleted || job.Status == StatusFailed {
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
