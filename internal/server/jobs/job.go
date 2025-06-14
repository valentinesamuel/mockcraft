package jobs

import "time"

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeGenerateData JobType = "generate_data"
)

// Job represents a data generation job
type Job struct {
	// Core fields
	ID     string    `json:"id"`
	Type   JobType   `json:"type"`
	Status JobStatus `json:"status"`
	Error  string    `json:"error,omitempty"`

	// Schema and output information
	SchemaPath   string `json:"schema_path,omitempty"`
	SchemaURL    string `json:"schema_url,omitempty"`
	OutputPath   string `json:"output_path,omitempty"`
	OutputURL    string `json:"output_url,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`

	// Progress tracking
	Progress   int                    `json:"progress"`
	TotalSteps int                    `json:"total_steps"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`

	// Contact information
	Email string `json:"email,omitempty"`

	// Timestamps
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}
