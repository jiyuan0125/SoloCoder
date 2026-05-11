package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"interval-overlap/common"
	"interval-overlap/core"
)

const (
	defaultPort = 8080
	envPortKey  = "PORT"
	adminHeader = "X-Admin"
)

var manager *core.Manager

func main() {
	port := getPort()

	manager = core.NewManager(core.Config{
		CheckSameBookerConflict: false,
	})

	http.HandleFunc("/bookings/add", handleAddBooking)
	http.HandleFunc("/bookings/list", handleListBookings)
	http.HandleFunc("/bookings/conflict", handleCheckConflict)
	http.HandleFunc("/bookings/delete", handleDeleteBooking)

	log.Printf("Server starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getPort() int {
	portFlag := flag.Int("port", 0, "Server port")
	flag.Parse()

	if *portFlag > 0 {
		return *portFlag
	}

	if envPort := os.Getenv(envPortKey); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			return p
		}
	}

	return defaultPort
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if response != nil {
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func bookingToInfo(booking *core.Booking) common.BookingInfo {
	return common.BookingInfo{
		ID:       booking.ID,
		Resource: booking.Resource,
		Start:    booking.Interval.Start.Format("2006-01-02T15:04:05Z07:00"),
		End:      booking.Interval.End.Format("2006-01-02T15:04:05Z07:00"),
		Booker:   booking.Booker,
	}
}

func handleAddBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req common.AddBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, common.AddBookingResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	interval, err := core.ParseInterval(req.Start, req.End)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, common.AddBookingResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid time format: %v", err),
		})
		return
	}

	booking, err := core.NewBooking(req.Resource, req.Booker, interval)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, common.AddBookingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create booking: %v", err),
		})
		return
	}

	result, err := manager.AddBooking(booking)
	if err != nil {
		var conflicts []common.BookingInfo
		for _, c := range result.Conflicts {
			conflicts = append(conflicts, bookingToInfo(c))
		}
		sendJSONResponse(w, http.StatusConflict, common.AddBookingResponse{
			Success:   false,
			Conflicts: conflicts,
			Message:   err.Error(),
		})
		return
	}

	sendJSONResponse(w, http.StatusCreated, common.AddBookingResponse{
		Success:   true,
		BookingID: booking.ID,
		Message:   "Booking created successfully",
	})
}

func handleListBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	resource := r.URL.Query().Get("resource")
	if resource == "" {
		sendJSONResponse(w, http.StatusBadRequest, common.ListBookingsResponse{
			Bookings: []common.BookingInfo{},
		})
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var interval *core.Interval
	var err error

	if startStr != "" && endStr != "" {
		interval, err = core.ParseInterval(startStr, endStr)
		if err != nil {
			sendJSONResponse(w, http.StatusBadRequest, common.ListBookingsResponse{
				Bookings: []common.BookingInfo{},
			})
			return
		}
	}

	bookings := manager.ListBookings(resource, interval)

	var bookingInfos []common.BookingInfo
	for _, booking := range bookings {
		bookingInfos = append(bookingInfos, bookingToInfo(booking))
	}

	sendJSONResponse(w, http.StatusOK, common.ListBookingsResponse{
		Bookings: bookingInfos,
	})
}

func handleCheckConflict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req common.CheckConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, common.CheckConflictResponse{
			HasConflict: false,
		})
		return
	}

	interval, err := core.ParseInterval(req.Start, req.End)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, common.CheckConflictResponse{
			HasConflict: false,
		})
		return
	}

	result, err := manager.CheckConflict(req.Resource, interval)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, common.CheckConflictResponse{
			HasConflict: false,
		})
		return
	}

	var conflicts []common.BookingInfo
	for _, c := range result.Conflicts {
		conflicts = append(conflicts, bookingToInfo(c))
	}

	sendJSONResponse(w, http.StatusOK, common.CheckConflictResponse{
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
	})
}

func handleDeleteBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSONResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	adminHeaderValue := r.Header.Get(adminHeader)
	isAdmin := strings.EqualFold(adminHeaderValue, "true")

	var req common.DeleteBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, common.DeleteBookingResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	err := manager.DeleteBooking(req.BookingID, req.Operator, isAdmin)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := err.Error()

		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "permission") {
			statusCode = http.StatusForbidden
		}

		sendJSONResponse(w, statusCode, common.DeleteBookingResponse{
			Success: false,
			Message: message,
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, common.DeleteBookingResponse{
		Success: true,
		Message: "Booking deleted successfully",
	})
}
