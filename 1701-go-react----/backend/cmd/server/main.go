package main

import (
	"cmms/internal/handlers"
	"cmms/internal/storage"
	"cmms/internal/ws"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := 8300

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	flagPort := flag.Int("port", 0, "服务端口")
	flag.Parse()
	if *flagPort > 0 {
		port = *flagPort
	}

	store := storage.NewStorage()
	wsManager := ws.NewManager()
	go wsManager.Run()

	h := handlers.NewHandler(store, wsManager)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.POST("/register", h.RegisterPatient)
		api.GET("/registrations", h.GetTodayRegistrations)
		api.POST("/call", h.CallPatient)
		api.POST("/registrations/:id/finish", h.FinishVisit)

		api.POST("/prescriptions", h.CreatePrescription)
		api.PUT("/prescriptions/:id", h.UpdatePrescription)
		api.POST("/prescriptions/:id/confirm", h.ConfirmPrescription)
		api.POST("/prescriptions/:id/void", h.VoidPrescription)
		api.POST("/prescriptions/:id/dispense", h.DispensePrescription)
		api.GET("/prescriptions", h.GetPrescriptions)
		api.GET("/registrations/:registration_id/prescriptions", h.GetPrescriptionsByRegistration)

		api.GET("/herbs", h.GetHerbs)
		api.PUT("/herbs", h.UpdateHerb)

		r.GET("/ws", h.WebSocket)
	}

	addr := ":" + strconv.Itoa(port)
	log.Printf("服务器启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动失败:", err)
	}
}
