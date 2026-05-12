package handler

import (
	"encoding/csv"
	"medical-exam-system/internal/model"
	"medical-exam-system/internal/service"
	"medical-exam-system/pkg/dto"
	"medical-exam-system/pkg/errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) errorResponse(c *gin.Context, err error) {
	if ec, ok := err.(errors.ErrorCode); ok {
		c.JSON(ec.GetStatus(), dto.ErrorResponse{
			Code:    ec.GetCode(),
			Message: ec.Error(),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Code:    50000,
		Message: err.Error(),
	})
}

func (h *Handler) GetPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	pkgs, total := h.svc.GetPackages(page, size)

	resp := make([]dto.PackageResponse, 0, len(pkgs))
	for _, pkg := range pkgs {
		items := h.svc.GetPackageItems(pkg)
		itemResp := make([]dto.ExamItemResponse, 0, len(items))
		for _, item := range items {
			itemResp = append(itemResp, dto.ExamItemResponse{
				ID:         item.ID,
				Name:       item.Name,
				Department: item.Department,
				RefRange:   formatRefRange(item.RefRange),
				Unit:       item.Unit,
				Price:      item.Price,
			})
		}
		resp = append(resp, dto.PackageResponse{
			ID:          pkg.ID,
			Name:        pkg.Name,
			Price:       pkg.Price,
			Description: pkg.Description,
			Items:       itemResp,
			CreatedAt:   pkg.CreatedAt,
			UpdatedAt:   pkg.UpdatedAt,
		})
	}

	totalPages := (total + int64(size) - 1) / int64(size)
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetPackageByID(c *gin.Context) {
	id := c.Param("id")
	pkg, err := h.svc.GetPackageByID(id)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	items := h.svc.GetPackageItems(pkg)
	itemResp := make([]dto.ExamItemResponse, 0, len(items))
	for _, item := range items {
		itemResp = append(itemResp, dto.ExamItemResponse{
			ID:         item.ID,
			Name:       item.Name,
			Department: item.Department,
			RefRange:   formatRefRange(item.RefRange),
			Unit:       item.Unit,
			Price:      item.Price,
		})
	}

	c.JSON(http.StatusOK, dto.PackageResponse{
		ID:          pkg.ID,
		Name:        pkg.Name,
		Price:       pkg.Price,
		Description: pkg.Description,
		Items:       itemResp,
		CreatedAt:   pkg.CreatedAt,
		UpdatedAt:   pkg.UpdatedAt,
	})
}

func (h *Handler) CreatePackage(c *gin.Context) {
	var req dto.CreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "请求参数错误"})
		return
	}

	pkg := &model.Package{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		ItemIDs:     req.ItemIDs,
	}

	if err := h.svc.CreatePackage(pkg); err != nil {
		h.errorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": pkg.ID})
}

func (h *Handler) UpdatePackage(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "请求参数错误"})
		return
	}

	pkg := &model.Package{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		ItemIDs:     req.ItemIDs,
	}

	if err := h.svc.UpdatePackage(pkg); err != nil {
		h.errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (h *Handler) DeletePackage(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeletePackage(id); err != nil {
		h.errorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) CalculatePrice(c *gin.Context) {
	var req dto.PriceCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "请求参数错误"})
		return
	}

	packagePrice, addItemsPrice, addItemsDiscount, addItemsFinalPrice, totalPrice, err := h.svc.CalculatePrice(req.PackageID, req.AddItemIDs)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.PriceCalculationResponse{
		PackagePrice:       packagePrice,
		AddItemsPrice:      addItemsPrice,
		AddItemsDiscount:   addItemsDiscount,
		AddItemsFinalPrice: addItemsFinalPrice,
		TotalPrice:         totalPrice,
	})
}

func (h *Handler) GetAvailability(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("20060102")
	}

	morningUsed, afternoonUsed := h.svc.GetDayAvailability(date)

	morningTotal, afternoonTotal := 90, 60

	c.JSON(http.StatusOK, dto.DayAvailability{
		Date: date,
		Morning: dto.Availability{
			Total:     morningTotal,
			Used:      morningUsed,
			Available: morningTotal - morningUsed,
			IsFull:    morningUsed >= morningTotal,
		},
		Afternoon: dto.Availability{
			Total:     afternoonTotal,
			Used:      afternoonUsed,
			Available: afternoonTotal - afternoonUsed,
			IsFull:    afternoonUsed >= afternoonTotal,
		},
	})
}

