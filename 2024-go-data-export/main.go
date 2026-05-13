package main

import (
	"data-export/config"
	"data-export/database"
	"data-export/export"
	"data-export/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.Init()
	export.InitManager()

	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/tables", handlers.ListTables)
		api.GET("/tables/:table", handlers.GetTableInfo)
		api.POST("/export", handlers.SubmitExport)
		api.GET("/export/:id/progress", handlers.GetProgress)
		api.GET("/export/:id/download", handlers.DownloadFile)
	}

	r.Run(":" + config.ServerPort)
}
