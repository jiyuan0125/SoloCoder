package main

import (
	"flag"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type BudgetCategory string

const (
	EquipmentFee       BudgetCategory = "equipment"
	MaterialFee        BudgetCategory = "material"
	TravelFee          BudgetCategory = "travel"
	LaborFee           BudgetCategory = "labor"
	ExpertConsultFee   BudgetCategory = "expert"
	OtherFee           BudgetCategory = "other"
)

type ReimbursementStatus string

const (
	Pending   ReimbursementStatus = "pending"
	Approved  ReimbursementStatus = "approved"
	Rejected  ReimbursementStatus = "rejected"
)

type Project struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	Principal     string                    `json:"principal"`
	StartDate     string                    `json:"start_date"`
	EndDate       string                    `json:"end_date"`
	TotalAmount   int64                     `json:"total_amount"`
	Budget        map[BudgetCategory]int64  `json:"budget"`
	Status        string                    `json:"status"`
	CreatedAt     time.Time                 `json:"created_at"`
}

type Reimbursement struct {
	ID          string                  `json:"id"`
	ProjectID   string                  `json:"project_id"`
	Category    BudgetCategory          `json:"category"`
	Amount      int64                   `json:"amount"`
	Reason      string                  `json:"reason"`
	Date        string                  `json:"date"`
	Status      ReimbursementStatus     `json:"status"`
	RejectReason string                 `json:"reject_reason,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type FinalReport struct {
	ProjectID     string                              `json:"project_id"`
	ProjectName   string                              `json:"project_name"`
	TotalBudget   int64                               `json:"total_budget"`
	TotalSpent    int64                               `json:"total_spent"`
	ExecutionRate float64                             `json:"execution_rate"`
	IsAbnormal    bool                                `json:"is_abnormal"`
	Details       []CategoryFinalDetail               `json:"details"`
	GeneratedAt   time.Time                           `json:"generated_at"`
}

type CategoryFinalDetail struct {
	Category      BudgetCategory `json:"category"`
	CategoryName  string         `json:"category_name"`
	Budget        int64          `json:"budget"`
	Spent         int64          `json:"spent"`
	ExecutionRate float64        `json:"execution_rate"`
}

var (
	projects         = make(map[string]*Project)
	reimbursements   = make(map[string]*Reimbursement)
	projectReimbs    = make(map[string][]*Reimbursement)
	mu               sync.RWMutex
	projectCounter   = 0
	reimbursementCounter = 0
)

var categoryNames = map[BudgetCategory]string{
	EquipmentFee:     "设备费",
	MaterialFee:      "材料费",
	TravelFee:        "差旅费",
	LaborFee:         "劳务费",
	ExpertConsultFee: "专家咨询费",
	OtherFee:         "其他费用",
}

func yuanToFen(yuanStr string) (int64, error) {
	yuanStr = strings.TrimSpace(yuanStr)
	if strings.Contains(yuanStr, ".") {
		parts := strings.Split(yuanStr, ".")
		if len(parts) > 2 {
			return 0, nil
		}
		intPart := parts[0]
		decimalPart := parts[1]
		if len(decimalPart) > 2 {
			return 0, nil
		}
		for len(decimalPart) < 2 {
			decimalPart += "0"
		}
		intVal, err := strconv.ParseInt(intPart, 10, 64)
		if err != nil {
			return 0, err
		}
		decVal, err := strconv.ParseInt(decimalPart, 10, 64)
		if err != nil {
			return 0, err
		}
		return intVal*100 + decVal, nil
	}
	val, err := strconv.ParseInt(yuanStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return val * 100, nil
}

func fenToYuan(fen int64) string {
	if fen < 0 {
		neg := true
		fen = -fen
		yuan := fen / 100
		cents := fen % 100
		if neg {
			return "-" + strconv.FormatInt(yuan, 10) + "." + padZero(cents)
		}
	}
	yuan := fen / 100
	cents := fen % 100
	return strconv.FormatInt(yuan, 10) + "." + padZero(cents)
}

func padZero(n int64) string {
	if n < 10 {
		return "0" + strconv.FormatInt(n, 10)
	}
	return strconv.FormatInt(n, 10)
}

func generateProjectID() string {
	mu.Lock()
	defer mu.Unlock()
	projectCounter++
	return "P" + time.Now().Format("20060102") + strconv.Itoa(projectCounter)
}

func generateReimbursementID() string {
	mu.Lock()
	defer mu.Unlock()
	reimbursementCounter++
	return "R" + time.Now().Format("20060102") + strconv.Itoa(reimbursementCounter)
}

func calculateExecutionRate(spent, budget int64) float64 {
	if budget == 0 {
		if spent == 0 {
			return 100.0
		}
		return 0.0
	}
	rate := float64(spent) / float64(budget) * 100
	return math.Round(rate*100) / 100
}

func getProjectSpent(projectID string) map[BudgetCategory]int64 {
	spent := map[BudgetCategory]int64{
		EquipmentFee:     0,
		MaterialFee:      0,
		TravelFee:        0,
		LaborFee:         0,
		ExpertConsultFee: 0,
		OtherFee:         0,
	}
	for _, r := range projectReimbs[projectID] {
		if r.Status == Approved {
			spent[r.Category] += r.Amount
		}
	}
	return spent
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/api/projects", createProject)
	r.GET("/api/projects", listProjects)
	r.GET("/api/projects/:id", getProject)
	r.POST("/api/projects/:id/close", closeProject)

	r.POST("/api/reimbursements", createReimbursement)
	r.GET("/api/reimbursements", listReimbursements)
	r.GET("/api/reimbursements/pending", listPendingReimbursements)
	r.POST("/api/reimbursements/:id/approve", approveReimbursement)
	r.POST("/api/reimbursements/:id/reject", rejectReimbursement)
	r.POST("/api/reimbursements/:id/resubmit", resubmitReimbursement)

	r.GET("/api/projects/:id/final-report", getFinalReport)

	return r
}

func createProject(c *gin.Context) {
	var req struct {
		Name      string            `json:"name" binding:"required"`
		Principal string            `json:"principal" binding:"required"`
		StartDate string            `json:"start_date" binding:"required"`
		EndDate   string            `json:"end_date" binding:"required"`
		Total     string            `json:"total_amount" binding:"required"`
		Budget    map[string]string `json:"budget" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	totalFen, err := yuanToFen(req.Total)
	if err != nil || totalFen <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的经费总额"})
		return
	}

	budgetFen := make(map[BudgetCategory]int64)
	var budgetSum int64 = 0
	for cat, val := range req.Budget {
		category := BudgetCategory(cat)
		if _, ok := categoryNames[category]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预算类别"})
			return
		}
		fen, err := yuanToFen(val)
		if err != nil || fen < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预算金额"})
			return
		}
		budgetFen[category] = fen
		budgetSum += fen
	}

	if budgetSum != totalFen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "预算各类别之和必须精确等于总额"})
		return
	}

	project := &Project{
		ID:          generateProjectID(),
		Name:        req.Name,
		Principal:   req.Principal,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		TotalAmount: totalFen,
		Budget:      budgetFen,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	mu.Lock()
	projects[project.ID] = project
	projectReimbs[project.ID] = []*Reimbursement{}
	mu.Unlock()

	c.JSON(http.StatusCreated, project)
}

