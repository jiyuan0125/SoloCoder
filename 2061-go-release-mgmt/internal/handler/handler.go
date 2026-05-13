package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"release-mgmt/internal/approval"
	"release-mgmt/internal/change"
	"release-mgmt/internal/database"
	"release-mgmt/internal/deployment"
	"release-mgmt/internal/model"
	"release-mgmt/internal/version"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, model.ErrorResponse{Error: err.Error()})
}

func getReleaseIDFromPath(path string, prefix string) (int64, error) {
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, errors.New("missing release_id")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func ListReleases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	releases, err := version.GetAllReleases()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, releases)
}

func CreateRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req version.CreateReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	release, err := version.CreateRelease(&req)
	if err != nil {
		if err.Error() == "invalid semver format, expected MAJOR.MINOR.PATCH" {
			writeError(w, http.StatusBadRequest, err)
		} else if err.Error() == "duplicate version" {
			writeError(w, http.StatusConflict, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, release)
}

func GetRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	prefix := "/api/releases/"
	versionStr := strings.TrimPrefix(r.URL.Path, prefix)
	if versionStr == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing version"))
		return
	}

	release, err := version.GetRelease(versionStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if release == nil {
		writeError(w, http.StatusNotFound, errors.New("release not found"))
		return
	}
	writeJSON(w, http.StatusOK, release)
}

func SubmitForReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var body struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := version.SubmitForReview(id, body.Operator); err != nil {
		if err.Error() == "release not found" {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted for review"})
}

func AddChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req change.AddChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.ReleaseID = id

	item, err := change.AddChange(&req)
	if err != nil {
		if err.Error() == "release not found" {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func GetChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	items, err := change.GetChanges(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func Approve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req approval.ApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.ReleaseID = id

	err = approval.Approve(&req)
	if err != nil {
		switch err.Error() {
		case "release not found":
			writeError(w, http.StatusNotFound, err)
		case "cannot approve your own release", "already submitted approval for this release":
			writeError(w, http.StatusForbidden, err)
		case "rejection reason is required":
			writeError(w, http.StatusBadRequest, err)
		case "release is not in pending review status":
			writeError(w, http.StatusBadRequest, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "approval recorded"})
}

func GetApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	approvals, err := approval.GetApprovals(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, approvals)
}

func DeployToStaging(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req deployment.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.ReleaseID = id
	req.Environment = "staging"

	deploy, err := deployment.DeployToStaging(&req)
	if err != nil {
		if err.Error() == "release not found" {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, deploy)
}

func DeployToProduction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req deployment.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.ReleaseID = id
	req.Environment = "production"

	deploy, err := deployment.DeployToProduction(&req)
	if err != nil {
		if err.Error() == "release not found" {
			writeError(w, http.StatusNotFound, err)
		} else if err.Error() == "staging smoke test not passed, cannot deploy to production" {
			writeError(w, http.StatusBadRequest, err)
		} else {
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, deploy)
}

func RollbackProduction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var body struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	deploy, err := deployment.RollbackProduction(id, body.Operator)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, deploy)
}

func GetOperationHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	id, err := getReleaseIDFromPath(r.URL.Path, "/api/releases/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	release, err := version.GetReleaseByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if release == nil {
		writeError(w, http.StatusNotFound, errors.New("release not found"))
		return
	}

	history, err := database.GetReleaseHistory(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func AddNoteToHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/releases/"), "/")
	if len(parts) < 3 {
		writeError(w, http.StatusBadRequest, errors.New("missing history_id"))
		return
	}

	historyID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var body struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := database.UpdateOperationNotes(historyID, body.Notes); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "note added"})
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}