func (h *Handler) CreateAppointment(c *gin.Context) {
	var req dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "请求参数错误: " + err.Error()})
		return
	}

	examDate, err := time.Parse("2006-01-02", req.ExamDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "日期格式错误，请使用 YYYY-MM-DD"})
		return
	}

	appt, err := h.svc.CreateAppointment(
		req.CustomerName,
		req.IDCard,
		req.Phone,
		req.Gender,
		req.Age,
		req.PackageID,
		req.AddItemIDs,
		examDate,
		req.TimeSlot,
	)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	pkg, _ := h.svc.GetPackageByID(appt.PackageID)
	pkgName := ""
	if pkg != nil {
		pkgName = pkg.Name
	}

	c.JSON(http.StatusCreated, dto.AppointmentResponse{
		ID:           appt.ID,
		ExamNumber:   appt.ExamNumber,
		CustomerName: appt.CustomerName,
		IDCard:       appt.IDCard,
		Phone:        appt.Phone,
		Gender:       appt.Gender,
		Age:          appt.Age,
		PackageID:    appt.PackageID,
		PackageName:  pkgName,
		AddItemIDs:   appt.AddItemIDs,
		TotalPrice:   appt.TotalPrice,
		ExamDate:     appt.ExamDate.Format("2006-01-02"),
		TimeSlot:     appt.TimeSlot,
		Status:       appt.Status,
		CreatedAt:    appt.CreatedAt,
	})
}

func (h *Handler) GetAppointments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	appts, total := h.svc.GetAppointments(page, size)

	resp := make([]dto.AppointmentResponse, 0, len(appts))
	for _, appt := range appts {
		pkg, _ := h.svc.GetPackageByID(appt.PackageID)
		pkgName := ""
		if pkg != nil {
			pkgName = pkg.Name
		}
		resp = append(resp, dto.AppointmentResponse{
			ID:           appt.ID,
			ExamNumber:   appt.ExamNumber,
			CustomerName: appt.CustomerName,
			IDCard:       appt.IDCard,
			Phone:        appt.Phone,
			Gender:       appt.Gender,
			Age:          appt.Age,
			PackageID:    appt.PackageID,
			PackageName:  pkgName,
			AddItemIDs:   appt.AddItemIDs,
			TotalPrice:   appt.TotalPrice,
			ExamDate:     appt.ExamDate.Format("2006-01-02"),
			TimeSlot:     appt.TimeSlot,
			Status:       appt.Status,
			CreatedAt:    appt.CreatedAt,
		})
	}

	totalPages := (total + int64(size) - 1) / int64(size)
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetAppointmentByID(c *gin.Context) {
	id := c.Param("id")
	appt, err := h.svc.GetAppointmentByID(id)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	pkg, _ := h.svc.GetPackageByID(appt.PackageID)
	pkgName := ""
	if pkg != nil {
		pkgName = pkg.Name
	}

	c.JSON(http.StatusOK, dto.AppointmentResponse{
		ID:           appt.ID,
		ExamNumber:   appt.ExamNumber,
		CustomerName: appt.CustomerName,
		IDCard:       appt.IDCard,
		Phone:        appt.Phone,
		Gender:       appt.Gender,
		Age:          appt.Age,
		PackageID:    appt.PackageID,
		PackageName:  pkgName,
		AddItemIDs:   appt.AddItemIDs,
		TotalPrice:   appt.TotalPrice,
		ExamDate:     appt.ExamDate.Format("2006-01-02"),
		TimeSlot:     appt.TimeSlot,
		Status:       appt.Status,
		CreatedAt:    appt.CreatedAt,
	})
}

func (h *Handler) CancelAppointment(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CancelAppointment(id); err != nil {
		h.errorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "取消成功"})
}

