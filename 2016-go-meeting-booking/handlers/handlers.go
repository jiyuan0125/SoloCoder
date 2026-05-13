package handlers

import (
	"errors"
	"meeting-booking/models"
	"meeting-booking/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateRoom(c *gin.Context) {
	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.CreateRoom(&room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

func ListRooms(c *gin.Context) {
	rooms, err := services.ListRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func CreateBooking(c *gin.Context) {
	var req services.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsRecurring && req.RecurringEndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recurring_end_date is required for recurring bookings"})
		return
	}

	if !req.IsRecurring {
		req.RecurringEndDate = ""
	}

	result, err := services.CreateBooking(&req)
	if err != nil {
		switch e := err.(type) {
		case *services.ConflictError:
			c.JSON(http.StatusConflict, gin.H{
				"error":           "Booking conflict",
				"conflict_booking": gin.H{
					"id":         e.ConflictBooking.ID,
					"date":       e.ConflictBooking.Date,
					"start_time": e.ConflictBooking.StartTime,
					"end_time":   e.ConflictBooking.EndTime,
					"employee_id": e.ConflictBooking.EmployeeID,
				},
			})
		case *services.DailyLimitError:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":            "Daily booking limit exceeded",
				"already_booked_minutes": e.AlreadyBooked,
				"limit_minutes":    4 * 60,
			})
		default:
			if err.Error() == "room not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

func CancelBooking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	var req struct {
		EmployeeID string `json:"employee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = services.CancelBooking(uint(id), req.EmployeeID)
	if err != nil {
		switch e := err.(type) {
		case *services.AlreadyStartedError:
			c.JSON(http.StatusBadRequest, gin.H{"error": e.Message})
		default:
			if err.Error() == "booking not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			} else if err.Error() == "unauthorized to cancel this booking" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}

func GetOccupancy(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query parameter is required"})
		return
	}

	occupancy, err := services.GetRoomOccupancy(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, occupancy)
}

func init() {
	_ = errors.New("")
}
