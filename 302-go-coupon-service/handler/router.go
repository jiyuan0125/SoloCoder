package handler

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	batchHandler := NewBatchHandler()
	couponHandler := NewCouponHandler()

	api := r.Group("/api/v1")
	{
		batches := api.Group("/batches")
		{
			batches.POST("", batchHandler.CreateBatch)
			batches.GET("", batchHandler.ListBatches)
			batches.GET("/:id", batchHandler.GetBatchByID)
			batches.PUT("/:id", batchHandler.UpdateBatch)
			batches.GET("/:id/progress", batchHandler.GetBatchProgress)
		}

		coupons := api.Group("/coupons")
		{
			coupons.POST("/claim", couponHandler.ClaimCoupon)
			coupons.POST("/redeem", couponHandler.RedeemCoupon)
			coupons.POST("/return", couponHandler.ReturnCoupons)
			coupons.GET("/code/:code", couponHandler.GetCouponByCode)
			coupons.GET("/user/:user_id", couponHandler.GetUserAvailableCoupons)
		}
	}

	return r
}