func (h *Handler) GetExamResultsByAppointment(c *gin.Context) {
	appointmentID := c.Param("id")

	results := h.svc.GetExamResultsByAppointment(appointmentID)
	items := h.svc.GetAllItems()

	itemMap := make(map[string]*model.ExamItem)
	for _, it := range items {
		itemMap[it.ID] = it
	}

	resp := make([]dto.ExamResultResponse, 0, len(results))
	for _, result := range results {
		var itemName string
		var dept model.Department
		var unit string
		var refRange string
		if it, ok := itemMap[result.ItemID]; ok {
			itemName = it.Name
			dept = it.Department
			unit = it.Unit
			refRange = formatRefRange(it.RefRange)
		}
		resp = append(resp, dto.ExamResultResponse{
			ID:            result.ID,
			AppointmentID: result.AppointmentID,
			ItemID:        result.ItemID,
			ItemName:      itemName,
			Department:    dept,
			ResultValue:   result.ResultValue,
			Unit:          unit,
			RefRange:      refRange,
			IsAbnormal:    result.IsAbnormal,
			AbnormalType:  result.AbnormalType,
			IsCritical:    result.IsCritical,
			Status:        result.Status,
			ModifyCount:   result.ModifyCount,
			CreatedAt:     result.CreatedAt,
			UpdatedAt:     result.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SubmitExamResult(c *gin.Context) {
	resultID := c.Param("id")
	var req dto.SubmitExamResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "请求参数错误"})
		return
	}

	result, err := h.svc.SubmitExamResult(resultID, req.ResultValue)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	items := h.svc.GetAllItems()
	var itemName string
	var dept model.Department
	var unit string
	var refRange string
	for _, it := range items {
		if it.ID == result.ItemID {
			itemName = it.Name
			dept = it.Department
			unit = it.Unit
			refRange = formatRefRange(it.RefRange)
			break
		}
	}

	c.JSON(http.StatusOK, dto.ExamResultResponse{
		ID:            result.ID,
		AppointmentID: result.AppointmentID,
		ItemID:        result.ItemID,
		ItemName:      itemName,
		Department:    dept,
		ResultValue:   result.ResultValue,
		Unit:          unit,
		RefRange:      refRange,
		IsAbnormal:    result.IsAbnormal,
		AbnormalType:  result.AbnormalType,
		IsCritical:    result.IsCritical,
		Status:        result.Status,
		ModifyCount:   result.ModifyCount,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.UpdatedAt,
	})
}

func (h *Handler) GetCriticalAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	alerts := h.svc.GetRecentCriticalAlerts(limit)

	resp := make([]dto.CriticalAlertResponse, 0, len(alerts))
	for _, alert := range alerts {
		resp = append(resp, dto.CriticalAlertResponse{
			ID:           alert.ID,
			ExamNumber:   alert.ExamNumber,
			CustomerName: alert.CustomerName,
			ItemName:     alert.ItemName,
			ResultValue:  alert.ResultValue,
			RefRange:     alert.RefRange,
			CreatedAt:    alert.CreatedAt,
			Resolved:     alert.Resolved,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	reports, total := h.svc.GetReports(page, size)

	resp := make([]dto.ReportResponse, 0, len(reports))
	for _, report := range reports {
		itemResp := make([]dto.ReportItemResponse, 0, len(report.Items))
		for _, item := range report.Items {
			itemResp = append(itemResp, dto.ReportItemResponse{
				ItemID:       item.ItemID,
				ItemName:     item.ItemName,
				Department:   item.Department,
				ResultValue:  item.ResultValue,
				Unit:         item.Unit,
				RefRange:     item.RefRange,
				IsAbnormal:   item.IsAbnormal,
				AbnormalType: item.AbnormalType,
				IsCritical:   item.IsCritical,
			})
		}
		resp = append(resp, dto.ReportResponse{
			ID:             report.ID,
			ExamNumber:     report.ExamNumber,
			CustomerName:   report.CustomerName,
			Gender:         report.Gender,
			Age:            report.Age,
			PackageName:    report.PackageName,
			Items:          itemResp,
			HasAbnormal:    report.HasAbnormal,
			GeneralAdvice:  report.GeneralAdvice,
			FollowUpAdvice: report.FollowUpAdvice,
			Status:         report.Status,
			CreatedAt:      report.CreatedAt,
			UpdatedAt:      report.UpdatedAt,
			PublishedAt:    report.PublishedAt,
		})
	}

	totalPages := (total + int64(size) - 1) / int64(size)
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetReportByID(c *gin.Context) {
	id := c.Param("id")
	report, err := h.svc.GetReportByID(id)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	itemResp := make([]dto.ReportItemResponse, 0, len(report.Items))
	for _, item := range report.Items {
		itemResp = append(itemResp, dto.ReportItemResponse{
			ItemID:       item.ItemID,
			ItemName:     item.ItemName,
			Department:   item.Department,
			ResultValue:  item.ResultValue,
			Unit:         item.Unit,
			RefRange:     item.RefRange,
			IsAbnormal:   item.IsAbnormal,
			AbnormalType: item.AbnormalType,
			IsCritical:   item.IsCritical,
		})
	}

	c.JSON(http.StatusOK, dto.ReportResponse{
		ID:             report.ID,
		ExamNumber:     report.ExamNumber,
		CustomerName:   report.CustomerName,
		Gender:         report.Gender,
		Age:            report.Age,
		PackageName:    report.PackageName,
		Items:          itemResp,
		HasAbnormal:    report.HasAbnormal,
		GeneralAdvice:  report.GeneralAdvice,
		FollowUpAdvice: report.FollowUpAdvice,
		Status:         report.Status,
		CreatedAt:      report.CreatedAt,
		UpdatedAt:      report.UpdatedAt,
		PublishedAt:    report.PublishedAt,
	})
}

func (h *Handler) UpdateReport(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "请求参数错误"})
		return
	}

	report, err := h.svc.UpdateReportContent(id, req.GeneralAdvice, req.FollowUpAdvice)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "status": report.Status})
}

