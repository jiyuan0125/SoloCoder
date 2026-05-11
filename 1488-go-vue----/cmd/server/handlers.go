package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"housekeeping/internal/api"
	"housekeeping/internal/housekeeping"
)

func convertAunt(a *housekeeping.Aunt) api.Aunt {
	return api.Aunt{
		ID:              a.ID,
		Name:            a.Name,
		Age:             a.Age,
		Phone:           a.Phone,
		ServiceCategory: api.ServiceCategory(a.ServiceCategory),
		YearsOfService:  a.YearsOfService,
		ServiceArea:     a.ServiceArea,
		Rating:          a.Rating,
	}
}

func convertCustomer(c *housekeeping.Customer) api.Customer {
	return api.Customer{
		ID:    c.ID,
		Name:  c.Name,
		Phone: c.Phone,
	}
}

func convertReview(r housekeeping.Review) api.Review {
	return api.Review{
		ID:         r.ID,
		AuntID:     r.AuntID,
		BookingID:  r.BookingID,
		CustomerID: r.CustomerID,
		Rating:     r.Rating,
		Comment:    r.Comment,
		Time:       r.Time,
	}
}

func convertBooking(b *housekeeping.Booking) api.Booking {
	return api.Booking{
		ID:                b.ID,
		CustomerID:        b.CustomerID,
		AuntID:            b.AuntID,
		ServiceCategory:   api.ServiceCategory(b.ServiceCategory),
		ServiceDate:       b.ServiceDate,
		EstimatedDuration: b.EstimatedDuration,
		ActualDuration:    b.ActualDuration,
		Address:           b.Address,
		Status:            api.BookingStatus(b.Status),
		StartTime:         b.StartTime,
		EndTime:           b.EndTime,
		Price:             b.Price,
		CreatedAt:         b.CreatedAt,
	}
}

func convertStats(s *housekeeping.MonthlyStats) api.MonthlyStats {
	return api.MonthlyStats{
		AuntID:        s.AuntID,
		Month:         s.Month,
		TotalBookings: s.TotalBookings,
		TotalHours:    s.TotalHours,
		AverageRating: s.AverageRating,
		RepeatRate:    s.RepeatRate,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func handleRegisterAunt(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.RegisterAuntRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("无效的请求参数"))
			return
		}

		aunt := &housekeeping.Aunt{
			ID:              req.ID,
			Name:            req.Name,
			Age:             req.Age,
			Phone:           req.Phone,
			ServiceCategory: housekeeping.ServiceCategory(req.ServiceCategory),
			YearsOfService:  req.YearsOfService,
			ServiceArea:     req.ServiceArea,
		}

		if err := p.RegisterAunt(aunt); err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusCreated, api.NewSuccessResponse(convertAunt(aunt)))
	}
}

func handleGetAunt(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			aunts := p.AuntManager.GetAllAunts()
			result := make([]api.Aunt, 0, len(aunts))
			for _, aunt := range aunts {
				result = append(result, convertAunt(aunt))
			}
			writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
			return
		}

		aunt, err := p.AuntManager.GetAunt(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertAunt(aunt)))
	}
}

func handleRegisterCustomer(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.RegisterCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("无效的请求参数"))
			return
		}

		customer := &housekeeping.Customer{
			ID:    req.ID,
			Name:  req.Name,
			Phone: req.Phone,
		}

		if err := p.RegisterCustomer(customer); err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusCreated, api.NewSuccessResponse(convertCustomer(customer)))
	}
}

func handleGetCustomer(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			customers := p.BookingManager.GetAllCustomers()
			result := make([]api.Customer, 0, len(customers))
			for _, customer := range customers {
				result = append(result, convertCustomer(customer))
			}
			writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
			return
		}

		customer, err := p.BookingManager.GetCustomer(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertCustomer(customer)))
	}
}

func handleCreateBooking(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.CreateBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("无效的请求参数"))
			return
		}

		bookingReq := housekeeping.BookingRequest{
			CustomerID:        req.CustomerID,
			ServiceCategory:   housekeeping.ServiceCategory(req.ServiceCategory),
			ServiceDate:       req.ServiceDate,
			EstimatedDuration: req.EstimatedDuration,
			Address:           req.Address,
		}

		booking, err := p.CreateBooking(bookingReq)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusCreated, api.NewSuccessResponse(convertBooking(booking)))
	}
}

func handleGetBooking(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		auntID := r.URL.Query().Get("aunt_id")
		customerID := r.URL.Query().Get("customer_id")
		pending := r.URL.Query().Get("pending")

		if id != "" {
			booking, err := p.BookingManager.GetBooking(id)
			if err != nil {
				writeJSON(w, http.StatusNotFound, api.NewErrorResponse(err.Error()))
				return
			}
			writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertBooking(booking)))
			return
		}

		var bookings []*housekeeping.Booking

		if auntID != "" {
			bookings = p.BookingManager.GetBookingsByAunt(auntID)
		} else if customerID != "" {
			bookings = p.BookingManager.GetBookingsByCustomer(customerID)
		} else if pending == "true" {
			bookings = p.BookingManager.GetPendingBookings()
		} else {
			bookings = p.BookingManager.GetAllBookings()
		}

		result := make([]api.Booking, 0, len(bookings))
		for _, booking := range bookings {
			result = append(result, convertBooking(booking))
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
	}
}

func handleStartService(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.StartServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("无效的请求参数"))
			return
		}

		booking, err := p.BookingManager.StartService(req.BookingID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertBooking(booking)))
	}
}

func handleCompleteService(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.CompleteServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("无效的请求参数"))
			return
		}

		booking, err := p.BookingManager.CompleteService(req.BookingID, req.ActualDuration)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertBooking(booking)))
	}
}

func handleAddReview(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.AddReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("无效的请求参数"))
			return
		}

		review, err := p.AddReview(req.AuntID, req.BookingID, req.CustomerID, req.Rating, req.Comment)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusCreated, api.NewSuccessResponse(convertReview(*review)))
	}
}

func handleGetStats(p *housekeeping.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auntID := r.URL.Query().Get("aunt_id")
		month := r.URL.Query().Get("month")

		if auntID == "" || month == "" {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("缺少必要参数: aunt_id, month"))
			return
		}

		stats, err := p.GetMonthlyStats(auntID, month)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.NewErrorResponse(err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertStats(stats)))
	}
}

func _unused() {
	_ = strconv.Atoi
}
