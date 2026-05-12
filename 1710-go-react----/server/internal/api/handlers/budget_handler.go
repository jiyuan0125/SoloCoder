package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"trial-management-system/internal/models"
	"trial-management-system/internal/services"
)

func CreateDepartment(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept := &models.Department{Name: req.Name, Code: req.Code}
	if err := services.CreateDepartment(dept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dept)
}

func ListDepartments(c *gin.Context) {
	depts, err := services.ListDepartments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, depts)
}

func CreateBudget(c *gin.Context) {
	var req struct {
		DepartmentID string  `json:"department_id" binding:"required"`
		Period       string  `json:"period" binding:"required"`
		TotalAmount  float64 `json:"total_amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deptID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	budget := &models.Budget{
		DepartmentID: deptID,
		Period:       req.Period,
		TotalAmount:  req.TotalAmount,
	}

	if err := services.CreateBudget(budget); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, budget)
}

func ListBudgets(c *gin.Context) {
	budgets, err := services.ListBudgets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, budgets)
}

func CreateBudgetItem(c *gin.Context) {
	var req struct {
		BudgetID string  `json:"budget_id" binding:"required"`
		Name     string  `json:"name" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budgetID, err := uuid.Parse(req.BudgetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预算ID"})
		return
	}

	budget, err := services.GetBudgetByID(budgetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预算不存在"})
		return
	}

	var allocated float64
	for _, item := range budget.Items {
		allocated += item.Amount
	}

	if allocated+req.Amount > budget.TotalAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "超出预算总额"})
		return
	}

	item := &models.BudgetItem{
		BudgetID: budgetID,
		Name:     req.Name,
		Amount:   req.Amount,
		Status:   models.ItemPending,
	}

	if err := services.CreateBudgetItem(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func ApproveBudgetItem(c *gin.Context) {
	id := c.Param("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	item, err := services.GetBudgetItemByID(itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预算项不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	budget, err := services.GetBudgetByID(item.BudgetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if budget.UsedAmount+item.Amount > budget.TotalAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "预算不足"})
		return
	}

	newUsed := budget.UsedAmount + item.Amount
	newMonthlyUsed := budget.MonthlyUsed + item.Amount

	threshold := budget.TotalAmount * 0.8

	if err := services.UpdateBudgetItem(itemID, map[string]interface{}{
		"status":      models.ItemApproved,
		"used_amount": item.Amount,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateBudget(budget.ID, map[string]interface{}{
		"used_amount":   newUsed,
		"monthly_used":  newMonthlyUsed,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if newMonthlyUsed > threshold {
		alert := &models.BudgetAlert{
			BudgetID:     budget.ID,
			DepartmentID: budget.DepartmentID,
			Message:      "本月支出已超过预算的80%",
			AlertType:    "warning",
		}
		services.CreateBudgetAlert(alert)
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批通过"})
}

func AdjustBudgetTotal(c *gin.Context) {
	id := c.Param("id")
	budgetID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预算ID"})
		return
	}

	var req struct {
		NewTotal float64 `json:"new_total" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := services.GetBudgetByID(budgetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预算不存在"})
		return
	}

	var completedTotal float64
	var pendingTotal float64
	for _, item := range budget.Items {
		if item.Status == models.ItemCompleted {
			completedTotal += item.Amount
		} else {
			pendingTotal += item.Amount
		}
	}

	if req.NewTotal < completedTotal {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新总额不能小于已完成子项的总额"})
		return
	}

	remaining := req.NewTotal - completedTotal

	if pendingTotal > 0 {
		for _, item := range budget.Items {
			if item.Status != models.ItemCompleted {
				ratio := item.Amount / pendingTotal
				newAmount := remaining * ratio
				services.UpdateBudgetItem(item.ID, map[string]interface{}{"amount": newAmount})
			}
		}
	}

	services.UpdateBudget(budgetID, map[string]interface{}{"total_amount": req.NewTotal})

	c.JSON(http.StatusOK, gin.H{"message": "预算调整完成"})
}

func ListBudgetAlerts(c *gin.Context) {
	alerts, err := services.ListBudgetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func MarkBudgetItemCompleted(c *gin.Context) {
	id := c.Param("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := services.UpdateBudgetItem(itemID, map[string]interface{}{"status": models.ItemCompleted}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "预算项已标记为完成"})
}
