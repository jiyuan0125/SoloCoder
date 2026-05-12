package routes

import (
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"

	"trial-management-system/internal/api/handlers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.Default())

	api := r.Group("/api")
	{
		protocols := api.Group("/protocols")
		{
			protocols.POST("", handlers.CreateProtocol)
			protocols.GET("", handlers.ListProtocols)
			protocols.GET("/:id", handlers.GetProtocol)
			protocols.PUT("/:id", handlers.UpdateProtocol)
			protocols.POST("/:id/sites", handlers.AddSiteToProtocol)
			protocols.POST("/:id/visits", handlers.AddVisitToProtocol)
		}

		subjects := api.Group("/subjects")
		{
			subjects.POST("", handlers.CreateSubject)
			subjects.GET("", handlers.ListSubjects)
			subjects.GET("/:id", handlers.GetSubject)
			subjects.PUT("/:id/status", handlers.UpdateSubjectStatus)
			subjects.POST("/:id/withdraw", handlers.WithdrawSubject)
		}

		visits := api.Group("/visits")
		{
			visits.POST("/records", handlers.RecordVisit)
			visits.GET("/records", handlers.ListVisitRecords)
			visits.GET("/records/:id", handlers.GetVisitRecord)
		}

		aes := api.Group("/aes")
		{
			aes.POST("", handlers.CreateAdverseEvent)
			aes.GET("", handlers.ListAdverseEvents)
			aes.GET("/:id", handlers.GetAdverseEvent)
			aes.PUT("/:id", handlers.UpdateAdverseEvent)
			aes.POST("/:id/submit-report", handlers.SubmitSAEReport)
		}

		todos := api.Group("/todos")
		{
			todos.GET("", handlers.ListTodos)
			todos.PUT("/:id/status", handlers.UpdateTodoStatus)
		}

		export := api.Group("/export")
		{
			export.GET("/data", handlers.ExportData)
			export.GET("/visits", handlers.ExportVisitData)
		}

		departments := api.Group("/departments")
		{
			departments.POST("", handlers.CreateDepartment)
			departments.GET("", handlers.ListDepartments)
		}

		budgets := api.Group("/budgets")
		{
			budgets.POST("", handlers.CreateBudget)
			budgets.GET("", handlers.ListBudgets)
			budgets.POST("/items", handlers.CreateBudgetItem)
			budgets.POST("/items/:id/approve", handlers.ApproveBudgetItem)
			budgets.PUT("/items/:id/complete", handlers.MarkBudgetItemCompleted)
			budgets.PUT("/:id/adjust", handlers.AdjustBudgetTotal)
		}

		alerts := api.Group("/alerts")
		{
			alerts.GET("", handlers.ListBudgetAlerts)
		}

		systemA := api.Group("/system-a")
		{
			systemA.POST("", handlers.CreateSystemA)
			systemA.GET("", handlers.ListSystemA)
			systemA.GET("/:id", handlers.GetSystemA)
		}

		systemB := api.Group("/system-b")
		{
			systemB.POST("", handlers.CreateSystemB)
			systemB.GET("", handlers.ListSystemB)
			systemB.GET("/:id", handlers.GetSystemB)
		}
	}

	return r
}
