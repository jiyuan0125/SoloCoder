package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"sqliteparser/pkg/shared"
	"sqliteparser/pkg/sqlite"
)

var (
	currentDB *sqlite.Database
	dbMutex   sync.RWMutex
)

func main() {
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/tables", handleTables)
	http.HandleFunc("/api/query", handleQuery)
	http.HandleFunc("/api/header", handleHeader)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, shared.ErrorResponse{Error: message})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to get file: "+err.Error())
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file: "+err.Error())
		return
	}

	db, err := sqlite.OpenBytes(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse database: "+err.Error())
		return
	}

	dbMutex.Lock()
	currentDB = db
	dbMutex.Unlock()

	writeJSON(w, http.StatusOK, shared.UploadResponse{
		Success: true,
		Message: "database uploaded successfully",
	})
}

func handleTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dbMutex.RLock()
	db := currentDB
	dbMutex.RUnlock()

	if db == nil {
		writeError(w, http.StatusBadRequest, "no database uploaded")
		return
	}

	tables := db.GetTables()
	result := make([]shared.TableSchema, 0, len(tables))
	for _, t := range tables {
		result = append(result, shared.TableSchema{
			Name:      t.Name,
			CreateSQL: t.SQL,
			RootPage:  t.RootPage,
		})
	}

	writeJSON(w, http.StatusOK, shared.TablesResponse{Tables: result})
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dbMutex.RLock()
	db := currentDB
	dbMutex.RUnlock()

	if db == nil {
		writeError(w, http.StatusBadRequest, "no database uploaded")
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeError(w, http.StatusBadRequest, "table parameter required")
		return
	}

	limit := 0
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	var rows []map[string]interface{}
	var err error

	if limit > 0 {
		rows, err = db.QueryTableWithLimit(tableName, limit, offset)
	} else {
		rows, err = db.QueryTable(tableName)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}

	if rows == nil {
		writeError(w, http.StatusNotFound, "table not found")
		return
	}

	processedRows := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		processedRow := make(map[string]interface{})
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				processedRow[k] = map[string]interface{}{
					"$type": "blob",
					"data":  b,
				}
			} else {
				processedRow[k] = v
			}
		}
		processedRows = append(processedRows, processedRow)
	}

	writeJSON(w, http.StatusOK, shared.QueryResponse{
		Table: tableName,
		Rows:  processedRows,
		Count: len(processedRows),
	})
}

func handleHeader(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dbMutex.RLock()
	db := currentDB
	dbMutex.RUnlock()

	if db == nil {
		writeError(w, http.StatusBadRequest, "no database uploaded")
		return
	}

	h := db.Header()
	writeJSON(w, http.StatusOK, shared.HeaderResponse{
		Header: &shared.HeaderInfo{
			MagicString:        h.MagicString,
			PageSize:           h.PageSize,
			FileFormatWrite:    h.FileFormatWrite,
			FileFormatRead:     h.FileFormatRead,
			ReservedSpace:      h.ReservedSpace,
			MaxEmbeddedPayload: h.MaxEmbeddedPayload,
			MinEmbeddedPayload: h.MinEmbeddedPayload,
			LeafPayloadFraction: h.LeafPayloadFraction,
			FileChangeCounter:  h.FileChangeCounter,
			PageCount:          h.PageCount,
			FirstFreelistPage:  h.FirstFreelistPage,
			FreelistPageCount:  h.FreelistPageCount,
			SchemaCookie:       h.SchemaCookie,
			SchemaFormat:       h.SchemaFormat,
			DefaultPageCache:   h.DefaultPageCache,
			AutoVacuumTop:      h.AutoVacuumTop,
			IncrementalVacuum:  h.IncrementalVacuum,
			TextEncoding:       h.TextEncoding,
			UserVersion:        h.UserVersion,
			ApplicationID:      h.ApplicationID,
			VersionValidFor:    h.VersionValidFor,
			SQLiteVersion:      h.SQLiteVersion,
		},
	})
}
