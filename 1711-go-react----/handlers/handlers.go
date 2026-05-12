package handlers

import (
	"fmt"
	"genedeck/models"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	samples     = make(map[string]*models.Sample)
	samplesMu   sync.RWMutex
	reports     = make(map[string]*models.Report)
	reportsMu   sync.RWMutex
	todos       = make(map[string]*models.Todo)
	todosMu     sync.RWMutex
	units       = make(map[string]*models.Unit)
	unitsMu     sync.RWMutex
	testCatalog = make(map[string]*models.TestItem)
	testMu      sync.RWMutex
	sampleSeqMu sync.Mutex
)

func InitDB() {
	unitsMu.Lock()
	units["unit001"] = &models.Unit{ID: "unit001", Name: "北京协和医院", Contact: "张医生", Phone: "010-12345678"}
	units["unit002"] = &models.Unit{ID: "unit002", Name: "上海瑞金医院", Contact: "李医生", Phone: "021-87654321"}
	units["unit003"] = &models.Unit{ID: "unit003", Name: "广州中山医院", Contact: "王医生", Phone: "020-11112222"}
	unitsMu.Unlock()

	testMu.Lock()
	testCatalog["GEN001"] = &models.TestItem{Code: "GEN001", Name: "遗传病筛查", Description: "检测常见遗传性疾病相关基因位点", Price: 1500.0, ReportDays: 7}
	testCatalog["GEN002"] = &models.TestItem{Code: "GEN002", Name: "肿瘤基因检测", Description: "检测肿瘤易感基因及驱动突变", Price: 3800.0, ReportDays: 14}
	testCatalog["GEN003"] = &models.TestItem{Code: "GEN003", Name: "药物基因组学", Description: "检测药物代谢相关基因，指导个性化用药", Price: 1200.0, ReportDays: 5}
	testCatalog["GEN004"] = &models.TestItem{Code: "GEN004", Name: "营养基因组", Description: "检测营养代谢相关基因，提供个性化膳食建议", Price: 980.0, ReportDays: 5}
	testCatalog["GEN005"] = &models.TestItem{Code: "GEN005", Name: "心血管疾病风险评估", Description: "评估心血管疾病遗传风险", Price: 1800.0, ReportDays: 7}
	testCatalog["GEN006"] = &models.TestItem{Code: "GEN006", Name: "全外显子测序", Description: "全外显子组测序分析", Price: 8500.0, ReportDays: 21}
	testMu.Unlock()
}

func getRole(c *gin.Context) models.Role {
	role := c.GetHeader("X-Role")
	if role == "" {
		return models.RoleAdmin
	}
	return models.Role(role)
}

func getUnitID(c *gin.Context) string {
	return c.GetHeader("X-Unit-ID")
}

