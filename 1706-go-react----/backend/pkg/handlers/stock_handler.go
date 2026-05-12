package handlers

import (
	"hospital-pharmacy/pkg/repositories"
	"hospital-pharmacy/pkg/services"
	"hospital-pharmacy/pkg/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StockInRequest struct {
	DrugID         uint   `json:"drug_id"`
	Quantity       int    `json:"quantity"`
	ProductionDate string `json:"production_date"`
	ExpiryDate     string `json:"expiry_date"`
	Supplier       string `json:"supplier"`
	SupplierCode   string `json:"supplier_code"`
	Operator1      string `json:"operator1"`
	Operator2      string `json:"operator2"`
}

type StockOutRequest struct {
	DrugID    uint   `json:"drug_id"`
	Quantity  int    `json:"quantity"`
	Operator1 string `json:"operator1"`
	Operator2 string `json:"operator2"`
	Remark    string `json:"remark"`
}

func StockIn(c *gin.Context) {
	var req StockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	prodDate, err := time.Parse("2006-01-02", req.ProductionDate)
	if err != nil {
		utils.BadRequestResponse(c, "生产日期格式错误")
		return
	}

	expDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		utils.BadRequestResponse(c, "有效期格式错误")
		return
	}

	err = services.StockIn(services.StockInRequest{
		DrugID:         req.DrugID,
		Quantity:       req.Quantity,
		ProductionDate: prodDate,
		ExpiryDate:     expDate,
		Supplier:       req.Supplier,
		SupplierCode:   req.SupplierCode,
		Operator1:      req.Operator1,
		Operator2:      req.Operator2,
	})

	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func StockOut(c *gin.Context) {
	var req StockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	err := services.StockOut(services.StockOutRequest{
		DrugID:    req.DrugID,
		Quantity:  req.Quantity,
		Operator1: req.Operator1,
		Operator2: req.Operator2,
		Remark:    req.Remark,
	})

	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func GetStockItems(c *gin.Context) {
	drugIDStr := c.Param("drugId")
	drugID, err := strconv.ParseUint(drugIDStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "药品ID参数错误")
		return
	}

	items, err := services.GetStockItems(uint(drugID))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, items)
}

func GetAlerts(c *gin.Context) {
	alerts, err := services.GetAlerts()
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, alerts)
}

func MarkAlertRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	if err := services.MarkAlertRead(uint(id)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func GetNearExpiry(c *gin.Context) {
	items, err := services.GetNearExpiryItems()
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, items)
}

func GetStockValue(c *gin.Context) {
	costValue, retailValue, err := services.GetStockValue()
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"cost_value":   costValue,
		"retail_value": retailValue,
	})
}

func GetTransactions(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			utils.BadRequestResponse(c, "开始日期格式错误")
			return
		}
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			utils.BadRequestResponse(c, "结束日期格式错误")
			return
		}
	}

	transactions, err := repositories.ListTransactions(start, end)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, transactions)
}

func SetupStockRoutes(r *gin.Engine) {
	stockGroup := r.Group("/api/stock")
	{
		stockGroup.POST("/in", StockIn)
		stockGroup.POST("/out", StockOut)
		stockGroup.GET("/items/:drugId", GetStockItems)
		stockGroup.GET("/alerts", GetAlerts)
		stockGroup.POST("/alerts/:id/read", MarkAlertRead)
		stockGroup.GET("/near-expiry", GetNearExpiry)
		stockGroup.GET("/value", GetStockValue)
		stockGroup.GET("/transactions", GetTransactions)
	}
}
