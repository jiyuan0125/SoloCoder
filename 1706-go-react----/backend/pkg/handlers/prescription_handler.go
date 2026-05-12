package handlers

import (
	"hospital-pharmacy/pkg/services"
	"hospital-pharmacy/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApproveRequest struct {
	Operator1 string `json:"operator1"`
	Operator2 string `json:"operator2"`
}

type RejectRequest struct {
	Comment string `json:"comment"`
}

func CreatePrescription(c *gin.Context) {
	var req services.CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	prescription, err := services.CreatePrescription(req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, prescription)
}

func GetPrescription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	prescription, err := services.GetPrescription(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "处方不存在")
		return
	}

	utils.SuccessResponse(c, prescription)
}

func ListPrescriptions(c *gin.Context) {
	status := c.Query("status")
	prescriptions, err := services.ListPrescriptions(status)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, prescriptions)
}

func ApprovePrescription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	if err := services.ApprovePrescription(uint(id), req.Operator1, req.Operator2); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func RejectPrescription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	if err := services.RejectPrescription(uint(id), req.Comment); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func CancelPrescription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	if err := services.CancelPrescription(uint(id)); err != nil {
		if err.Error() == "已审核通过的处方不能取消" {
			utils.ConflictResponse(c, err.Error())
			return
		}
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func SetupPrescriptionRoutes(r *gin.Engine) {
	presGroup := r.Group("/api/prescriptions")
	{
		presGroup.GET("", ListPrescriptions)
		presGroup.GET("/:id", GetPrescription)
		presGroup.POST("", CreatePrescription)
		presGroup.POST("/:id/approve", ApprovePrescription)
		presGroup.POST("/:id/reject", RejectPrescription)
		presGroup.POST("/:id/cancel", CancelPrescription)
	}
}
