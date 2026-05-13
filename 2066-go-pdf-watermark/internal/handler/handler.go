package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"pdf-watermark/internal/cache"
	"pdf-watermark/internal/pdf"
	"pdf-watermark/internal/store"
	"pdf-watermark/internal/types"
)

const (
	maxFileSize = 50 * 1024 * 1024 // 50MB
)

type Handler struct {
	store   *store.Store
	cache   *cache.Cache
	tmpl    *template.Template
	workDir string
	mu      sync.Mutex
}

func New() (*Handler, error) {
	workDir := "./data"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(workDir, "output"), 0755); err != nil {
		return nil, err
	}

	s, err := store.New(workDir)
	if err != nil {
		return nil, err
	}

	c, err := cache.New(workDir)
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(template.New("home").Parse(htmlTemplate))

	return &Handler{
		store:   s,
		cache:   c,
		tmpl:    tmpl,
		workDir: workDir,
	}, nil
}

func (h *Handler) Close() error {
	return h.store.Close()
}

func (h *Handler) ServeHome(w http.ResponseWriter, _ *http.Request) {
	jobs, _ := h.store.ListJobs()
	h.tmpl.Execute(w, map[string]interface{}{
		"Jobs": jobs,
	})
}

func (h *Handler) ListJobs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobs, _ := h.store.ListJobs()
	json.NewEncoder(w).Encode(jobs)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	id := parts[2]
	job, err := h.store.GetJob(id)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	id := parts[2]
	job, err := h.store.GetJob(id)
	if err != nil || job.Status != "completed" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	outputPath := filepath.Join(h.workDir, "output", id+".pdf")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_watermarked.pdf", job.Filename))
	http.ServeFile(w, r, outputPath)
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1024)

	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			http.Error(w, "文件超过50MB", http.StatusRequestEntityTooLarge)
		}
		return
	}

	pdfFile, fh, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "请上传PDF文件", http.StatusBadRequest)
		return
	}
	defer pdfFile.Close()

	fhName := strings.ToLower(fh.Filename)
	if !strings.HasSuffix(fhName, ".pdf") {
		http.Error(w, "只支持PDF文件", http.StatusBadRequest)
		return
	}

	config, err := parseConfig(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if fh.Size > maxFileSize {
		http.Error(w, "文件超过50MB", http.StatusRequestEntityTooLarge)
		return
	}

	job := &types.Job{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Filename:  fh.Filename,
		Filesize:  fh.Size,
		Status:   "processing",
		CreatedAt: time.Now(),
	}
	h.store.SaveJob(job)

	tempPDF := filepath.Join(h.workDir, job.ID+".pdf")
	tempPDFOut, err := os.Create(tempPDF)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		h.store.SaveJob(job)
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}
	io.Copy(tempPDFOut, pdfFile)
	tempPDFOut.Close()

	pdfData, err := pdf.NewPDF(tempPDF)
	if err != nil {
		os.Remove(tempPDF)
		job.Status = "failed"
		job.Error = err.Error()
		h.store.SaveJob(job)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if pdfData.IsEncrypted() {
		os.Remove(tempPDF)
		job.Status = "failed"
		job.Error = "PDF文件已加密"
		h.store.SaveJob(job)
		http.Error(w, "PDF文件已加密", http.StatusBadRequest)
		return
	}

	pageCount := pdfData.PageCount()
	job.TotalPages = pageCount

	if config.Mode == "image" {
		imgFile, imgFh, err := r.FormFile("image")
		if err != nil {
			os.Remove(tempPDF)
			job.Status = "failed"
			job.Error = "请上传水印图片"
			h.store.SaveJob(job)
			http.Error(w, "请上传水印图片", http.StatusBadRequest)
			return
		}
		defer imgFile.Close()

		imgName := strings.ToLower(imgFh.Filename)
		if !strings.HasSuffix(imgName, ".png") && !strings.HasSuffix(imgName, ".jpg") && !strings.HasSuffix(imgName, ".jpeg") {
			os.Remove(tempPDF)
			job.Status = "failed"
			job.Error = "只支持PNG/JPEG图片"
			h.store.SaveJob(job)
			http.Error(w, "只支持PNG/JPEG图片", http.StatusBadRequest)
			return
		}

		tempImg := filepath.Join(h.workDir, job.ID+".img")
		tempImgOut, _ := os.Create(tempImg)
		io.Copy(tempImgOut, imgFile)
		tempImgOut.Close()
		defer os.Remove(tempImg)

		config.ImagePath = tempImg
	}

	outputPath := filepath.Join(h.workDir, "output", job.ID+".pdf")

	if err := pdfData.AddWatermark(outputPath, config); err != nil {
		os.Remove(tempPDF)
		job.Status = "failed"
		job.Error = err.Error()
		h.store.SaveJob(job)
		if strings.Contains(err.Error(), "页数") {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	os.Remove(tempPDF)

	outData, _ := pdf.NewPDF(outputPath)
	if outData.PageCount() != pageCount {
		os.Remove(outputPath)
		job.Status = "failed"
		job.Error = "处理后页数不一致"
		h.store.SaveJob(job)
		http.Error(w, "处理失败", http.StatusInternalServerError)
		return
	}

	job.Status = "completed"
	job.OutputPath = outputPath
	h.store.SaveJob(job)

	h.notifyCacheUpdate(job.Filesize, config.Mode, job.Filesize)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_watermarked.pdf", job.Filename))
	http.ServeFile(w, r, outputPath)
}

func (h *Handler) notifyCacheUpdate(fileSize int64, mode string, totalBytes int64) {
	go func() {
		sizeCategory := "small"
		switch {
		case totalBytes > 10*1024*1024:
			sizeCategory = "large"
		case totalBytes > 1*1024*1024:
			sizeCategory = "medium"
		}

		h.cache.UpdateUsage(totalBytes, mode, sizeCategory)

		jobs, _ := h.store.ListJobs()
		total := int64(len(jobs))
		if total > 0 {
			h.cache.AllocateQuotas(total)
		}
	}()
}

func parseConfig(r *http.Request) (*types.WatermarkConfig, error) {
	mode := r.FormValue("mode")
	if mode == "" {
		mode = "text"
	}

	config := &types.WatermarkConfig{
		Mode: mode,
		Text: r.FormValue("text"),
	}

	if mode == "text" {
		if config.Text == "" {
			config.Text = "WATERMARK"
		}

		if v := r.FormValue("font_size"); v != "" {
			if s, err := strconv.ParseFloat(v, 64); err == nil {
				config.FontSize = s
			}
		}
		if config.FontSize <= 0 {
			config.FontSize = 24
		}

		if v := r.FormValue("color"); v != "" {
			config.Color = v
		}
		if config.Color == "" {
			config.Color = "0,0,0"
		}

		if v := r.FormValue("opacity"); v != "" {
			if s, err := strconv.ParseFloat(v, 64); err == nil {
				config.Opacity = s
			}
		}
		if config.Opacity <= 0 || config.Opacity > 1 {
			config.Opacity = 0.5
		}

		if v := r.FormValue("rotation"); v != "" {
			if s, err := strconv.ParseFloat(v, 64); err == nil {
				config.Rotation = s
			}
		}
	}

	if v := r.FormValue("position"); v != "" {
		config.Position = v
	}
	if config.Position == "" {
		config.Position = "center"
	}

	return config, nil
}

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>PDF水印系统</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 900px; margin: 40px auto; padding: 20px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; }
        .form-group { margin: 15px 0; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .row { display: grid; grid-template-columns: 1fr 1fr; gap: 15px; }
        button { background: #007bff; color: white; border: none; padding: 12px 30px; border-radius: 4px; cursor: pointer; font-size: 16px; }
        button:hover { background: #0056b3; }
        .jobs { margin-top: 40px; }
        .job { border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 4px; }
        .status-completed { color: #28a745; }
        .status-failed { color: #dc3545; }
        .hidden { display: none; }
    </style>
</head>
<body>
<div class="container">
    <h1>PDF水印系统</h1>
    <form action="/upload" method="POST" enctype="multipart/form-data">
        <div class="form-group">
            <label>PDF文件:</label>
            <input type="file" name="pdf" accept=".pdf" required>
        </div>
        <div class="row">
            <div class="form-group">
                <label>水印模式:</label>
                <select name="mode" id="mode" onchange="toggleMode()">
                    <option value="text">文字水印</option>
                    <option value="image">图片水印</option>
                </select>
            </div>
            <div class="form-group">
                <label>位置:</label>
                <select name="position">
                    <option value="center">居中</option>
                    <option value="top-left">左上</option>
                    <option value="top-right">右上</option>
                    <option value="bottom-left">左下</option>
                    <option value="bottom-right">右下</option>
                    <option value="tile">平铺</option>
                </select>
            </div>
        </div>
        <div id="text-options">
            <div class="form-group">
                <label>水印文字:</label>
                <input type="text" name="text" placeholder="WATERMARK">
            </div>
            <div class="row">
                <div class="form-group">
                    <label>字体大小:</label>
                    <input type="number" name="font_size" value="24">
                </div>
                <div class="form-group">
                    <label>颜色(R,G,B):</label>
                    <input type="text" name="color" placeholder="0,0,0" value="0,0,0">
                </div>
            </div>
            <div class="row">
                <div class="form-group">
                    <label>透明度(0-1):</label>
                    <input type="number" step="0.1" name="opacity" value="0.5">
                </div>
                <div class="form-group">
                    <label>旋转角度:</label>
                    <input type="number" name="rotation" value="0">
                </div>
            </div>
        </div>
        <div id="image-options" class="hidden">
            <div class="form-group">
                <label>水印图片(PNG/JPEG):</label>
                <input type="file" name="image" accept=".png,.jpg,.jpeg">
            </div>
        </div>
        <button type="submit">添加水印</button>
    </form>
</div>
<div class="container jobs">
    <h2>最近任务</h2>
    {{range .Jobs}}
    <div class="job">
        <strong>{{.Filename}}</strong> - 
        <span class="status-{{.Status}}">{{.Status}}</span>
        {{if .Error}}<br><small style="color:#dc3545">{{.Error}}</small>{{end}}
        {{if eq .Status "completed"}}
        <br><a href="/download/{{.ID}}">下载</a>
        {{end}}
    </div>
    {{else}}
    <p>暂无任务</p>
    {{end}}
</div>
<script>
function toggleMode() {
    var mode = document.getElementById('mode').value;
    document.getElementById('text-options').style.display = mode === 'text' ? 'block' : 'none';
    document.getElementById('image-options').style.display = mode === 'image' ? 'block' : 'none';
}
</script>
</body>
</html>`
