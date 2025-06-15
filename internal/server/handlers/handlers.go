package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/valentinesamuel/mockcraft/internal/generators"
	"github.com/valentinesamuel/mockcraft/internal/server/jobs"
)

const (
	MaxUploadSize = 30 * 1024 * 1024 // 30MB
	MaxOutputSize = 50 * 1024 * 1024 // 50MB
)

type Handler struct {
	jobManager *jobs.Manager
	engine     *generators.Engine
}

func NewHandler(jobManager *jobs.Manager) *Handler {
	return &Handler{
		jobManager: jobManager,
		engine:     generators.GetGlobalEngine(),
	}
}

// handleGenerate generates a single fake data value
func (h *Handler) HandleGenerate(c *gin.Context) {
	industry := c.Param("industry")
	generator := c.Param("generator")

	// Validate generator exists
	if err := h.engine.ValidateGenerator(industry, generator); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Generator not found: %s for industry %s", generator, industry),
		})
		return
	}

	// Get parameters from query string
	params := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	value, err := h.engine.Generate(industry, generator, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"industry":  industry,
		"generator": generator,
		"value":     value,
		"params":    params,
	})
}

func (h *Handler) HandleListGenerators(c *gin.Context) {
	// Check if detailed parameter info is requested
	showParams := c.Query("params") == "true"
	
	if showParams {
		// Get all generator information with parameters
		generatorInfo := h.engine.GetAllGeneratorInfo()
		c.JSON(http.StatusOK, generatorInfo)
	} else {
		// Get simple generator list
		generatorMap := h.engine.GetAllGenerators()
		c.JSON(http.StatusOK, generatorMap)
	}
}

// HandleGeneratorInfo returns detailed information about a specific generator
func (h *Handler) HandleGeneratorInfo(c *gin.Context) {
	industry := c.Param("industry")
	generator := c.Param("generator")

	info, err := h.engine.GetGeneratorInfo(industry, generator)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Generator not found: %s for industry %s", generator, industry),
		})
		return
	}

	c.JSON(http.StatusOK, info)
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

	// Get output format from form data
	outputFormat := c.PostForm("output_format")
	if outputFormat != "" && outputFormat != "json" && outputFormat != "csv" && outputFormat != "sql" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Output format must be one of: json, csv, sql",
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
	job, err := h.jobManager.CreateJob(c.Request.Context(), src, email, outputFormat)
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
