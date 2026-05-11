package main

import (
	"constructionms/common"
	"constructionms/core"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var store = core.NewStore()

type Server struct {
	store *core.Store
}

func NewServer() *Server {
	return &Server{store: core.NewStore()}
}

func (s *Server) writeResponse(w http.ResponseWriter, resp *common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Code)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	var req common.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "请求参数解析失败"))
		return
	}

	project, err := s.store.CreateProject(&req)
	if err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, err.Error()))
		return
	}

	s.writeResponse(w, common.Success(project))
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	projects := s.store.ListProjects()

	summaries := make([]common.ProjectSummary, 0, len(projects))
	for _, p := range projects {
		progress, _ := s.store.GetOverallProgress(p.ID)
		isRisk, _ := s.store.IsRiskProject(p.ID)
		
		summaries = append(summaries, common.ProjectSummary{
			ID:             p.ID,
			Name:           p.Name,
			ProjectManager: p.ProjectManager,
			Progress:       progress,
			Status:         getProjectStatus(progress),
			IsRiskProject:  isRisk,
		})
	}

	s.writeResponse(w, common.Success(common.ProjectListResponse{Projects: summaries}))
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	if projectID == "" {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "项目ID不能为空"))
		return
	}

	project, err := s.store.GetProject(projectID)
	if err != nil {
		s.writeResponse(w, common.Error(http.StatusNotFound, err.Error()))
		return
	}

	progress, _ := s.store.GetOverallProgress(projectID)
	isRisk, _ := s.store.IsRiskProject(projectID)

	budgetSummary := make([]common.BudgetCategory, 0, 4)
	for _, cat := range common.ExpenseCategories {
		budget := s.store.GetCategoryBudget(project.ContractFee, cat)
		actual, _ := s.store.GetCategoryActual(projectID, cat)
		budgetSummary = append(budgetSummary, common.BudgetCategory{
			Name:       cat,
			Budget:     budget,
			Actual:     actual,
			OverBudget: actual > budget,
		})
	}

	detail := common.ProjectDetailResponse{
		Project:         project,
		OverallProgress: progress,
		IsRiskProject:   isRisk,
		BudgetSummary:   budgetSummary,
	}

	s.writeResponse(w, common.Success(detail))
}

func (s *Server) handleUpdateProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/progress")
	if projectID == "" {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "项目ID不能为空"))
		return
	}

	var req common.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "请求参数解析失败"))
		return
	}

	if err := s.store.UpdateProgress(projectID, &req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, err.Error()))
		return
	}

	s.writeResponse(w, common.Success(map[string]string{"status": "success"}))
}

func (s *Server) handleAddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/expenses")
	if projectID == "" {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "项目ID不能为空"))
		return
	}

	var req common.AddExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "请求参数解析失败"))
		return
	}

	expense, err := s.store.AddExpense(projectID, &req)
	if err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, err.Error()))
		return
	}

	overBudget, _ := s.store.IsCategoryOverBudget(projectID, req.Category)
	response := map[string]interface{}{
		"expense":    expense,
		"over_budget": overBudget,
	}

	s.writeResponse(w, common.Success(response))
}

func (s *Server) handleAddQualityCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/quality")
	if projectID == "" {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "项目ID不能为空"))
		return
	}

	var req common.AddQualityCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "请求参数解析失败"))
		return
	}

	check, err := s.store.AddQualityCheck(projectID, &req)
	if err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, err.Error()))
		return
	}

	isRisk, _ := s.store.IsRiskProject(projectID)
	response := map[string]interface{}{
		"quality_check": check,
		"is_risk_project": isRisk,
	}

	s.writeResponse(w, common.Success(response))
}

func (s *Server) handleRectification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/rectification")
	if projectID == "" {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "项目ID不能为空"))
		return
	}

	var req common.RectificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, "请求参数解析失败"))
		return
	}

	if err := s.store.AddRectification(projectID, req.CheckID, &req); err != nil {
		s.writeResponse(w, common.Error(http.StatusBadRequest, err.Error()))
		return
	}

	s.writeResponse(w, common.Success(map[string]string{"status": "success"}))
}

func getProjectStatus(progress float64) string {
	if progress == 0 {
		return "未开始"
	} else if progress >= 100 {
		return "已完成"
	}
	return "进行中"
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "服务端口 (环境变量 SERVER_PORT 优先)")
	flag.Parse()

	envPort := os.Getenv("SERVER_PORT")
	if envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	server := NewServer()

	mux := http.NewServeMux()
	
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			server.handleListProjects(w, r)
		} else if r.Method == http.MethodPost {
			server.handleCreateProject(w, r)
		} else {
			server.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
		}
	})

	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		if strings.Contains(path, "/progress") {
			server.handleUpdateProgress(w, r)
			return
		}
		if strings.Contains(path, "/expenses") {
			server.handleAddExpense(w, r)
			return
		}
		if strings.Contains(path, "/quality") {
			server.handleAddQualityCheck(w, r)
			return
		}
		if strings.Contains(path, "/rectification") {
			server.handleRectification(w, r)
			return
		}

		if r.Method == http.MethodGet {
			server.handleGetProject(w, r)
			return
		}

		server.writeResponse(w, common.Error(http.StatusMethodNotAllowed, "方法不允许"))
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("建筑工程项目管理系统服务端启动，监听端口: %d\n", port)
	fmt.Printf("启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