func maskString(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= 1 {
		return "*"
	}
	if len(runes) <= 3 {
		return string(runes[0]) + "**"
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
}

func maskIDCard(id string) string {
	if len(id) < 6 {
		return maskString(id)
	}
	return id[:6] + "********" + id[14:]
}

func maskPhone(p string) string {
	if len(p) < 7 {
		return maskString(p)
	}
	return p[:3] + "****" + p[7:]
}

func applyPrivacy(s *models.Sample, role models.Role) *models.Sample {
	if s == nil {
		return nil
	}
	copied := *s
	if role == models.RoleTechnician {
		copied.Patient.Name = "***"
		copied.Patient.IDCard = "***"
		copied.Patient.Phone = "***"
	} else if role == models.RoleSubmitter {
		copied.Patient.IDCard = maskIDCard(copied.Patient.IDCard)
		copied.Patient.Phone = maskPhone(copied.Patient.Phone)
	}
	return &copied
}

func applyReportPrivacy(r *models.Report, role models.Role) *models.Report {
	if r == nil {
		return nil
	}
	copied := *r
	if role == models.RoleSubmitter {
		for i := range copied.Content.TestItems {
			if copied.Content.TestItems[i].RawData != "" {
				copied.Content.TestItems[i].RawData = ""
			}
		}
	}
	return &copied
}

func canViewArchived(role models.Role) bool {
	return role == models.RoleAdmin
}

func canAccessSample(s *models.Sample, role models.Role, unitID string) bool {
	if s.Archived && !canViewArchived(role) {
		return false
	}
	if role == models.RoleSubmitter && s.UnitID != unitID {
		return false
	}
	return true
}

func generateSampleCode() string {
	sampleSeqMu.Lock()
	defer sampleSeqMu.Unlock()
	now := time.Now()
	dateStr := now.Format("20060102")
	var maxSeq int
	for _, s := range samples {
		if strings.HasPrefix(s.SampleCode, "GT"+dateStr) {
			seqPart := s.SampleCode[len("GT"+dateStr):]
			if seq, err := strconv.Atoi(seqPart); err == nil && seq > maxSeq {
				maxSeq = seq
			}
		}
	}
	maxSeq++
	return fmt.Sprintf("GT%s%04d", dateStr, maxSeq)
}

func calculatePrice(items []models.TestItem) (total, discount, final float64) {
	for _, item := range items {
		if t, ok := testCatalog[item.Code]; ok {
			total += t.Price
		}
	}
	count := len(items)
	if count >= 5 {
		discount = 0.85
	} else if count >= 3 {
		discount = 0.90
	} else {
		discount = 1.0
	}
	final = math.Round(total*discount*100) / 100
	return total, discount, final
}

func isValidStatusTransition(from, to models.SampleStatus) bool {
	order := []models.SampleStatus{
		models.SampleStatusReceived,
		models.SampleStatusTesting,
		models.SampleStatusTestComplete,
		models.SampleStatusReportGenerating,
		models.SampleStatusReported,
	}
	fromIdx := -1
	toIdx := -1
	for i, s := range order {
		if s == from {
			fromIdx = i
		}
		if s == to {
			toIdx = i
		}
	}
	return toIdx == fromIdx+1
}

func createTodo(typ models.TodoType, sampleID, sampleCode, assignee string) {
	titleMap := map[models.TodoType]string{
		models.TodoArrangeTest:  "安排检测",
		models.TodoReviewReport: "审核报告",
		models.TodoNotifyClient: "通知客户取报告",
	}
	todosMu.Lock()
	defer todosMu.Unlock()
	todo := &models.Todo{
		ID:         uuid.New().String(),
		Type:       typ,
		Title:      titleMap[typ],
		SampleID:   sampleID,
		SampleCode: sampleCode,
		Assignee:   assignee,
		Completed:  false,
		CreatedAt:  time.Now(),
	}
	todos[todo.ID] = todo
}

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/catalog/test-items", listTestCatalog)
		api.GET("/units", listUnits)

		s := api.Group("/samples")
		{
			s.POST("", createSample)
			s.GET("", listSamples)
			s.GET("/:id", getSample)
			s.PATCH("/:id/status", updateSampleStatus)
			s.PATCH("/:id/test-items/:code", updateTestResult)
			s.GET("/:id/sub", getSampleSubResources)
		}

		rep := api.Group("/reports")
		{
			rep.GET("", listReports)
			rep.GET("/:id", getReport)
			rep.POST("/:id/submit", submitReport)
			rep.POST("/:id/review", reviewReport)
			rep.POST("/:id/supplement", createSupplementReport)
			rep.GET("/:id/sub", getReportSubResources)
		}

		t := api.Group("/todos")
		{
			t.GET("", listTodos)
			t.PATCH("/:id/complete", completeTodo)
		}

		r2 := api.Group("/r")
		{
			r2.GET("", listGenericResources)
			r2.GET("/:id", getGenericResource)
			r2.GET("/:id/sub", getGenericSubResources)
		}
	}
}

type CreateSampleRequest struct {
	SampleType      models.SampleType `json:"sample_type" binding:"required"`
	CollectionDate  string            `json:"collection_date"`
	UnitID          string            `json:"unit_id" binding:"required"`
	Submitter       string            `json:"submitter" binding:"required"`
	Patient         models.Patient    `json:"patient" binding:"required"`
	TestItemCodes   []string          `json:"test_item_codes" binding:"required"`
}

