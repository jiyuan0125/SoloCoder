package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"excel-report-system/database"
	"excel-report-system/excelgen"
	"excel-report-system/template"
	"excel-report-system/workflow"
)

type Handler struct {
	db        *sql.DB
	tmplDir   string
	exportDir string
}

func NewHandler(db *sql.DB, tmplDir, exportDir string) *Handler {
	return &Handler{
		db:        db,
		tmplDir:   tmplDir,
		exportDir: exportDir,
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parseID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid path")
	}
	last := parts[len(parts)-1]
	return strconv.ParseInt(last, 10, 64)
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Name       string `json:"name"`
			TemplateID int64  `json:"template_id"`
			Data       string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" || req.TemplateID == 0 {
			respondError(w, http.StatusBadRequest, "name and template_id are required")
			return
		}

		id, err := database.CreateReport(h.db, req.Name, req.TemplateID, req.Data)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusCreated, map[string]int64{"id": id})
		return
	}

	reports, err := database.ListReports(h.db)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, reports)
}

func (h *Handler) ReportDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		report, err := database.GetReport(h.db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusNotFound, "report not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, report)

	case http.MethodPut:
		var req struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := database.UpdateReportData(h.db, id, req.Data); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	case http.MethodDelete:
		_, err := database.GetReport(h.db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusNotFound, "report not found")
				return
			}
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) WorkflowAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ReportID     int64  `json:"report_id"`
		Action       string `json:"action"`
		RejectReason string `json:"reject_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	report, err := database.GetReport(h.db, req.ReportID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "report not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if workflow.IsClosed(report.Status) {
		respondError(w, http.StatusBadRequest, "closed report cannot be modified")
		return
	}

	if req.Action == workflow.ActionReject && strings.TrimSpace(req.RejectReason) == "" {
		respondError(w, http.StatusBadRequest, "reject reason is required")
		return
	}

	nextStatus, err := workflow.GetNextStatus(report.Status, req.Action)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var reason string
	if req.Action == workflow.ActionReject {
		reason = req.RejectReason
	}

	if err := database.UpdateReportStatus(h.db, req.ReportID, nextStatus, reason); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "success",
		"new_status":  nextStatus,
		"status_name": workflow.StatusDisplayName(nextStatus),
	})
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.Method == http.MethodPost {
		var tmpl template.TemplateDef
		if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := template.SaveTemplate(h.tmplDir, &tmpl); err != nil {
			if pe, ok := err.(*template.ParseError); ok {
				respondJSON(w, http.StatusBadRequest, map[string]interface{}{
					"error":       "invalid template",
					"field":       pe.Field,
					"field_error": pe.Message,
				})
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		id, err := database.CreateTemplate(h.db, tmpl.Name, filepath.Join(h.tmplDir, tmpl.Name+".json"), tmpl.Description)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, map[string]int64{"id": id})
		return
	}

	templates, err := database.ListTemplates(h.db)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, templates)
}

func (h *Handler) TemplateDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if path == "" {
		respondError(w, http.StatusBadRequest, "template name or id required")
		return
	}

	if r.Method == http.MethodGet {
		if strings.HasSuffix(path, ".json") {
			name := strings.TrimSuffix(path, ".json")
			tmpl, err := template.LoadTemplate(h.tmplDir, name)
			if err != nil {
				respondError(w, http.StatusNotFound, "template not found")
				return
			}
			respondJSON(w, http.StatusOK, tmpl)
			return
		}

		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid id")
			return
		}
		tmpl, err := database.GetTemplate(h.db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusNotFound, "template not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, tmpl)
		return
	}

	if r.Method == http.MethodDelete {
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid id")
			return
		}
		tmpl, err := database.GetTemplate(h.db, id)
		if err == nil {
			template.DeleteTemplate(h.tmplDir, tmpl.Name)
			database.DeleteTemplate(h.db, id)
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		return
	}

	respondError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			respondError(w, http.StatusBadRequest, "name is required")
			return
		}

		id, err := database.CreateEntity(h.db, req.Name, req.Description)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusCreated, map[string]int64{"id": id})
		return
	}

	entities, err := database.ListEntities(h.db)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, entities)
}

func (h *Handler) EntityDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/entities/")
	if strings.Contains(path, "/details") {
		entityIDStr := strings.Split(path, "/")[0]
		entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entity id")
			return
		}

		if r.Method == http.MethodGet {
			details, err := database.ListEntityDetails(h.db, entityID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, details)
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				Content    string `json:"content"`
				ChangeType string `json:"change_type"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			if req.Content == "" || req.ChangeType == "" {
				respondError(w, http.StatusBadRequest, "content and change_type are required")
				return
			}

			id, err := database.AddEntityDetail(h.db, entityID, req.Content, req.ChangeType)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusCreated, map[string]int64{"id": id})
			return
		}
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		entity, err := database.GetEntity(h.db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusNotFound, "entity not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, entity)

	case http.MethodDelete:
		if err := database.DeleteEntity(h.db, id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) ListEntityDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	entityIDStr := r.URL.Query().Get("entity_id")
	if entityIDStr == "" {
		respondError(w, http.StatusBadRequest, "entity_id is required")
		return
	}

	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	details, err := database.ListEntityDetails(h.db, entityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, details)
}

func (h *Handler) EntityDetailOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := parseID(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := database.DeleteEntityDetail(h.db, id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) ExportReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		TemplateName string          `json:"template_name"`
		DataSource   json.RawMessage `json:"data_source"`
		DBQuery      string          `json:"db_query,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TemplateName == "" {
		respondError(w, http.StatusBadRequest, "template_name is required")
		return
	}

	tmpl, err := template.LoadTemplate(h.tmplDir, req.TemplateName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		if pe, ok := err.(*template.ParseError); ok {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "template format error",
				"field":       pe.Field,
				"field_error": pe.Message,
			})
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var ds *excelgen.DataSource
	if req.DBQuery != "" {
		ds, err = h.queryDataSource(req.DBQuery, len(tmpl.Sheets))
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		ds, err = excelgen.ParseDataSource(string(req.DataSource))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid data source")
			return
		}
	}

	fileName := fmt.Sprintf("report_%s.xlsx", time.Now().Format("20060102_150405"))
	outputPath := filepath.Join(h.exportDir, fileName)

	chartErrors, err := excelgen.GenerateExcel(tmpl, ds, outputPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	file, err := os.Open(outputPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	if len(chartErrors) > 0 {
		errJSON, _ := json.Marshal(chartErrors)
		w.Header().Set("X-Chart-Errors", string(errJSON))
	}

	io.Copy(w, file)
}

func (h *Handler) queryDataSource(query string, sheetCount int) (*excelgen.DataSource, error) {
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns error: %w", err)
	}

	var tableData []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			switch v := values[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		tableData = append(tableData, row)
	}

	ds := &excelgen.DataSource{
		StaticData: make(map[string]interface{}),
		TableData:  make(map[string][]map[string]interface{}),
	}

	if len(tableData) > 0 {
		for i := 0; i < sheetCount; i++ {
			ds.TableData[fmt.Sprintf("table_%d", i)] = tableData
		}
	}

	return ds, nil
}
