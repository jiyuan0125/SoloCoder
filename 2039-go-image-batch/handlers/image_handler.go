package handlers

import (
	"encoding/json"
	"fmt"
	"image-batch/config"
	"image-batch/models"
	"image-batch/services"
	"image-batch/utils"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	batchService *services.BatchService
	imageService *services.ImageService
}

func NewImageHandler(batchService *services.BatchService, imageService *services.ImageService) *ImageHandler {
	return &ImageHandler{
		batchService: batchService,
		imageService: imageService,
	}
}

func (h *ImageHandler) ProcessBatch(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No images uploaded"})
		return
	}

	operationsJSON := c.PostForm("operations")
	if operationsJSON == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No operations specified"})
		return
	}

	var request models.ProcessRequest
	if err := json.Unmarshal([]byte(operationsJSON), &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operations JSON: " + err.Error()})
		return
	}

	if err := h.imageService.ValidateOperations(request.Operations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hasUnsupportedFormat := false
	for _, fh := range files {
		if !utils.IsSupportedFormat(fh.Filename) {
			hasUnsupportedFormat = true
			break
		}
	}

	if hasUnsupportedFormat {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":              "One or more files have unsupported format",
			"supported_formats":  utils.GetSupportedFormats(),
			"format_list":        utils.GetSupportedFormatList(),
		})
		return
	}

	task, err := h.batchService.ProcessBatch(files, request.Operations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start processing: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Processing started",
		"task_id": task.ID,
		"status":  task.Status,
		"check_status": fmt.Sprintf("/api/tasks/%d", task.ID),
	})
}

func (h *ImageHandler) GetTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.batchService.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	response := gin.H{
		"id":         task.ID,
		"status":     task.Status,
		"total":      task.Total,
		"success":    task.Success,
		"failed":     task.Failed,
		"created_at": task.CreatedAt,
	}

	if task.Status == "completed" || task.Status == "failed" {
		response["finished_at"] = task.FinishedAt
		if task.Report != "" {
			var report models.ProcessResult
			json.Unmarshal([]byte(task.Report), &report)
			response["report"] = report
		}
		if task.ZipPath != "" {
			response["download_url"] = fmt.Sprintf("/api/tasks/%d/download", task.ID)
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *ImageHandler) DownloadZip(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.batchService.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task not completed yet. Status: " + task.Status})
		return
	}

	if task.ZipPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No processed files to download"})
		return
	}

	if _, err := os.Stat(task.ZipPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Zip file no longer exists"})
		return
	}

	filename := fmt.Sprintf("processed_images_%d.zip", task.ID)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")

	c.File(task.ZipPath)
}

func (h *ImageHandler) GetSupportedFormats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"formats": utils.GetSupportedFormatList(),
		"formats_string": utils.GetSupportedFormats(),
		"extensions": []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
		"max_upload_size": config.MaxUploadSize,
		"max_memory_per_image": config.MaxMemoryPerImage,
	})
}

func (h *ImageHandler) GetOperations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"operations": []gin.H{
			{
				"type":        "resize",
				"description": "Resize image to specified dimensions",
				"parameters":  []string{"width", "height"},
				"example":     `{"type":"resize","width":800,"height":600}`,
			},
			{
				"type":        "crop",
				"description": "Crop image from specified coordinates",
				"parameters":  []string{"x", "y", "width", "height"},
				"example":     `{"type":"crop","x":100,"y":100,"width":400,"height":400}`,
			},
			{
				"type":        "convert",
				"description": "Convert image to different format",
				"parameters":  []string{"format", "quality (optional)"},
				"supported_formats": utils.GetSupportedFormatList(),
				"example":     `{"type":"convert","format":"webp","quality":80}`,
			},
			{
				"type":        "watermark",
				"description": "Add text watermark to image",
				"parameters":  []string{"watermark", "opacity"},
				"example":     `{"type":"watermark","watermark":"Copyright 2024","opacity":0.5}`,
			},
		},
		"request_format": `{
  "operations": [
    {"type":"resize","width":800,"height":600},
    {"type":"convert","format":"webp"}
  ]
}`,
		"notes": []string{
			"Multiple operations can be chained in sequence",
			"Operations are applied in the order specified",
			"GIF animations are preserved during resize and crop operations",
			"EXIF information is preserved during format conversion",
			strings.Replace(config.SupportedFormats, ", ", ", ", -1),
		},
	})
}

func (h *ImageHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "image-batch",
		"port":    config.ServerPort,
	})
}