func createSample(c *gin.Context) {
	var req CreateSampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unitsMu.RLock()
	unit, ok := units[req.UnitID]
	unitsMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		return
	}

	testMu.RLock()
	for _, code := range req.TestItemCodes {
		if _, ok := testCatalog[code]; !ok {
			testMu.RUnlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid test item code: %s", code)})
			return
		}
	}
	testMu.RUnlock()

	samplesMu.Lock()
	code := generateSampleCode()
	for _, existing := range samples {
		if existing.SampleCode == code {
			samplesMu.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": "sample code conflict"})
			return
		}
	}

	var items []models.TestItem
	for _, cd := range req.TestItemCodes {
		tmpl := testCatalog[cd]
		items = append(items, models.TestItem{
			Code:        tmpl.Code,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Price:       tmpl.Price,
			ReportDays:  tmpl.ReportDays,
			Completed:   false,
		})
	}

	total, disc, final := calculatePrice(items)

	sample := &models.Sample{
		ID:             uuid.New().String(),
		SampleCode:     code,
		SampleType:     req.SampleType,
		ReceivedDate:   time.Now().Format("2006-01-02"),
		CollectionDate: req.CollectionDate,
		UnitID:         req.UnitID,
		UnitName:       unit.Name,
		Submitter:      req.Submitter,
		Status:         models.SampleStatusReceived,
		Patient:        req.Patient,
		TestItems:      items,
		TotalPrice:     total,
		Discount:       disc,
		FinalPrice:     final,
		Archived:       false,
		CreatedAt:      time.Now(),
	}
	samples[sample.ID] = sample
	samplesMu.Unlock()

	createTodo(models.TodoArrangeTest, sample.ID, sample.SampleCode, "实验室组长")

	c.JSON(http.StatusCreated, gin.H{"id": sample.ID, "sample_code": sample.SampleCode})
}

func listSamples(c *gin.Context) {
	role := getRole(c)
	unitID := getUnitID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	samplesMu.RLock()
	list := make([]*models.Sample, 0, len(samples))
	for _, s := range samples {
		if canAccessSample(s, role, unitID) {
			list = append(list, applyPrivacy(s, role))
		}
	}
	samplesMu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  size,
		"items": list[start:end],
	})
}

func getSample(c *gin.Context) {
	id := c.Param("id")
	role := getRole(c)
	unitID := getUnitID(c)

	samplesMu.RLock()
	s, ok := samples[id]
	samplesMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
		return
	}

	if !canAccessSample(s, role, unitID) {
		if s.Archived {
			c.JSON(http.StatusForbidden, gin.H{"error": "archived data only accessible by admin"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, applyPrivacy(s, role))
}

func getSampleSubResources(c *gin.Context) {
	id := c.Param("id")
	role := getRole(c)
	unitID := getUnitID(c)

	samplesMu.RLock()
	s, ok := samples[id]
	samplesMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
		return
	}

	if !canAccessSample(s, role, unitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	reportsMu.RLock()
	related := make([]*models.Report, 0)
	for _, r := range reports {
		if r.SampleID == id {
			related = append(related, applyReportPrivacy(r, role))
		}
	}
	reportsMu.RUnlock()

	todosMu.RLock()
	relatedTodos := make([]*models.Todo, 0)
	for _, t := range todos {
		if t.SampleID == id {
			relatedTodos = append(relatedTodos, t)
		}
	}
	todosMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"reports": related,
		"todos":   relatedTodos,
	})
}

type UpdateStatusRequest struct {
	NewStatus models.SampleStatus `json:"new_status" binding:"required"`
}

func updateSampleStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	samplesMu.Lock()
	s, ok := samples[id]
	if !ok {
		samplesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
		return
	}

	if !isValidStatusTransition(s.Status, req.NewStatus) {
		samplesMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
		return
	}

	s.Status = req.NewStatus

	if req.NewStatus == models.SampleStatusTestComplete {
		allComplete := true
		for _, item := range s.TestItems {
			if !item.Completed {
				allComplete = false
				break
			}
		}
		if !allComplete {
			samplesMu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": "not all test items completed"})
			return
		}
	}

	if req.NewStatus == models.SampleStatusReportGenerating {
		generateReportForSample(s)
	}

	if req.NewStatus == models.SampleStatusTestComplete {
		createTodo(models.TodoReviewReport, s.ID, s.SampleCode, "审核员")
	}
	if req.NewStatus == models.SampleStatusReported {
		createTodo(models.TodoNotifyClient, s.ID, s.SampleCode, "客服人员")
	}

	samplesMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"status": s.Status})
}

type UpdateTestResultRequest struct {
	RawData        string `json:"raw_data"`
	ResultSummary  string `json:"result_summary"`
	RiskLevel      string `json:"risk_level"`
	RiskConclusion string `json:"risk_conclusion"`
}

