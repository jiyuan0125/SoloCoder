package handler

import (
	"medical-exam-system/internal/model"
	"medical-exam-system/pkg/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GenericResource struct {
	ID   string      `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (h *Handler) GetGenericList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	pkgs, pkgTotal := h.svc.GetPackages(page, size)
	appts, apptTotal := h.svc.GetAppointments(page, size)
	reports, reportTotal := h.svc.GetReports(page, size)

	resources := make([]GenericResource, 0)

	for _, pkg := range pkgs {
		resources = append(resources, GenericResource{
			ID:   pkg.ID,
			Type: "package",
			Data: pkg,
		})
	}

	for _, appt := range appts {
		resources = append(resources, GenericResource{
			ID:   appt.ID,
			Type: "appointment",
			Data: appt,
		})
	}

	for _, report := range reports {
		resources = append(resources, GenericResource{
			ID:   report.ID,
			Type: "report",
			Data: report,
		})
	}

	total := pkgTotal + apptTotal + reportTotal
	totalPages := (total + int64(size) - 1) / int64(size)

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resources,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetGenericDetail(c *gin.Context) {
	id := c.Param("id")

	if appt, err := h.svc.GetAppointmentByID(id); err == nil {
		c.JSON(http.StatusOK, GenericResource{
			ID:   appt.ID,
			Type: "appointment",
			Data: appt,
		})
		return
	}

	if pkg, err := h.svc.GetPackageByID(id); err == nil {
		c.JSON(http.StatusOK, GenericResource{
			ID:   pkg.ID,
			Type: "package",
			Data: pkg,
		})
		return
	}

	if report, err := h.svc.GetReportByID(id); err == nil {
		c.JSON(http.StatusOK, GenericResource{
			ID:   report.ID,
			Type: "report",
			Data: report,
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
}

func (h *Handler) GetGenericSubResources(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if _, err := h.svc.GetAppointmentByID(id); err == nil {
		results := h.svc.GetExamResultsByAppointment(id)
		subResources := make([]GenericResource, 0, len(results))
		for _, result := range results {
			subResources = append(subResources, GenericResource{
				ID:   result.ID,
				Type: "exam_result",
				Data: result,
			})
		}
		c.JSON(http.StatusOK, dto.PaginatedResponse{
			Data:       subResources,
			Page:       page,
			Size:       size,
			Total:      int64(len(subResources)),
			TotalPages: 1,
		})
		return
	}

	if pkg, err := h.svc.GetPackageByID(id); err == nil {
		items := h.svc.GetPackageItems(pkg)
		subResources := make([]GenericResource, 0, len(items))
		for _, item := range items {
			subResources = append(subResources, GenericResource{
				ID:   item.ID,
				Type: "exam_item",
				Data: item,
			})
		}
		c.JSON(http.StatusOK, dto.PaginatedResponse{
			Data:       subResources,
			Page:       page,
			Size:       size,
			Total:      int64(len(subResources)),
			TotalPages: 1,
		})
		return
	}

	if report, err := h.svc.GetReportByID(id); err == nil {
		subResources := make([]GenericResource, 0, len(report.Items))
		for i, item := range report.Items {
			subResources = append(subResources, GenericResource{
				ID:   report.ID + "_" + strconv.Itoa(i),
				Type: "report_item",
				Data: item,
			})
		}
		c.JSON(http.StatusOK, dto.PaginatedResponse{
			Data:       subResources,
			Page:       page,
			Size:       size,
			Total:      int64(len(subResources)),
			TotalPages: 1,
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
}

var _ model.ExamItem
