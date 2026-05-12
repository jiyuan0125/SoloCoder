package controller

import (
	"net/http"

	"medical-quality-system/internal/service"

	"github.com/gin-gonic/gin"
)

type AggregationController struct {
	service *service.AggregationService
}

func NewAggregationController() *AggregationController {
	return &AggregationController{
		service: service.NewAggregationService(),
	}
}

func (c *AggregationController) ByDepartment(ctx *gin.Context) {
	month := ctx.Query("month")

	aggregations, err := c.service.ByDepartment(month)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, aggregations)
}

func (c *AggregationController) ByCategory(ctx *gin.Context) {
	fromMonth := ctx.Query("from")
	toMonth := ctx.Query("to")

	aggregations, err := c.service.ByCategory(fromMonth, toMonth)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, aggregations)
}

func (c *AggregationController) ByTime(ctx *gin.Context) {
	periodType := ctx.Query("period")
	fromTime := ctx.Query("from")
	toTime := ctx.Query("to")

	aggregations, err := c.service.ByTime(periodType, fromTime, toTime)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, aggregations)
}
