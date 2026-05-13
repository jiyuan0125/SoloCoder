package main

import (
	"log"

	"json-validator/config"
	"json-validator/handlers"
	"json-validator/middleware"
	"json-validator/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := storage.InitDB(); err != nil {
		log.Printf("警告: 数据库初始化失败: %v", err)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(middleware.RecoveryMiddleware())

	r.POST("/validate", handlers.ValidateHandler)
	r.GET("/validate/:id", handlers.GetValidationHandler)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("JSON 校验服务启动在端口 %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
