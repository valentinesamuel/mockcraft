package jobs

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
	_ "github.com/valentinesamuel/mockcraft/internal/generators/all"
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

	return &Processor{
		schemaStorage: schemaStorage,
		outputStorage: outputStorage,
		formatter:     formatter,
		outputDir:     outputDir,
		manager:       manager,
		rdb:           rdb,
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
	// Download schema file - always use YAML format
	tempFile, err := os.CreateTemp("", "schema-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if err := p.schemaStorage.DownloadFile(ctx, job.SchemaURL, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to download schema: %w", err)
	}

	// Load schema
	loadedSchema, err := schema.LoadSchema(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Validate schema
	if err := loadedSchema.Validate(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Create output directory
	outputDir := filepath.Join(os.TempDir(), fmt.Sprintf("mockcraft_%s", job.ID))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	defer os.RemoveAll(outputDir)

	// Store generated IDs for each table
	tableIDs := make(map[string][]string)

	// Generate data for each table
	data := make(map[string][][]interface{})
	for _, table := range loadedSchema.Tables {
		// Convert schema.Table to types.Table
		dbTable := &types.Table{
			Name:    table.Name,
			Count:   table.Count,
			Columns: make([]types.Column, len(table.Columns)),
		}

		// Convert columns
		for i, col := range table.Columns {
			dbTable.Columns[i] = types.Column{
				Name:       col.Name,
				Type:       col.Type,
				IsPrimary:  col.IsPrimary,
				IsNullable: col.IsNullable,
				Generator:  col.Generator,
				Industry:   col.Industry,
				Params:     col.Params,
			}
		}

		// Convert relations
		dbRelations := make([]types.Relationship, len(loadedSchema.Relations))
		for i, rel := range loadedSchema.Relations {
			dbRelations[i] = types.Relationship{
				Type:       rel.Type,
				FromTable:  rel.FromTable,
				FromColumn: rel.FromColumn,
				ToTable:    rel.ToTable,
				ToColumn:   rel.ToColumn,
			}
		}

		// Generate data for the table
		tableData := make([][]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			// Generate row using the same function as the seed module
			row, err := generators.GenerateRow(dbTable, tableIDs, dbRelations, i)
			if err != nil {
				return fmt.Errorf("failed to generate data for table %s: %w", table.Name, err)
			}

			// Convert row to array of values in column order
			values := make([]interface{}, len(table.Columns))
			for j, col := range table.Columns {
				values[j] = row[col.Name]
			}
			tableData[i] = values

			// Store primary key for foreign key references
			for _, col := range table.Columns {
				if col.IsPrimary {
					if id, ok := row[col.Name].(string); ok {
						tableIDs[table.Name] = append(tableIDs[table.Name], id)
					}
					break
				}
			}
		}
		data[table.Name] = tableData
	}

	// Create output files based on the specified format
	outputFormat := job.OutputFormat
	if outputFormat == "" {
		outputFormat = "json" // Default to JSON if not specified
	}

	// Create a zip file to store all output files
	// Use email and timestamp as part of the filename to prevent clashes
	emailPrefix := strings.ReplaceAll(job.Email, "@", "_at_")
	emailPrefix = strings.ReplaceAll(emailPrefix, ".", "_dot_")
	timestamp := time.Now().Format("20060102_150405")
	zipPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_output.zip", emailPrefix, timestamp))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Generate output files for each table
	for tableName, tableData := range data {
		// Find the current table's schema
		var currentTable *schema.Table
		for _, t := range loadedSchema.Tables {
			if t.Name == tableName {
				currentTable = &t
				break
			}
		}
		if currentTable == nil {
			return fmt.Errorf("table %s not found in schema", tableName)
		}

		// Convert table data to the appropriate format
		var outputData []byte
		var err error

		switch outputFormat {
		case "json":
			outputData, err = json.MarshalIndent(tableData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON data for table %s: %w", tableName, err)
			}
		case "csv":
			// Create CSV writer
			var buf bytes.Buffer
			writer := csv.NewWriter(&buf)

			// Write headers
			if len(tableData) > 0 {
				headers := make([]string, len(currentTable.Columns))
				for i, col := range currentTable.Columns {
					headers[i] = col.Name
				}
				if err := writer.Write(headers); err != nil {
					return fmt.Errorf("failed to write CSV headers for table %s: %w", tableName, err)
				}
			}

			// Write data
			for _, row := range tableData {
				record := make([]string, len(row))
				for i, val := range row {
					record[i] = fmt.Sprintf("%v", val)
				}
				if err := writer.Write(record); err != nil {
					return fmt.Errorf("failed to write CSV record for table %s: %w", tableName, err)
				}
			}
			writer.Flush()
			outputData = buf.Bytes()

		case "sql":
			// Create SQL file content
			var buf bytes.Buffer

			// Write CREATE TABLE statement
			fmt.Fprintf(&buf, "CREATE TABLE IF NOT EXISTS `%s` (\n", tableName)
			for i, col := range currentTable.Columns {
				fmt.Fprintf(&buf, "  `%s` TEXT", col.Name)
				if col.IsPrimary {
					fmt.Fprintf(&buf, " PRIMARY KEY")
				}
				if !col.IsNullable {
					fmt.Fprintf(&buf, " NOT NULL")
				}
				if i < len(currentTable.Columns)-1 {
					fmt.Fprintf(&buf, ",")
				}
				fmt.Fprintf(&buf, "\n")
			}
			fmt.Fprintf(&buf, ");\n\n")

			// Write INSERT statements
			for _, row := range tableData {
				values := make([]string, len(row))
				for i, val := range row {
					values[i] = fmt.Sprintf("'%v'", val)
				}
				fmt.Fprintf(&buf, "INSERT INTO `%s` VALUES (%s);\n",
					tableName,
					strings.Join(values, ", "),
				)
			}
			outputData = buf.Bytes()

		default:
			return fmt.Errorf("unsupported output format: %s", outputFormat)
		}

		// Create file in zip
		writer, err := zipWriter.Create(fmt.Sprintf("%s.%s", tableName, outputFormat))
		if err != nil {
			return fmt.Errorf("failed to create zip entry for table %s: %w", tableName, err)
		}

		// Write data to zip
		if _, err := writer.Write(outputData); err != nil {
			return fmt.Errorf("failed to write data to zip for table %s: %w", tableName, err)
		}
	}

	// Explicitly close the zip writer before uploading
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	// Upload zip file with the email-prefixed name
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