func updateTestResult(c *gin.Context) {
	sampleID := c.Param("id")
	code := c.Param("code")

	var req UpdateTestResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	samplesMu.Lock()
	s, ok := samples[sampleID]
	if !ok {
		samplesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
		return
	}

	found := false
	for i := range s.TestItems {
		if s.TestItems[i].Code == code {
			s.TestItems[i].RawData = req.RawData
			s.TestItems[i].ResultSummary = req.ResultSummary
			s.TestItems[i].RiskLevel = req.RiskLevel
			s.TestItems[i].RiskConclusion = req.RiskConclusion
			s.TestItems[i].Completed = true
			s.TestItems[i].CompletedAt = time.Now().Format("2006-01-02 15:04:05")
			found = true
			break
		}
	}
	if !found {
		samplesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "test item not found in sample"})
		return
	}
	samplesMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func listTestCatalog(c *gin.Context) {
	testMu.RLock()
	list := make([]*models.TestItem, 0, len(testCatalog))
	for _, t := range testCatalog {
		list = append(list, t)
	}
	testMu.RUnlock()
	sort.Slice(list, func(i, j int) bool {
		return list[i].Code < list[j].Code
	})
	c.JSON(http.StatusOK, list)
}

func listUnits(c *gin.Context) {
	unitsMu.RLock()
	list := make([]*models.Unit, 0, len(units))
	for _, u := range units {
		list = append(list, u)
	}
	unitsMu.RUnlock()
	c.JSON(http.StatusOK, list)
}

func generateReportForSample(s *models.Sample) {
	testItems := make([]models.TestItemSummary, 0, len(s.TestItems))
	highRisks := make([]string, 0)
	lowRisks := make([]string, 0)
	overallLevel := "低风险"

	for _, item := range s.TestItems {
		summary := models.TestItemSummary{
			Code:          item.Code,
			Name:          item.Name,
			ResultSummary: item.ResultSummary,
			RiskLevel:     item.RiskLevel,
			RawData:       item.RawData,
		}
		testItems = append(testItems, summary)

		if item.RiskLevel == "高风险" || item.RiskLevel == "high" {
			highRisks = append(highRisks, item.Name+": "+item.RiskConclusion)
			overallLevel = "高风险"
		} else if item.RiskLevel == "中风险" || item.RiskLevel == "medium" {
			if overallLevel != "高风险" {
				overallLevel = "中风险"
			}
		} else {
			lowRisks = append(lowRisks, item.Name+": "+item.RiskConclusion)
		}
	}

	content := models.ReportContent{
		SampleInfo: models.SampleSummary{
			SampleCode:     s.SampleCode,
			SampleType:     string(s.SampleType),
			ReceivedDate:   s.ReceivedDate,
			CollectionDate: s.CollectionDate,
			UnitName:       s.UnitName,
			Submitter:      s.Submitter,
			PatientName:    s.Patient.Name,
		},
		TestItems: testItems,
		RiskAssessment: models.RiskAssessment{
			OverallLevel: overallLevel,
			HighRisks:    highRisks,
			LowRisks:     lowRisks,
		},
		Recommendations: generateRecommendations(overallLevel, highRisks),
		GeneratedAt:     time.Now().Format("2006-01-02 15:04:05"),
	}

	reportsMu.Lock()
	var maxVersion int
	for _, r := range reports {
		if r.SampleID == s.ID && !r.IsSupplementary {
			if r.Version > maxVersion {
				maxVersion = r.Version
			}
		}
	}

	report := &models.Report{
		ID:              uuid.New().String(),
		SampleID:        s.ID,
		SampleCode:      s.SampleCode,
		Version:         maxVersion + 1,
		Status:          models.ReportStatusDraft,
		ApprovalStatus:  models.ApprovalSubmitted,
		Content:         content,
		ReviewComments:  make([]models.ReviewComment, 0),
		CreatedAt:       time.Now(),
		IsSupplementary: false,
	}
	reports[report.ID] = report
	reportsMu.Unlock()
}

func generateRecommendations(level string, highRisks []string) string {
	if level == "高风险" {
		return "建议：1. 针对高风险项目，建议咨询相关专科医生进行进一步检查；2. 调整生活方式，降低疾病风险；3. 定期进行相关体检和监测；4. 家族成员可考虑进行相关基因检测。"
	}
	if level == "中风险" {
		return "建议：1. 保持健康的生活方式；2. 定期进行体检；3. 注意风险因素的控制和管理。"
	}
	return "建议：1. 继续保持健康的生活习惯；2. 定期进行常规体检；3. 保持良好的心态和均衡的饮食。"
}

func listReports(c *gin.Context) {
	role := getRole(c)
	unitID := getUnitID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	samplesMu.RLock()
	sampleMap := make(map[string]*models.Sample)
	for id, s := range samples {
		sampleMap[id] = s
	}
	samplesMu.RUnlock()

	reportsMu.RLock()
	list := make([]*models.Report, 0, len(reports))
	for _, r := range reports {
		if s, ok := sampleMap[r.SampleID]; ok {
			if !canAccessSample(s, role, unitID) {
				continue
			}
		}
		list = append(list, applyReportPrivacy(r, role))
	}
	reportsMu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  size,
		"items": list[start:end],
	})
}

