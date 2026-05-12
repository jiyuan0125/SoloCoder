package routes

import (
	"research-collaboration/src/handlers"
	"research-collaboration/src/storage"

	"github.com/gin-gonic/gin"
)

func SetupRouter(store *storage.Storage) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	projectHandler := handlers.NewProjectHandler(store)
	taskHandler := handlers.NewTaskHandler(store)
	dataHandler := handlers.NewDataHandler(store)
	achievementHandler := handlers.NewAchievementHandler(store)
	approvalHandler := handlers.NewApprovalHandler(store)

	api := r.Group("/api")
	{
		users := api.Group("/users")
		{
			users.GET("", func(c *gin.Context) {
				c.JSON(200, store.GetUsers())
			})
		}

		departments := api.Group("/departments")
		{
			departments.GET("", func(c *gin.Context) {
				c.JSON(200, store.GetDepartments())
			})
		}

		projects := api.Group("/projects")
		{
			projects.GET("", projectHandler.ListProjects)
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.PUT("/:id/status", projectHandler.UpdateProjectStatus)
			projects.DELETE("/:id", projectHandler.DeleteProject)
		}

		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.GetAllTasks)
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("/project/:projectId", taskHandler.GetTasksByProject)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.PUT("/:id/status", taskHandler.UpdateTaskStatus)
		}

		data := api.Group("/data")
		{
			data.GET("", dataHandler.ListData)
			data.POST("", dataHandler.UploadData)
			data.GET("/:id", dataHandler.GetData)
			data.POST("/:id/download", dataHandler.DownloadData)
			data.GET("/:id/logs", dataHandler.GetDownloadLogs)
		}

		achievements := api.Group("/achievements")
		{
			achievements.GET("", achievementHandler.ListAchievements)
			achievements.POST("", achievementHandler.CreateAchievement)
			achievements.GET("/export", achievementHandler.ExportAchievements)
			achievements.GET("/:id", achievementHandler.GetAchievement)
			achievements.PUT("/:id", achievementHandler.UpdateAchievement)
		}

		approvals := api.Group("/approvals")
		{
			approvals.POST("", approvalHandler.CreateApproval)
			approvals.GET("/:id", approvalHandler.GetApproval)
			approvals.POST("/:id/process", approvalHandler.ProcessApproval)
		}

		todos := api.Group("/todos")
		{
			todos.GET("", func(c *gin.Context) {
				userID := c.GetHeader("X-User-ID")
				if userID == "" {
					userID = "user-1"
				}
				c.JSON(200, store.GetTodosByUser(userID))
			})
		}
	}

	return r
}