func (h *Handler) SubmitReportForReview(c *gin.Context) {
	id := c.Param("id")
	report, err := h.svc.SubmitReportForReview(id)
	if err != nil {
		h.errorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已提交审核", "status": report.Status})
}

func (h *Handler) PublishReport(c *gin.Context) {
	id := c.Param("id")
	report, err := h.svc.PublishReport(id)
	if err != nil {
		h.errorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已发布", "status": report.Status})
}

func (h *Handler) ExportRecords(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 40000, Message: "缺少日期参数"})
		return
	}

	reports, err := h.svc.GetReportsByDateRange(startDate, endDate)
	if err != nil {
		h.errorResponse(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="medical_records_`+startDate+`_`+endDate+`.csv"`)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{"体检编号", "姓名", "性别", "年龄", "套餐名称", "是否有异常", "总检建议", "复查建议", "检查结果详情"}
	writer.Write(header)

	for _, report := range reports {
		var itemDetails []string
		for _, item := range report.Items {
			resultVal := ""
			if item.ResultValue != nil {
				resultVal = strconv.FormatFloat(*item.ResultValue, 'f', -1, 64)
			}
			abnormal := "否"
			if item.IsAbnormal {
				abnormal = "是(" + item.AbnormalType + ")"
			}
			critical := ""
			if item.IsCritical {
				critical = "[危急]"
			}
			itemDetails = append(itemDetails, item.ItemName+": "+resultVal+item.Unit+"("+item.RefRange+")"+abnormal+critical)
		}

		hasAbnormal := "否"
		if report.HasAbnormal {
			hasAbnormal = "是"
		}

		row := []string{
			report.ExamNumber,
			report.CustomerName,
			report.Gender,
			strconv.Itoa(report.Age),
			report.PackageName,
			hasAbnormal,
			report.GeneralAdvice,
			report.FollowUpAdvice,
			strings.Join(itemDetails, "; "),
		}
		writer.Write(row)
	}
}

func (h *Handler) GetItems(c *gin.Context) {
	items := h.svc.GetAllItems()
	resp := make([]dto.ExamItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.ExamItemResponse{
			ID:         item.ID,
			Name:       item.Name,
			Department: item.Department,
			RefRange:   formatRefRange(item.RefRange),
			Unit:       item.Unit,
			Price:      item.Price,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func formatRefRange(r model.RefRange) string {
	if r.Min != nil && r.Max != nil {
		return strconv.FormatFloat(*r.Min, 'f', -1, 64) + "-" + strconv.FormatFloat(*r.Max, 'f', -1, 64)
	} else if r.Min != nil {
		return ">" + strconv.FormatFloat(*r.Min, 'f', -1, 64)
	} else if r.Max != nil {
		return "<" + strconv.FormatFloat(*r.Max, 'f', -1, 64)
	}
	return ""
}
