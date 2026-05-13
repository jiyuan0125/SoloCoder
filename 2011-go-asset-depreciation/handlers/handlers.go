package handlers

import (
	"asset-depreciation/database"
	"asset-depreciation/models"
	"asset-depreciation/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIError{
		Error:   http.StatusText(status),
		Message: message,
	})
}

func parseID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "assets" && i+1 < len(parts) {
			return strconv.ParseInt(parts[i+1], 10, 64)
		}
	}
	return 0, errors.New("invalid path")
}

func parseHistoryID(path string) (int64, int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var assetID int64
	for i, part := range parts {
		if part == "assets" && i+1 < len(parts) {
			id, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				return 0, 0, err
			}
			assetID = id
		}
		if part == "history" && i+1 < len(parts) {
			historyID, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				return 0, 0, err
			}
			return assetID, historyID, nil
		}
	}
	return 0, 0, errors.New("invalid path")
}

func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listAssets(w, r)
	case http.MethodPost:
		createAsset(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	assetID, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	writeJSON(w, http.StatusOK, asset)
}

func AssetStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	assetID, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	if asset.Status == models.StatusScrapped || asset.Status == models.StatusCompleted {
		writeError(w, http.StatusBadRequest, "asset is already completed or scrapped")
		return
	}

	updated, err := service.TransitionAsset(assetID, req.Action, req.Remark)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func AssetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	assetID, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	histories, err := database.GetOperationHistory(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, histories)
}

func HistoryRemarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	assetID, historyID, err := parseHistoryID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	history, err := database.GetOperationHistoryByID(historyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if history == nil || history.AssetID != assetID {
		writeError(w, http.StatusNotFound, "history record not found")
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := database.AddRemarkToHistory(historyID, req.Remark); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := database.GetOperationHistoryByID(historyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func AssetDepreciationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	assetID, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	records, err := database.GetDepreciationRecords(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, records)
}

func AssetScrapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	assetID, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	if asset.Status == models.StatusScrapped {
		writeError(w, http.StatusBadRequest, "asset is already scrapped")
		return
	}

	var req models.ScrapAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ScrappedValue < 0 {
		writeError(w, http.StatusBadRequest, "scrapped_value must be >= 0")
		return
	}

	updated, err := database.ScrapAsset(assetID, req.ScrappedValue, req.Remark)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func DepartmentSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	summaries, err := database.GetDepartmentSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if summaries == nil {
		summaries = []*models.DepartmentSummary{}
	}

	writeJSON(w, http.StatusOK, summaries)
}

func DepreciationTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := service.ProcessMonthlyDepreciation(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "depreciation process completed",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func listAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := database.ListAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if assets == nil {
		assets = []*models.Asset{}
	}

	writeJSON(w, http.StatusOK, assets)
}

func createAsset(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SalvageRate == 0 {
		req.SalvageRate = 0.05
	}

	if err := service.ValidateCreateAssetRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := time.Parse("2006-01-02", req.PurchaseDate); err != nil {
		writeError(w, http.StatusBadRequest, "invalid purchase_date format, expected YYYY-MM-DD")
		return
	}

	asset, err := database.CreateAsset(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/assets/%d", asset.ID))
	writeJSON(w, http.StatusCreated, asset)
}
