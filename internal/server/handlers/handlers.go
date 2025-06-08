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

	// Create a new job
	jobID, err := h.jobManager.CreateJob(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create job",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id": jobID,
		"status": "processing",
	})
}

// handleJobStatus returns the status of a job
func (h *Handler) HandleJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	status, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// handleDownload returns the generated files
func (h *Handler) HandleDownload(c *gin.Context) {
	jobID := c.Param("id")
	filePath, err := h.jobManager.GetJobOutput(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job output not found",
		})
		return
	}

	c.File(filePath)
}

// handleMetrics returns server metrics
func (h *Handler) HandleMetrics(c *gin.Context) {
	// TODO: Implement actual metrics
	c.JSON(http.StatusOK, gin.H{
		"active_jobs": 0,
		"total_jobs":  0,
	})
}
