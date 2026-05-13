package services

import (
	"encoding/json"
	"fmt"
	"image-batch/config"
	"image-batch/models"
	"image-batch/repositories"
	"image-batch/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type BatchService struct {
	imageService *ImageService
	zipService   *ZipService
	taskRepo     *repositories.TaskRepository
}

func NewBatchService(imageService *ImageService, zipService *ZipService, taskRepo *repositories.TaskRepository) *BatchService {
	return &BatchService{
		imageService: imageService,
		zipService:   zipService,
		taskRepo:     taskRepo,
	}
}

func (s *BatchService) ProcessBatch(files []*multipart.FileHeader, operations []models.ImageOperation) (*models.ProcessTask, error) {
	totalFiles := len(files)
	task, err := s.taskRepo.CreateTask(totalFiles)
	if err != nil {
		return nil, err
	}

	go s.processAsync(task, files, operations)

	return task, nil
}

func (s *BatchService) processAsync(task *models.ProcessTask, files []*multipart.FileHeader, operations []models.ImageOperation) {
	tempDir := filepath.Join(config.TempDir, fmt.Sprintf("task_%d_%d", task.ID, time.Now().Unix()))
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	os.MkdirAll(inputDir, 0755)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(tempDir)

	result := &models.ProcessResult{
		Total:      len(files),
		Success:    0,
		Failed:     0,
		Successful: make([]string, 0),
		Skipped:    make([]models.SkippedFile, 0),
	}

	var processedFiles []string

	for _, fh := range files {
		inputPath, err := s.saveUploadedFile(fh, inputDir)
		if err != nil {
			result.Skipped = append(result.Skipped, models.SkippedFile{
				Filename: fh.Filename,
				Reason:   "Failed to save file: " + err.Error(),
			})
			result.Failed++
			continue
		}

		if !utils.IsSupportedFormat(fh.Filename) {
			result.Skipped = append(result.Skipped, models.SkippedFile{
				Filename: fh.Filename,
				Reason:   "Unsupported format. Supported: " + utils.GetSupportedFormats(),
			})
			result.Failed++
			os.Remove(inputPath)
			continue
		}

		file, err := os.Open(inputPath)
		if err != nil {
			result.Skipped = append(result.Skipped, models.SkippedFile{
				Filename: fh.Filename,
				Reason:   "Failed to open file: " + err.Error(),
			})
			result.Failed++
			continue
		}

		if err := utils.ValidateImageFile(file); err != nil {
			file.Close()
			result.Skipped = append(result.Skipped, models.SkippedFile{
				Filename: fh.Filename,
				Reason:   err.Error(),
			})
			result.Failed++
			os.Remove(inputPath)
			continue
		}
		file.Close()

		outputPath, err := s.imageService.ProcessImage(inputPath, operations, outputDir)
		if err != nil {
			result.Skipped = append(result.Skipped, models.SkippedFile{
				Filename: fh.Filename,
				Reason:   "Processing failed: " + err.Error(),
			})
			result.Failed++
		} else {
			result.Successful = append(result.Successful, filepath.Base(outputPath))
			processedFiles = append(processedFiles, outputPath)
			result.Success++
		}

		os.Remove(inputPath)
	}

	zipPath := filepath.Join(config.OutputDir, fmt.Sprintf("batch_%d.zip", task.ID))
	os.MkdirAll(config.OutputDir, 0755)

	if len(processedFiles) > 0 {
		err := s.zipService.CreateZip(processedFiles, result, zipPath)
		if err != nil {
			task.Status = "failed"
		} else {
			task.Status = "completed"
		}
	} else {
		task.Status = "completed"
		zipPath = ""
	}

	reportJSON, _ := json.Marshal(result)

	task.Success = result.Success
	task.Failed = result.Failed
	task.FinishedAt = time.Now()
	task.ZipPath = zipPath
	task.Report = string(reportJSON)

	s.taskRepo.UpdateTask(task)
}

func (s *BatchService) saveUploadedFile(fh *multipart.FileHeader, dir string) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filePath := filepath.Join(dir, fh.Filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *BatchService) GetTask(id int64) (*models.ProcessTask, error) {
	return s.taskRepo.GetTaskByID(id)
}
