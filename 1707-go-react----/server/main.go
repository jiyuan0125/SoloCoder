package main

import (
	"flag"
	"medical-device-manager/config"
	"medical-device-manager/database"
	"medical-device-manager/handlers"
	"os"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

func main() {
	portFlag := flag.String("port", "", "服务端口")
	flag.Parse()

	cfg := config.LoadConfig()

	port := *portFlag
	if port == "" {
		port = cfg.Port
	}

	if err := database.InitDB(cfg.DBPath); err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		devices := api.Group("/devices")
		{
			devices.POST("", handlers.CreateDevice)
			devices.GET("", handlers.GetDevices)
			devices.GET("/export", handlers.ExportDevicesCSV)
			devices.GET("/:id", handlers.GetDevice)
			devices.PUT("/:id/status", handlers.UpdateDeviceStatus)
		}

		agencies := api.Group("/calibration-agencies")
		{
			agencies.POST("", handlers.CreateCalibrationAgency)
			agencies.GET("", handlers.GetCalibrationAgencies)
		}

		calibration := api.Group("/calibration-records")
		{
			calibration.POST("", handlers.CreateCalibrationRecord)
			calibration.GET("", handlers.GetCalibrationRecords)
			calibration.GET("/reminders", handlers.GetCalibrationReminders)
			calibration.PUT("/:id/status", handlers.UpdateCalibrationStatus)
		}

		maintenance := api.Group("/maintenance-plans")
		{
			maintenance.POST("", handlers.CreateMaintenancePlan)
			maintenance.GET("", handlers.GetMaintenancePlans)
			maintenance.POST("/generate", handlers.GenerateMaintenancePlans)
			maintenance.PUT("/:id/complete", handlers.CompleteMaintenancePlan)
			maintenance.PUT("/:id/cancel", handlers.CancelMaintenancePlan)
		}

		workorders := api.Group("/workorders")
		{
			workorders.POST("", handlers.CreateWorkOrder)
			workorders.GET("", handlers.GetWorkOrders)
			workorders.GET("/:id", handlers.GetWorkOrder)
			workorders.PUT("/:id/status", handlers.UpdateWorkOrderStatus)
			workorders.POST("/escalate", handlers.EscalateWorkOrders)
			workorders.POST("/auto-close", handlers.AutoCloseWorkOrders)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	addr := ":" + port
	println("Server starting on " + addr)
	if err := r.Run(addr); err != nil {
		println("Failed to start server:", err.Error())
		os.Exit(1)
	}
}
