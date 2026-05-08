package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/example/url-sign-verify/pkg/models"
	"github.com/example/url-sign-verify/pkg/signature"
)

var (
	users     = make(map[int]*models.UserInfo)
	usersMu   sync.Mutex
	nextUserID = 1
)

func main() {
	config := signature.DefaultVerifierConfig()
	
	config.AddApp(
		"partner-a", 
		"secret-key-for-partner-a", 
		[]string{"/api/users", "/api/health"},
	)
	
	config.AddApp(
		"partner-b", 
		"secret-key-for-partner-b", 
		[]string{"/api/users"},
	)
	
	config.IsProduction = false
	
	middleware := signature.NewMiddleware(config)
	
	http.HandleFunc("/api/health", middleware.Handler(handleHealth))
	http.HandleFunc("/api/users", middleware.Handler(handleUsers))
	
	println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"app":    "signature-verify-service",
	}
	
	writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(response))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetUsers(w, r)
	case http.MethodPost:
		handleCreateUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	usersMu.Lock()
	defer usersMu.Unlock()
	
	userList := make([]*models.UserInfo, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	
	debugInfo := signature.GetDebugInfo(r.Context())
	
	if debugInfo != nil {
		modelDebug := &models.DebugInfo{
			SignString: debugInfo.SignString,
			Signature:  debugInfo.Signature,
			ClientSign: debugInfo.ClientSign,
		}
		writeJSONResponse(w, http.StatusOK, models.NewDebugResponse(userList, modelDebug))
	} else {
		writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(userList))
	}
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	
	usersMu.Lock()
	defer usersMu.Unlock()
	
	user := &models.UserInfo{
		ID:    nextUserID,
		Name:  req.Name,
		Age:   req.Age,
		Email: req.Email,
	}
	
	users[user.ID] = user
	nextUserID++
	
	debugInfo := signature.GetDebugInfo(r.Context())
	
	if debugInfo != nil {
		modelDebug := &models.DebugInfo{
			SignString: debugInfo.SignString,
			Signature:  debugInfo.Signature,
			ClientSign: debugInfo.ClientSign,
		}
		writeJSONResponse(w, http.StatusCreated, models.NewDebugResponse(user, modelDebug))
	} else {
		writeJSONResponse(w, http.StatusCreated, models.NewSuccessResponse(user))
	}
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getIntParam(r *http.Request, name string, defaultValue int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultValue
	}
	if parsed, err := strconv.Atoi(val); err == nil {
		return parsed
	}
	return defaultValue
}
