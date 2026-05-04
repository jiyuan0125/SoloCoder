package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"gym-membership/types"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func validatePhone(phone string) bool {
	match, _ := regexp.MatchString(`^1\d{10}$`, phone)
	return match
}

func validateClassName(name string) bool {
	return len(name) > 0 && len(name) <= 30
}

func validateStartTime(timeStr string) bool {
	match, _ := regexp.MatchString(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`, timeStr)
	return match
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func (h *Handler) CreateMember(w http.ResponseWriter, r *http.Request) {
	var req types.CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	if !validatePhone(req.Phone) {
		h.respondError(w, http.StatusBadRequest, ErrInvalidPhone.Error())
		return
	}

	member, err := h.store.CreateMember(req.Name, req.Phone, req.Password)
	if err != nil {
		if err == ErrMemberExists {
			h.respondError(w, http.StatusConflict, err.Error())
		} else {
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, types.CreateMemberResponse{
		Success: true,
		Member:  member,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	member, err := h.store.GetMemberByPhone(req.Phone)
	if err != nil {
		h.respondError(w, http.StatusNotFound, ErrMemberNotFound.Error())
		return
	}

	passwordHash := hashPassword(req.Password)
	if member.Password != passwordHash {
		h.respondError(w, http.StatusUnauthorized, ErrInvalidPassword.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, types.LoginResponse{
		Success: true,
		Member:  member,
		Token:   member.ID,
	})
}

func (h *Handler) PurchaseCard(w http.ResponseWriter, r *http.Request) {
	var req types.PurchaseCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	card, err := h.store.PurchaseCard(req.MemberID, req.CardType)
	if err != nil {
		switch err {
		case ErrMemberNotFound:
			h.respondError(w, http.StatusNotFound, err.Error())
		case ErrActiveCardExists:
			h.respondError(w, http.StatusConflict, err.Error())
		case ErrInvalidCardType:
			h.respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, types.PurchaseCardResponse{
		Success: true,
		Card:    card,
	})
}

func (h *Handler) RenewCard(w http.ResponseWriter, r *http.Request) {
	var req types.RenewCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	card, err := h.store.RenewCard(req.MemberID, req.CardType)
	if err != nil {
		if err == ErrInvalidCardType {
			h.respondError(w, http.StatusBadRequest, err.Error())
		} else {
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, types.RenewCardResponse{
		Success: true,
		Card:    card,
	})
}

func (h *Handler) GetMemberInfo(w http.ResponseWriter, r *http.Request) {
	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		h.respondError(w, http.StatusBadRequest, "缺少会员ID")
		return
	}

	member, err := h.store.GetMemberByID(memberID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, ErrMemberNotFound.Error())
		return
	}

	card, _ := h.store.GetMemberCards(memberID)
	activeBookings := h.store.GetMemberBookings(memberID, true)
	attendances := h.store.GetMemberAttendances(memberID)

	var bookingDetails []types.BookingDetail
	for _, b := range activeBookings {
		instance, _ := h.store.classInstances[b.ClassInstanceID]
		class, _ := h.store.classes[instance.ClassID]
		member, _ := h.store.members[b.MemberID]

		bookingDetails = append(bookingDetails, types.BookingDetail{
			Booking:     b,
			MemberName:  member.Name,
			MemberPhone: member.Phone,
			ClassName:   class.Name,
			Date:        instance.Date.Format("2006-01-02"),
		})
	}

	var attendanceDetails []types.AttendanceDetail
	for _, a := range attendances {
		instance, _ := h.store.classInstances[a.ClassInstanceID]
		class, _ := h.store.classes[instance.ClassID]
		member, _ := h.store.members[a.MemberID]

		attendanceDetails = append(attendanceDetails, types.AttendanceDetail{
			Attendance:  a,
			MemberName:  member.Name,
			MemberPhone: member.Phone,
			ClassName:   class.Name,
			Date:        instance.Date.Format("2006-01-02"),
		})
	}

	h.respondJSON(w, http.StatusOK, types.GetMemberInfoResponse{
		Success:           true,
		Member:            member,
		CurrentCard:       card,
		ActiveBookings:    bookingDetails,
		AttendanceHistory: attendanceDetails,
	})
}

func (h *Handler) CreateClass(w http.ResponseWriter, r *http.Request) {
	var req types.CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	if !validateClassName(req.Name) {
		h.respondError(w, http.StatusBadRequest, ErrInvalidClassName.Error())
		return
	}

	class, err := h.store.CreateClass(&req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, types.CreateClassResponse{
		Success: true,
		Class:   class,
	})
}

func (h *Handler) ListClasses(w http.ResponseWriter, r *http.Request) {
	var weekday *time.Weekday
	weekdayStr := r.URL.Query().Get("weekday")
	if weekdayStr != "" {
		var wd time.Weekday
		switch weekdayStr {
		case "0":
			wd = time.Sunday
		case "1":
			wd = time.Monday
		case "2":
			wd = time.Tuesday
		case "3":
			wd = time.Wednesday
		case "4":
			wd = time.Thursday
		case "5":
			wd = time.Friday
		case "6":
			wd = time.Saturday
		}
		weekday = &wd
	}

	classes := h.store.ListClasses(weekday)

	h.respondJSON(w, http.StatusOK, types.ListClassesResponse{
		Success: true,
		Classes: classes,
	})
}

func (h *Handler) BookClass(w http.ResponseWriter, r *http.Request) {
	var req types.BookClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "日期格式错误，应为2006-01-02")
		return
	}

	instance, err := h.store.GetClassInstance(req.ClassID, date)
	if err != nil {
		h.respondError(w, http.StatusNotFound, ErrClassInstanceNotFound.Error())
		return
	}

	booking, err := h.store.BookClass(req.MemberID, instance.ID)
	if err != nil {
		switch err {
		case ErrClassInstanceNotFound:
			h.respondError(w, http.StatusNotFound, err.Error())
		case ErrClassNotAvailable:
			h.respondError(w, http.StatusBadRequest, err.Error())
		case ErrAlreadyBooked:
			h.respondError(w, http.StatusConflict, err.Error())
		case ErrClassFull:
			h.respondError(w, http.StatusConflict, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, types.BookClassResponse{
		Success: true,
		Booking: booking,
	})
}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	var req types.CancelBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	err := h.store.CancelBooking(req.BookingID, req.MemberID)
	if err != nil {
		switch err {
		case ErrBookingNotFound:
			h.respondError(w, http.StatusNotFound, err.Error())
		case ErrInvalidBookingStatus:
			h.respondError(w, http.StatusBadRequest, err.Error())
		case ErrCancelTooLate:
			h.respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, types.CancelBookingResponse{
		Success: true,
		Message: "预约已取消",
	})
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	var req types.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "日期格式错误，应为2006-01-02")
		return
	}

	instance, err := h.store.GetClassInstance(req.ClassID, date)
	if err != nil {
		h.respondError(w, http.StatusNotFound, ErrClassInstanceNotFound.Error())
		return
	}

	attendance, err := h.store.CheckIn(req.MemberID, instance.ID)
	if err != nil {
		switch err {
		case ErrClassInstanceNotFound:
			h.respondError(w, http.StatusNotFound, err.Error())
		case ErrClassNotAvailable:
			h.respondError(w, http.StatusBadRequest, err.Error())
		case ErrNoBookingForClass:
			h.respondError(w, http.StatusBadRequest, err.Error())
		case ErrAlreadyCheckedIn:
			h.respondError(w, http.StatusConflict, err.Error())
		case ErrCheckInTimeInvalid:
			h.respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, types.CheckInResponse{
		Success:    true,
		Attendance: attendance,
	})
}

func (h *Handler) GetClassBookings(w http.ResponseWriter, r *http.Request) {
	classID := r.URL.Query().Get("class_id")
	dateStr := r.URL.Query().Get("date")

	if classID == "" || dateStr == "" {
		h.respondError(w, http.StatusBadRequest, "缺少必要参数")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "日期格式错误，应为2006-01-02")
		return
	}

	class, err := h.store.GetClassByID(classID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, ErrClassNotFound.Error())
		return
	}

	instance, err := h.store.GetClassInstance(classID, date)
	if err != nil {
		h.respondError(w, http.StatusNotFound, ErrClassInstanceNotFound.Error())
		return
	}

	bookings := h.store.GetInstanceBookings(instance.ID)
	attendances := h.store.GetInstanceAttendances(instance.ID)

	var bookingDetails []types.BookingDetail
	for _, b := range bookings {
		member, _ := h.store.members[b.MemberID]

		bookingDetails = append(bookingDetails, types.BookingDetail{
			Booking:     b,
			MemberName:  member.Name,
			MemberPhone: member.Phone,
			ClassName:   class.Name,
			Date:        instance.Date.Format("2006-01-02"),
		})
	}

	h.respondJSON(w, http.StatusOK, types.GetClassBookingsResponse{
		Success:         true,
		Class:           class,
		ClassInstance:   instance,
		Bookings:        bookingDetails,
		AttendanceCount: len(attendances),
	})
}

func (h *Handler) SetCardPrice(w http.ResponseWriter, r *http.Request) {
	var req types.SetCardPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	price, err := h.store.SetCardPrice(req.CardType, req.Price, req.ValidDays)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, types.SetCardPriceResponse{
		Success: true,
		Price:   price,
	})
}

func (h *Handler) ListCardPrices(w http.ResponseWriter, r *http.Request) {
	prices := h.store.GetCardPrices()

	h.respondJSON(w, http.StatusOK, types.ListCardPricesResponse{
		Success: true,
		Prices:  prices,
	})
}

func (h *Handler) GenerateClassInstances(w http.ResponseWriter, r *http.Request) {
	var req types.GenerateClassInstancesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "开始日期格式错误，应为2006-01-02")
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "结束日期格式错误，应为2006-01-02")
		return
	}

	instances, err := h.store.GenerateClassInstances(startDate, endDate)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, types.GenerateClassInstancesResponse{
		Success:   true,
		Instances: instances,
	})
}