func getReport(c *gin.Context) {
	id := c.Param("id")
	role := getRole(c)
	unitID := getUnitID(c)

	reportsMu.RLock()
	r, ok := reports[id]
	reportsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	samplesMu.RLock()
	s, ok := samples[r.SampleID]
	samplesMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
		return
	}

	if !canAccessSample(s, role, unitID) {
		if s.Archived {
			c.JSON(http.StatusForbidden, gin.H{"error": "archived data only accessible by admin"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, applyReportPrivacy(r, role))
}

func getReportSubResources(c *gin.Context) {
	id := c.Param("id")
	role := getRole(c)
	unitID := getUnitID(c)

	reportsMu.RLock()
	r, ok := reports[id]
	reportsMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	samplesMu.RLock()
	s, ok := samples[r.SampleID]
	samplesMu.RUnlock()
	if !ok || !canAccessSample(s, role, unitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	reportsMu.RLock()
	supplements := make([]*models.Report, 0)
	for _, rep := range reports {
		if rep.ParentReportID == id {
			supplements = append(supplements, applyReportPrivacy(rep, role))
		}
	}
	reportsMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"sample":         applyPrivacy(s, role),
		"supplement_reports": supplements,
	})
}

func submitReport(c *gin.Context) {
	id := c.Param("id")

	reportsMu.Lock()
	r, ok := reports[id]
	if !ok {
		reportsMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	if r.Status != models.ReportStatusDraft {
		reportsMu.Unlock()
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "only draft report can be submitted"})
		return
	}

	r.Status = models.ReportStatusPending
	r.ApprovalStatus = models.ApprovalFirst
	reportsMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": r.Status, "approval": r.ApprovalStatus})
}

type ReviewRequest struct {
	Action  string `json:"action" binding:"required"`
	Comment string `json:"comment"`
}

func reviewReport(c *gin.Context) {
	id := c.Param("id")
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reportsMu.Lock()
	r, ok := reports[id]
	if !ok {
		reportsMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	if r.Status == models.ReportStatusPublished {
		reportsMu.Unlock()
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "published report cannot be modified"})
		return
	}

	if req.Action == "reject" {
		r.ApprovalStatus = models.ApprovalRejected
		r.Status = models.ReportStatusDraft
		r.ReviewComments = append(r.ReviewComments, models.ReviewComment{
			Reviewer: "审核员",
			Comment:  req.Comment,
			Stage:    string(r.ApprovalStatus),
			Time:     time.Now().Format("2006-01-02 15:04:05"),
		})
		reportsMu.Unlock()
		c.JSON(http.StatusOK, gin.H{"status": r.Status, "approval": r.ApprovalStatus})
		return
	}

	if req.Action != "approve" {
		reportsMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	order := []models.ApprovalStatus{
		models.ApprovalFirst,
		models.ApprovalSecond,
		models.ApprovalFinal,
		models.ApprovalApproved,
	}

	var currentIdx int
	for i, s := range order {
		if s == r.ApprovalStatus {
			currentIdx = i
			break
		}
	}

	smallAmount := true
	if s, ok := samples[r.SampleID]; ok && s.FinalPrice >= 10000 {
		smallAmount = false
	}

	nextIdx := currentIdx + 1
	if smallAmount && r.ApprovalStatus == models.ApprovalFirst {
		nextIdx = 3
	}

	if nextIdx >= len(order) {
		reportsMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "already approved"})
		return
	}

	r.ApprovalStatus = order[nextIdx]
	if req.Comment != "" {
		r.ReviewComments = append(r.ReviewComments, models.ReviewComment{
			Reviewer: "审核员",
			Comment:  req.Comment,
			Stage:    string(r.ApprovalStatus),
			Time:     time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	if r.ApprovalStatus == models.ApprovalApproved {
		now := time.Now()
		r.Status = models.ReportStatusPublished
		r.PublishedAt = &now
		r.Amount = func() float64 {
			if s, ok := samples[r.SampleID]; ok {
				return s.FinalPrice
			}
			return 0
		}()

		samplesMu.Lock()
		if s, ok := samples[r.SampleID]; ok {
			s.Status = models.SampleStatusReported
			createTodo(models.TodoNotifyClient, s.ID, s.SampleCode, "客服人员")
		}
		samplesMu.Unlock()
	}

	reportsMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"status": r.Status, "approval": r.ApprovalStatus})
}

