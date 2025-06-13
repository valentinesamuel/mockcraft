package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/valentinesamuel/mockcraft/internal/generators/all"
	"github.com/valentinesamuel/mockcraft/internal/registry"
	"github.com/valentinesamuel/mockcraft/internal/server/jobs"
)

const (
	MaxUploadSize = 30 * 1024 * 1024 // 30MB
	MaxOutputSize = 50 * 1024 * 1024 // 50MB
)

type Handler struct {
	jobManager *jobs.Manager
}

func NewHandler(jobManager *jobs.Manager) *Handler {
	return &Handler{
		jobManager: jobManager,
	}
}

// handleGenerate generates a single fake data value
func (h *Handler) HandleGenerate(c *gin.Context) {
	dataType := c.Param("type")
	generatorType := c.Param("generator")

	generator, err := registry.CreateGenerator(generatorType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Generator not found: %s", generatorType),
		})
		return
	}

	value, err := generator.GenerateByType(dataType, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":      dataType,
		"value":     value,
		"generator": generatorType,
	})
}

func (h *Handler) HandleListGenerators(c *gin.Context) {
	generators := registry.GetAvailableGenerators()

	// Create a map of generator name to its available data types
	generatorMap := make(map[string][]string)

	for _, genName := range generators {
		gen, err := registry.CreateGenerator(genName)
		if err == nil {
			generatorMap[genName] = gen.GetAvailableTypes()
		} else {
			// If generator creation fails, still include it with empty array
			generatorMap[genName] = []string{}
		}
	}

	c.JSON(http.StatusOK, generatorMap)
}

// handleSeed processes a schema file upload and starts generation
func (h *Handler) HandleSeed(c *gin.Context) {
	// Check file size
	if c.Request.ContentLength > MaxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size exceeds 30MB limit",
		})
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("schema")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	// Validate file extension
	ext := filepath.Ext(file.Filename)
	if ext != ".yaml" && ext != ".yml" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only YAML files are allowed",
		})
		return
	}

	// Get email from form data
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email is required",
		})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read uploaded file",
		})
		return
	}
	defer src.Close()

	// Create a new job
	job, err := h.jobManager.CreateJob(c.Request.Context(), src, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create job: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id": job.ID,
		"status": job.Status,
		"email":  job.Email,
	})
}

// handleJobStatus returns the status of a job
func (h *Handler) HandleJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	job, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Failed to get job status: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
		"error":      job.Error,
		"progress":   job.Progress,
		"email":      job.Email,
		"output_url": job.OutputPath,
	})
}

// handleDownload returns the generated files
func (h *Handler) HandleDownload(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	job, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Failed to get job: %v", err),
		})
		return
	}

	if job.Status != jobs.JobStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job is not completed",
		})
		return
	}

	if job.OutputPath == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No output file found",
		})
		return
	}

	// Redirect to the Supabase storage URL
	c.Redirect(http.StatusTemporaryRedirect, job.OutputPath)
}

// handleMetrics returns server metrics
func (h *Handler) HandleMetrics(c *gin.Context) {
	// TODO: Implement actual metrics
	c.JSON(http.StatusOK, gin.H{
		"active_jobs": 0,
		"total_jobs":  0,
	})
}
