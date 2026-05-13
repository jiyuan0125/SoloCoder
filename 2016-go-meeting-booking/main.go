package main

import (
	"meeting-booking/database"
	"meeting-booking/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()

	r := gin.Default()

	api := r.Group("/api")
	{
		rooms := api.Group("/rooms")
		{
			rooms.GET("", handlers.ListRooms)
			rooms.POST("", handlers.CreateRoom)
		}

		bookings := api.Group("/bookings")
		{
			bookings.POST("", handlers.CreateBooking)
			bookings.DELETE("/:id", handlers.CancelBooking)
		}

		api.GET("/occupancy", handlers.GetOccupancy)
	}

	r.Run(":9701")
}
