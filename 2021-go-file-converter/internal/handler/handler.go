package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"file-converter/internal/converter"
	"file-converter/internal/db"
)

type ErrorResponse struct {
	Error              string              `json:"error"`
	Step               string              `json:"step,omitempty"`
	SupportedFormats   map[string][]string `json:"supported_formats,omitempty"`
}

func InitRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/formats", handleFormats)
	mux.HandleFunc("/convert", handleConvert)
	mux.HandleFunc("/jobs/", handleJob)
	mux.HandleFunc("/resources", handleResources)
	mux.HandleFunc("/resources/", handleResourceDetail)
	mux.HandleFunc("/relations", handleRelations)
}

func handleFormats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(converter.GetSupportedFormats())
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(1024 * 1024 * 1024); err != nil && err != http.ErrNotMultipart {
		sendError(w, "解析表单失败: "+err.Error(), "解析请求", http.StatusBadRequest)
		return
	}

	sourceFormat := strings.ToLower(r.FormValue("source_format"))
	targetFormat := strings.ToLower(r.FormValue("target_format"))
	resourceIDStr := r.FormValue("resource_id")

	if sourceFormat == "" || targetFormat == "" {
		sendError(w, "缺少必需参数: source_format 或 target_format", "参数验证", http.StatusBadRequest)
		return
	}

	if !converter.IsSupported(sourceFormat, targetFormat) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:            fmt.Sprintf("不支持的转换: %s -> %s", sourceFormat, targetFormat),
			Step:             "格式验证",
			SupportedFormats: converter.GetSupportedFormats(),
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		sendError(w, "获取上传文件失败: "+err.Error(), "文件读取", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var resourceID sql.NullInt64
	if resourceIDStr != "" {
		id, err := strconv.ParseInt(resourceIDStr, 10, 64)
		if err != nil {
			sendError(w, "无效的resource_id: "+err.Error(), "参数验证", http.StatusBadRequest)
			return
		}
		resourceID = sql.NullInt64{Int64: id, Valid: true}
	}

	job := &db.ConversionJob{
		ResourceID:   resourceID,
		SourceFormat: sourceFormat,
		TargetFormat: targetFormat,
		OriginalFile: header.Filename,
	}

	if err := db.CreateJob(job); err != nil {
		sendError(w, "创建转换任务失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
		return
	}

	if err := db.UpdateJobStatus(job.ID, db.StatusConverting, "开始读取文件"); err != nil {
		sendError(w, "更新任务状态失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		failedJob := &db.ConversionJob{
			ID:           job.ID,
			Status:       db.StatusFailed,
			ErrorMessage: "读取文件失败: " + err.Error(),
			CurrentStep:  "文件读取",
		}
		db.UpdateJob(failedJob)
		sendError(w, "读取文件失败: "+err.Error(), "文件读取", http.StatusInternalServerError)
		return
	}

	db.UpdateJobStatus(job.ID, db.StatusConverting, "开始格式转换")

	result, err := converter.Convert(sourceFormat, targetFormat, data)
	if err != nil {
		failedJob := &db.ConversionJob{
			ID:           job.ID,
			Status:       db.StatusFailed,
			ErrorMessage: "格式转换失败: " + err.Error(),
			CurrentStep:  "格式转换",
		}
		db.UpdateJob(failedJob)
		sendError(w, "格式转换失败: "+err.Error(), "格式转换", http.StatusInternalServerError)
		return
	}

	db.UpdateJobStatus(job.ID, db.StatusConverting, "保存转换结果")

	job.ProcessedCount = result.ProcessedCount
	job.SkippedCount = result.SkippedCount
	job.TotalCount = result.TotalCount
	job.ConvertedFile = generateFilename(header.Filename, targetFormat)
	job.Status = db.StatusCompleted
	job.CurrentStep = "转换完成"

	if len(result.Errors) > 0 {
		job.ErrorMessage = strings.Join(result.Errors, "\n")
	}

	if err := db.UpdateJob(job); err != nil {
		sendError(w, "更新任务信息失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", getContentType(targetFormat))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", job.ConvertedFile))
	w.Header().Set("X-Job-ID", fmt.Sprintf("%d", job.ID))
	w.Header().Set("X-Processed-Count", strconv.Itoa(result.ProcessedCount))
	w.Header().Set("X-Skipped-Count", strconv.Itoa(result.SkippedCount))
	w.Header().Set("X-Total-Count", strconv.Itoa(result.TotalCount))

	if len(result.Errors) > 0 {
		w.Header().Set("X-Warnings", strconv.Itoa(len(result.Errors)))
	}

	w.Write(result.Data)
}

func handleJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/jobs/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendError(w, "无效的任务ID", "参数验证", http.StatusBadRequest)
		return
	}

	job, err := db.GetJob(id)
	if err == sql.ErrNoRows {
		http.Error(w, "任务不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		sendError(w, "获取任务信息失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func handleResources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resources, err := db.ListResources()
		if err != nil {
			sendError(w, "获取资源列表失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resources)

	case http.MethodPost:
		var res db.Resource
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			sendError(w, "解析请求体失败: "+err.Error(), "参数验证", http.StatusBadRequest)
			return
		}

		if res.Name == "" || res.Type == "" {
			sendError(w, "缺少必需字段: name 或 type", "参数验证", http.StatusBadRequest)
			return
		}

		if err := db.CreateResource(&res); err != nil {
			sendError(w, "创建资源失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)

	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func handleResourceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/resources/")
	
	if strings.Contains(idStr, "/jobs") {
		parts := strings.Split(idStr, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			sendError(w, "无效的资源ID", "参数验证", http.StatusBadRequest)
			return
		}

		jobs, err := db.ListJobsByResource(id)
		if err != nil {
			sendError(w, "获取资源任务列表失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
			return
		}

		var totalProcessed, totalSkipped, totalTotal int
		statusSummary := make(map[string]int)
		for _, job := range jobs {
			totalProcessed += job.ProcessedCount
			totalSkipped += job.SkippedCount
			totalTotal += job.TotalCount
			statusSummary[string(job.Status)]++
		}

		relations, _ := db.GetResourceRelations(id)

		summary := map[string]interface{}{
			"resource_id":     id,
			"total_jobs":      len(jobs),
			"total_processed": totalProcessed,
			"total_skipped":   totalSkipped,
			"total_records":   totalTotal,
			"status_summary":  statusSummary,
			"jobs":            jobs,
			"relations":       relations,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendError(w, "无效的资源ID", "参数验证", http.StatusBadRequest)
		return
	}

	resource, err := db.GetResource(id)
	if err == sql.ErrNoRows {
		http.Error(w, "资源不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		sendError(w, "获取资源信息失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

func handleRelations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var rel db.ResourceRelation
		if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
			sendError(w, "解析请求体失败: "+err.Error(), "参数验证", http.StatusBadRequest)
			return
		}

		if rel.SourceID == 0 || rel.TargetID == 0 || rel.RelationType == "" {
			sendError(w, "缺少必需字段: source_id, target_id 或 relation_type", "参数验证", http.StatusBadRequest)
			return
		}

		if err := db.CreateRelation(&rel); err != nil {
			sendError(w, "创建关联关系失败: "+err.Error(), "数据库操作", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rel)

	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func sendError(w http.ResponseWriter, message, step string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Step:  step,
	})
}

func generateFilename(original, ext string) string {
	name := strings.TrimSuffix(filepath.Base(original), filepath.Ext(original))
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s.%s", name, timestamp, ext)
}

func getContentType(format string) string {
	switch format {
	case "json":
		return "application/json"
	case "yaml", "yml":
		return "application/yaml"
	case "csv":
		return "text/csv"
	case "html":
		return "text/html"
	default:
		return "application/octet-stream"
	}
}
