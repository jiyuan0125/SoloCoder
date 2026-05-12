package app

import (
	"medical-quality-system/internal/controller"
	"medical-quality-system/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	InitDB()

	r := gin.Default()

	r.Use(cors.Default())
	r.Use(middleware.ErrorHandler())

	api := r.Group("/api")
	{
		indicatorCtrl := controller.NewIndicatorController()
		pdcaCtrl := controller.NewPDCAController()
		todoCtrl := controller.NewTodoController()
		meetingCtrl := controller.NewMeetingController()
		aggCtrl := controller.NewAggregationController()

		indicators := api.Group("/indicators")
		{
			indicators.POST("", indicatorCtrl.Create)
			indicators.GET("", indicatorCtrl.List)
			indicators.GET("/:id", indicatorCtrl.Get)
			indicators.PUT("/:id", indicatorCtrl.Update)
			indicators.PUT("/:id/target", indicatorCtrl.UpdateTarget)
			indicators.DELETE("/:id", indicatorCtrl.Delete)

			indicators.POST("/:id/data", indicatorCtrl.CreateData)
			indicators.GET("/:id/data", indicatorCtrl.ListData)
			indicators.GET("/:id/trend", indicatorCtrl.GetTrend)
		}

		pdca := api.Group("/pdca")
		{
			pdca.POST("", pdcaCtrl.Create)
			pdca.GET("", pdcaCtrl.List)
			pdca.GET("/:id", pdcaCtrl.Get)
			pdca.PUT("/:id", pdcaCtrl.Update)
			pdca.DELETE("/:id", pdcaCtrl.Delete)

			pdca.POST("/:id/phases", pdcaCtrl.CreatePhase)
			pdca.PUT("/:id/phases/:phase", pdcaCtrl.UpdatePhase)
			pdca.POST("/:id/next-phase", pdcaCtrl.NextPhase)
			pdca.POST("/:id/start-next-cycle", pdcaCtrl.StartNextCycle)
		}

		todos := api.Group("/todos")
		{
			todos.GET("", todoCtrl.List)
			todos.GET("/:id", todoCtrl.Get)
			todos.PUT("/:id", todoCtrl.Update)
			todos.PUT("/:id/status", todoCtrl.UpdateStatus)
			todos.DELETE("/:id", todoCtrl.Delete)
		}

		meetings := api.Group("/meetings")
		{
			meetings.POST("", meetingCtrl.Create)
			meetings.GET("", meetingCtrl.List)
			meetings.GET("/:id", meetingCtrl.Get)
			meetings.PUT("/:id", meetingCtrl.Update)
			meetings.DELETE("/:id", meetingCtrl.Delete)
		}

		aggregations := api.Group("/aggregations")
		{
			aggregations.GET("/department", aggCtrl.ByDepartment)
			aggregations.GET("/category", aggCtrl.ByCategory)
			aggregations.GET("/time", aggCtrl.ByTime)
		}
	}

	return r
}
