package main

import (
	"flag"
	"os"

	"diploma-auth-system/config"
	"diploma-auth-system/handlers"
	"diploma-auth-system/middleware"
	"diploma-auth-system/models"
	"diploma-auth-system/store"

	"github.com/gin-gonic/gin"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务端口")
	flag.Parse()
	
	if port == "" {
		port = os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
	}
	
	cfg := config.NewConfig(port)
	dataStore := store.NewStore()
	handler := handlers.NewHandler(dataStore)
	authMiddleware := middleware.NewAuthMiddleware(dataStore)
	
	r := gin.Default()
	
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	
	r.POST("/api/login", handler.Login)
	
	api := r.Group("/api")
	api.Use(authMiddleware.AuthRequired())
	
	api.GET("/me", handler.GetCurrentUser)
	
	diplomas := api.Group("/diplomas")
	diplomas.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		diplomas.GET("", handler.ListDiplomas)
		diplomas.GET("/:id", handler.GetDiploma)
		diplomas.POST("", handler.CreateDiploma)
		diplomas.PUT("/:id", handler.UpdateDiploma)
		diplomas.DELETE("/:id", handler.DeleteDiploma)
	}
	
	verify := api.Group("/verify")
	verify.Use(authMiddleware.RequireRole(models.RoleAdmin, models.RoleVerifier))
	{
		verify.POST("", handler.VerifyDiploma)
	}
	
	logs := api.Group("/logs")
	logs.Use(authMiddleware.RequireRole(models.RoleAdmin, models.RoleVerifier, models.RoleViewer))
	{
		logs.GET("", handler.ListLogs)
	}
	
	users := api.Group("/users")
	users.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		users.GET("", handler.ListUsers)
		users.POST("", handler.CreateUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}
	
	r.Run(":" + cfg.Port)
}