func listProjects(c *gin.Context) {
	mu.RLock()
	result := make([]*Project, 0, len(projects))
	for _, p := range projects {
		result = append(result, p)
	}
	mu.RUnlock()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	c.JSON(http.StatusOK, result)
}

func getProject(c *gin.Context) {
	id := c.Param("id")
	mu.RLock()
	project, exists := projects[id]
	if !exists {
		mu.RUnlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}
	reimbs := make([]*Reimbursement, len(projectReimbs[id]))
	copy(reimbs, projectReimbs[id])
	mu.RUnlock()

	spent := getProjectSpent(id)
	result := gin.H{
		"project": project,
		"budget_details": gin.H{
			"total":   project.TotalAmount,
			"budget":  project.Budget,
			"spent":   spent,
		},
		"reimbursements": reimbs,
	}

	c.JSON(http.StatusOK, result)
}

func closeProject(c *gin.Context) {
	id := c.Param("id")
	mu.Lock()
	defer mu.Unlock()
	project, exists := projects[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}
	if project.Status == "closed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目已结题"})
		return
	}
	project.Status = "closed"
	c.JSON(http.StatusOK, project)
}

func createReimbursement(c *gin.Context) {
	var req struct {
		ProjectID string `json:"project_id" binding:"required"`
		Category  string `json:"category" binding:"required"`
		Amount    string `json:"amount" binding:"required"`
		Reason    string `json:"reason" binding:"required"`
		Date      string `json:"date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	category := BudgetCategory(req.Category)
	if _, ok := categoryNames[category]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的费用类别"})
		return
	}

	amountFen, err := yuanToFen(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的金额"})
		return
	}

	if amountFen <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "金额必须为正数"})
		return
	}

	mu.RLock()
	project, exists := projects[req.ProjectID]
	if !exists {
		mu.RUnlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}
	if project.Status == "closed" {
		mu.RUnlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目已结题，无法提交报销"})
		return
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	spent := getProjectSpent(req.ProjectID)
	if spent[category]+amountFen > project.Budget[category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "累计报销超出预算"})
		return
	}

	reimb := &Reimbursement{
		ID:         generateReimbursementID(),
		ProjectID:  req.ProjectID,
		Category:   category,
		Amount:     amountFen,
		Reason:     req.Reason,
		Date:       req.Date,
		Status:     Pending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	reimbursements[reimb.ID] = reimb
	projectReimbs[req.ProjectID] = append(projectReimbs[req.ProjectID], reimb)

	c.JSON(http.StatusCreated, reimb)
}

func listReimbursements(c *gin.Context) {
	projectID := c.Query("project_id")
	mu.RLock()
	defer mu.RUnlock()

	var result []*Reimbursement
	if projectID != "" {
		result = make([]*Reimbursement, len(projectReimbs[projectID]))
		copy(result, projectReimbs[projectID])
	} else {
		result = make([]*Reimbursement, 0, len(reimbursements))
		for _, r := range reimbursements {
			result = append(result, r)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	c.JSON(http.StatusOK, result)
}

func listPendingReimbursements(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]*Reimbursement, 0)
	for _, r := range reimbursements {
		if r.Status == Pending {
			result = append(result, r)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	c.JSON(http.StatusOK, result)
}

func approveReimbursement(c *gin.Context) {
	id := c.Param("id")
	mu.Lock()
	defer mu.Unlock()

	reimb, exists := reimbursements[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "报销单不存在"})
		return
	}

	if reimb.Status != Pending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有待审批状态的报销单才能审批"})
		return
	}

	project := projects[reimb.ProjectID]
	spent := getProjectSpent(reimb.ProjectID)
	if spent[reimb.Category]+reimb.Amount > project.Budget[reimb.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "累计报销超出预算"})
		return
	}

	reimb.Status = Approved
	reimb.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, reimb)
}

func rejectReimbursement(c *gin.Context) {
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写驳回理由"})
		return
	}

	id := c.Param("id")
	mu.Lock()
	defer mu.Unlock()

	reimb, exists := reimbursements[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "报销单不存在"})
		return
	}

	if reimb.Status != Pending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有待审批状态的报销单才能驳回"})
		return
	}

	reimb.Status = Rejected
	reimb.RejectReason = req.Reason
	reimb.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, reimb)
}

func resubmitReimbursement(c *gin.Context) {
	var req struct {
		Amount string `json:"amount"`
		Reason string `json:"reason"`
		Date   string `json:"date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	id := c.Param("id")
	mu.Lock()
	defer mu.Unlock()

	reimb, exists := reimbursements[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "报销单不存在"})
		return
	}

	if reimb.Status != Rejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有已驳回状态的报销单才能重新提交"})
		return
	}

	if req.Amount != "" {
		amountFen, err := yuanToFen(req.Amount)
		if err != nil || amountFen <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的金额"})
			return
		}
		project := projects[reimb.ProjectID]
		spent := getProjectSpent(reimb.ProjectID)
		if spent[reimb.Category]+amountFen > project.Budget[reimb.Category] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "累计报销超出预算"})
			return
		}
		reimb.Amount = amountFen
	}

	if req.Reason != "" {
		reimb.Reason = req.Reason
	}
	if req.Date != "" {
		reimb.Date = req.Date
	}

	reimb.Status = Pending
	reimb.RejectReason = ""
	reimb.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, reimb)
}

