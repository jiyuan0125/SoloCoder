package router

import (
	"ambulance-scheduler/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handler.Handler) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", h.Health)

	api := r.Group("/api")
	{
		calls := api.Group("/calls")
		{
			calls.POST("", h.ReceiveCall)
			calls.GET("", h.ListCalls)
			calls.GET("/pending", h.GetPendingDispatchCalls)
			calls.POST("/:id/accept", h.AcceptCall)
			calls.POST("/:id/process", h.ProcessCall)
			calls.POST("/:id/submit-review", h.SubmitForReview)
			calls.POST("/:id/complete", h.CompleteCall)
			calls.POST("/:id/reject", h.RejectCall)
		}

		vehicles := api.Group("/vehicles")
		{
			vehicles.POST("", h.CreateVehicle)
			vehicles.GET("", h.ListVehicles)
			vehicles.PUT("/:id/status", h.UpdateVehicleStatus)
			vehicles.POST("/:id/maintenance", h.SetVehicleMaintenance)
		}

		dispatch := api.Group("/dispatch")
		{
			dispatch.POST("", h.DispatchVehicle)
			dispatch.GET("/recommend/:callId", h.RecommendVehicle)
			dispatch.GET("/records", h.ListDispatchRecords)
		}

		triage := api.Group("/triage")
		{
			triage.POST("", h.CreateTriageRecord)
			triage.GET("/records", h.ListTriageRecords)
		}

		api.GET("/stats", h.GetStatistics)
	}

	return r
}
