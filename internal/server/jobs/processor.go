package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"
	"github.com/valentinesamuel/mockcraft/internal/registry"
	"github.com/valentinesamuel/mockcraft/internal/server/output"
	"github.com/valentinesamuel/mockcraft/internal/server/schema"
)

const (
	TypeGenerateData = "generate:data"
)

type GenerateDataPayload struct {
	JobID      string `json:"job_id"`
	SchemaPath string `json:"schema_path"`
}

type Processor struct {
	outputDir string
	formatter *output.Formatter
	inspector *asynq.Inspector
}

func NewProcessor(redisOpt asynq.RedisClientOpt, outputDir string) (*Processor, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Processor{
		outputDir: outputDir,
		formatter: output.New(outputDir),
		inspector: asynq.NewInspector(redisOpt),
	}, nil
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
	// Parse the schema
	schema, err := schema.Parse(payload.SchemaPath)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Get the appropriate generator for the industry
	generator, err := registry.CreateGenerator(schema.Industry)
	if err != nil {
		// Fallback to base generator if industry-specific generator not found
		generator, err = registry.GetDefaultGenerator()
		if err != nil {
			return fmt.Errorf("failed to get generator: %w", err)
		}
	}

	// Generate data for each table
	data := make(map[string][][]interface{})
	for _, table := range schema.Tables {
		rows := make([][]interface{}, table.Count)
		for j := 0; j < table.Count; j++ {
			row := make([]interface{}, len(table.Columns))
			for k, col := range table.Columns {
				value, err := generator.GenerateByType(col.Type, nil)
				if err != nil {
					return fmt.Errorf("failed to generate data: %w", err)
				}
				row[k] = value
			}
			rows[j] = row
		}
		data[table.Name] = rows
	}

	// Format the data
	outputDir, err := p.formatter.Format(schema, data, "csv") // Default to CSV for now
	if err != nil {
		return fmt.Errorf("failed to format data: %w", err)
	}

	// Create a zip file
	zipPath := filepath.Join(filepath.Dir(payload.SchemaPath), "output.zip")
	if err := createZipFile(outputDir, zipPath); err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	// Check zip file size
	info, err := os.Stat(zipPath)
	if err != nil {
		return fmt.Errorf("failed to get zip file info: %w", err)
	}

	if info.Size() > 50*1024*1024 { // 50MB
		return fmt.Errorf("output file size exceeds 50MB limit")
	}

	return nil
}

func (p *Processor) GetTaskInfo(taskID string) (*asynq.TaskInfo, error) {
	return p.inspector.GetTaskInfo("default", taskID)
}

// Helper function to create a zip file
func createZipFile(sourceDir, zipPath string) error {
	// ... existing createZipFile implementation ...
	return nil
}