func getFinalReport(c *gin.Context) {
	id := c.Param("id")
	mu.RLock()
	defer mu.RUnlock()

	project, exists := projects[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}
	if project.Status != "closed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目未结题，无法生成决算报告"})
		return
	}

	spent := getProjectSpent(id)
	var totalSpent int64 = 0
	for _, s := range spent {
		totalSpent += s
	}

	details := make([]CategoryFinalDetail, 0)
	categories := []BudgetCategory{EquipmentFee, MaterialFee, TravelFee, LaborFee, ExpertConsultFee, OtherFee}
	for _, cat := range categories {
		details = append(details, CategoryFinalDetail{
			Category:      cat,
			CategoryName:  categoryNames[cat],
			Budget:        project.Budget[cat],
			Spent:         spent[cat],
			ExecutionRate: calculateExecutionRate(spent[cat], project.Budget[cat]),
		})
	}

	totalRate := calculateExecutionRate(totalSpent, project.TotalAmount)

	report := &FinalReport{
		ProjectID:     project.ID,
		ProjectName:   project.Name,
		TotalBudget:   project.TotalAmount,
		TotalSpent:    totalSpent,
		ExecutionRate: totalRate,
		IsAbnormal:    totalRate < 30.0,
		Details:       details,
		GeneratedAt:   time.Now(),
	}

	c.JSON(http.StatusOK, report)
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8300, "服务端口")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	r := setupRouter()
	addr := ":" + strconv.Itoa(port)
	log.Printf("服务启动于 %s", addr)
	log.Fatal(r.Run(addr))
}