func createSupplementReport(c *gin.Context) {
	id := c.Param("id")

	reportsMu.RLock()
	parent, ok := reports[id]
	reportsMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	if parent.Status != models.ReportStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only published report can have supplement"})
		return
	}

	reportsMu.Lock()
	var maxVersion int
	for _, r := range reports {
		if r.ParentReportID == parent.ID || r.ID == parent.ID {
			if r.Version > maxVersion {
				maxVersion = r.Version
			}
		}
	}

	content := parent.Content
	content.GeneratedAt = time.Now().Format("2006-01-02 15:04:05")

	newReport := &models.Report{
		ID:              uuid.New().String(),
		SampleID:        parent.SampleID,
		SampleCode:      parent.SampleCode,
		Version:         maxVersion + 1,
		Status:          models.ReportStatusDraft,
		ApprovalStatus:  models.ApprovalSubmitted,
		Content:         content,
		ReviewComments:  make([]models.ReviewComment, 0),
		CreatedAt:       time.Now(),
		IsSupplementary: true,
		ParentReportID:  parent.ID,
	}
	reports[newReport.ID] = newReport
	reportsMu.Unlock()

	c.JSON(http.StatusCreated, gin.H{"id": newReport.ID, "version": newReport.Version})
}

func listTodos(c *gin.Context) {
	todosMu.RLock()
	list := make([]*models.Todo, 0, len(todos))
	for _, t := range todos {
		list = append(list, t)
	}
	todosMu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	c.JSON(http.StatusOK, list)
}

func completeTodo(c *gin.Context) {
	id := c.Param("id")
	todosMu.Lock()
	t, ok := todos[id]
	if !ok {
		todosMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	}
	t.Completed = true
	todosMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"completed": true})
}

type GenericResource struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	CreateAt string      `json:"created_at"`
}

func listGenericResources(c *gin.Context) {
	role := getRole(c)
	unitID := getUnitID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	resourceType := c.DefaultQuery("type", "samples")

	var all []GenericResource

	if resourceType == "samples" {
		samplesMu.RLock()
		for _, s := range samples {
			if canAccessSample(s, role, unitID) {
				all = append(all, GenericResource{
					ID:       s.ID,
					Type:     "sample",
					Data:     applyPrivacy(s, role),
					CreateAt: s.CreatedAt.Format(time.RFC3339),
				})
			}
		}
		samplesMu.RUnlock()
	} else if resourceType == "reports" {
		samplesMu.RLock()
		sampleMap := make(map[string]*models.Sample)
		for id, s := range samples {
			sampleMap[id] = s
		}
		samplesMu.RUnlock()

		reportsMu.RLock()
		for _, r := range reports {
			if s, ok := sampleMap[r.SampleID]; ok && canAccessSample(s, role, unitID) {
				all = append(all, GenericResource{
					ID:       r.ID,
					Type:     "report",
					Data:     applyReportPrivacy(r, role),
					CreateAt: r.CreatedAt.Format(time.RFC3339),
				})
			}
		}
		reportsMu.RUnlock()
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreateAt > all[j].CreateAt
	})

	total := len(all)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  size,
		"items": all[start:end],
	})
}

func getGenericResource(c *gin.Context) {
	id := c.Param("id")
	role := getRole(c)
	unitID := getUnitID(c)

	samplesMu.RLock()
	if s, ok := samples[id]; ok {
		samplesMu.RUnlock()
		if !canAccessSample(s, role, unitID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusOK, GenericResource{
			ID:       s.ID,
			Type:     "sample",
			Data:     applyPrivacy(s, role),
			CreateAt: s.CreatedAt.Format(time.RFC3339),
		})
		return
	}
	samplesMu.RUnlock()

	reportsMu.RLock()
	if r, ok := reports[id]; ok {
		reportsMu.RUnlock()
		samplesMu.RLock()
		s, ok := samples[r.SampleID]
		samplesMu.RUnlock()
		if !ok || !canAccessSample(s, role, unitID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusOK, GenericResource{
			ID:       r.ID,
			Type:     "report",
			Data:     applyReportPrivacy(r, role),
			CreateAt: r.CreatedAt.Format(time.RFC3339),
		})
		return
	}
	reportsMu.RUnlock()

	c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
}

func getGenericSubResources(c *gin.Context) {
	getSampleSubResources(c)
}
