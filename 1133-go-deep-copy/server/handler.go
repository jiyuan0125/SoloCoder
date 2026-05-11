package main

import (
	"encoding/json"
	"io"
	"net/http"

	"deepcopy/deepcopy"
	"deepcopy/shared"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, &shared.ErrorResponse{Error: err.Error()})
}

func handleRegisterPrototype(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var req shared.RegisterPrototypeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, &errNameRequired)
		return
	}

	deepcopy.Register(req.Name, req.Data)

	resp := &shared.RegisterPrototypeResponse{
		Success: true,
		Name:    req.Name,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleListPrototypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := &shared.ListPrototypesResponse{
		Prototypes: deepcopy.ListPrototypes(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var req shared.CloneRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, &errNameRequired)
		return
	}

	cloned, err := deepcopy.Clone(req.Name)
	if err != nil {
		resp := &shared.CloneResponse{
			Success: false,
			Error:   err.Error(),
		}
		writeJSON(w, http.StatusNotFound, resp)
		return
	}

	clonedMap, ok := cloned.(map[string]interface{})
	if !ok {
		resp := &shared.CloneResponse{
			Success: false,
			Error:   "克隆结果类型不正确",
		}
		writeJSON(w, http.StatusInternalServerError, resp)
		return
	}

	resp := &shared.CloneResponse{
		Success: true,
		Name:    req.Name,
		Data:    clonedMap,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleDeepCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var req shared.DeepCopyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	copied, err := deepcopy.DeepCopy(req.Data)
	if err != nil {
		resp := &shared.DeepCopyResponse{
			Success: false,
			Error:   err.Error(),
		}
		writeJSON(w, http.StatusInternalServerError, resp)
		return
	}

	copiedMap, ok := copied.(map[string]interface{})
	if !ok {
		resp := &shared.DeepCopyResponse{
			Success: false,
			Error:   "拷贝结果类型不正确",
		}
		writeJSON(w, http.StatusInternalServerError, resp)
		return
	}

	resp := &shared.DeepCopyResponse{
		Success: true,
		Data:    copiedMap,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var req shared.CompareRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	mode := deepcopy.CompareMode(req.Mode)
	if mode != deepcopy.CompareShallow && mode != deepcopy.CompareDeep && mode != deepcopy.CompareReference {
		mode = deepcopy.CompareDeep
	}

	equal, err := deepcopy.Compare(req.A, req.B, mode)
	if err != nil {
		resp := &shared.CompareResponse{
			Success: false,
			Mode:    string(mode),
			Error:   err.Error(),
		}
		writeJSON(w, http.StatusInternalServerError, resp)
		return
	}

	resp := &shared.CompareResponse{
		Success: true,
		Equal:   equal,
		Mode:    string(mode),
	}
	writeJSON(w, http.StatusOK, resp)
}

type errMsg string

func (e *errMsg) Error() string {
	return string(*e)
}

var errNameRequired errMsg = "name is required"
