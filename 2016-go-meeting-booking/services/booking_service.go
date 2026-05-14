package services

import (
	"errors"
	"fmt"
	"meeting-booking/database"
	"meeting-booking/models"
	"strings"
	"time"
)

const (
	bufferMinutes = 10
	maxDailyHours = 4
)

func ParseTime(dateStr, timeStr string) (time.Time, error) {
	layout := "2006-01-02 15:04"
	return time.ParseInLocation(layout, fmt.Sprintf("%s %s", dateStr, timeStr), time.Local)
}

func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatTime(t time.Time) string {
	return t.Format("15:04")
}

func HasConflict(roomID uint, date string, startTime, endTime time.Time, excludeID uint) (bool, *models.Booking, error) {
	var bookings []models.Booking
	
	err := database.DB.Where("room_id = ? AND date = ? AND cancelled = ? AND id != ?",
		roomID, date, false, excludeID).Find(&bookings).Error
	if err != nil {
		return false, nil, err
	}

	for _, booking := range bookings {
		bookingStart, _ := ParseTime(booking.Date, booking.StartTime)
		bookingEnd, _ := ParseTime(booking.Date, booking.EndTime)
		bookingEndWithBuffer := AddMinutes(bookingEnd, bufferMinutes)
		
		if startTime.Before(bookingEndWithBuffer) && endTime.After(bookingStart) {
			return true, &booking, nil
		}
	}

	return false, nil, nil
}

func GetDailyBookingDuration(employeeID string, date string) (int, error) {
	var bookings []models.Booking
	err := database.DB.Where("employee_id = ? AND date = ? AND cancelled = ?",
		employeeID, date, false).Find(&bookings).Error
	if err != nil {
		return 0, err
	}

	total := 0
	for _, b := range bookings {
		total += b.DurationMinutes
	}
	return total, nil
}

func CreateBooking(req *CreateBookingRequest) (*BookingResponse, error) {
	var room models.Room
	if err := database.DB.First(&room, req.RoomID).Error; err != nil {
		return nil, errors.New("room not found")
	}

	startTime, err := ParseTime(req.Date, req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}
	endTime := AddMinutes(startTime, req.DurationMinutes)

	if req.IsRecurring {
		return createRecurringBookings(req, &room, startTime, endTime)
	}

	return createSingleBooking(req, &room, startTime, endTime)
}

func createSingleBooking(req *CreateBookingRequest, room *models.Room, startTime, endTime time.Time) (*BookingResponse, error) {
	conflict, conflictBooking, err := HasConflict(req.RoomID, req.Date, startTime, endTime, 0)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, &ConflictError{
			Message: fmt.Sprintf("Conflict with booking on %s %s-%s",
				conflictBooking.Date, conflictBooking.StartTime, conflictBooking.EndTime),
			ConflictBooking: conflictBooking,
		}
	}

	dailyUsed, err := GetDailyBookingDuration(req.EmployeeID, req.Date)
	if err != nil {
		return nil, err
	}
	if dailyUsed+req.DurationMinutes > maxDailyHours*60 {
		return nil, &DailyLimitError{
			Message:       fmt.Sprintf("Daily booking limit exceeded. Already booked %d minutes, requested %d minutes. Limit is %d hours",
				dailyUsed, req.DurationMinutes, maxDailyHours),
			AlreadyBooked: dailyUsed,
		}
	}

	booking := models.Booking{
		RoomID:          req.RoomID,
		EmployeeID:      req.EmployeeID,
		Title:           req.Title,
		Date:            req.Date,
		StartTime:       FormatTime(startTime),
		EndTime:         FormatTime(endTime),
		DurationMinutes: req.DurationMinutes,
		IsRecurring:     false,
	}

	if err := database.DB.Create(&booking).Error; err != nil {
		return nil, err
	}

	return &BookingResponse{
		Bookings:      []models.Booking{booking},
		SkippedWeeks:  []models.SkippedWeek{},
	}, nil
}

func createRecurringBookings(req *CreateBookingRequest, room *models.Room, startTime, endTime time.Time) (*BookingResponse, error) {
	endDate, err := time.ParseInLocation("2006-01-02", req.RecurringEndDate, time.Local)
	if err != nil {
		return nil, errors.New("invalid recurring end date")
	}
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	currentDate := startTime
	weekday := currentDate.Weekday()

	var createdBookings []models.Booking
	var skippedWeeks []models.SkippedWeek

	for !currentDate.After(endDate) {
		dateStr := FormatDate(currentDate)
		currStart, _ := ParseTime(dateStr, req.StartTime)
		currEnd := AddMinutes(currStart, req.DurationMinutes)

		conflict, conflictBooking, err := HasConflict(req.RoomID, dateStr, currStart, currEnd, 0)
		if err != nil {
			return nil, err
		}

		if conflict {
			skippedWeek := models.SkippedWeek{
				Date:   dateStr,
				Reason: fmt.Sprintf("Conflict with booking on %s %s-%s",
					conflictBooking.Date, conflictBooking.StartTime, conflictBooking.EndTime),
			}
			skippedWeeks = append(skippedWeeks, skippedWeek)
		} else {
			dailyUsed, err := GetDailyBookingDuration(req.EmployeeID, dateStr)
			if err != nil {
				return nil, err
			}
			if dailyUsed+req.DurationMinutes > maxDailyHours*60 {
				skippedWeek := models.SkippedWeek{
					Date:   dateStr,
					Reason: fmt.Sprintf("Daily limit exceeded: already booked %d minutes", dailyUsed),
				}
				skippedWeeks = append(skippedWeeks, skippedWeek)
			} else {
				booking := models.Booking{
					RoomID:          req.RoomID,
					EmployeeID:      req.EmployeeID,
					Title:           req.Title,
					Date:            dateStr,
					StartTime:       FormatTime(currStart),
					EndTime:         FormatTime(currEnd),
					DurationMinutes: req.DurationMinutes,
					IsRecurring:     true,
					RecurringEndDate: req.RecurringEndDate,
					RecurringWeekDay: weekday,
				}
				createdBookings = append(createdBookings, booking)
			}
		}

		currentDate = AddDays(currentDate, 7)
	}

	if len(createdBookings) == 0 {
		return nil, errors.New("no bookings could be created for the recurring series")
	}

	for i := range createdBookings {
		if err := database.DB.Create(&createdBookings[i]).Error; err != nil {
			return nil, err
		}
		for j := range skippedWeeks {
			skippedWeeks[j].BookingID = createdBookings[0].ID
		}
	}

	for i := range skippedWeeks {
		if err := database.DB.Create(&skippedWeeks[i]).Error; err != nil {
			return nil, err
		}
	}

	return &BookingResponse{
		Bookings:     createdBookings,
		SkippedWeeks: skippedWeeks,
	}, nil
}

func CancelBooking(bookingID uint, employeeID string) error {
	var booking models.Booking
	if err := database.DB.First(&booking, bookingID).Error; err != nil {
		return errors.New("booking not found")
	}

	if booking.EmployeeID != employeeID {
		return errors.New("unauthorized to cancel this booking")
	}

	if booking.Cancelled {
		return errors.New("booking already cancelled")
	}

	startTime, err := ParseTime(booking.Date, booking.StartTime)
	if err != nil {
		return err
	}

	if time.Now().After(startTime) {
		return &AlreadyStartedError{
			Message: "预约已开始",
		}
	}

	booking.Cancelled = true
	return database.DB.Save(&booking).Error
}

func GetRoomOccupancy(date string) ([]models.RoomOccupancy, error) {
	var rooms []models.Room
	if err := database.DB.Find(&rooms).Error; err != nil {
		return nil, err
	}

	if len(rooms) == 0 {
		return []models.RoomOccupancy{}, nil
	}

	var result []models.RoomOccupancy

	for _, room := range rooms {
		var bookings []models.Booking
		err := database.DB.Where("room_id = ? AND date = ? AND cancelled = ?",
			room.ID, date, false).Order("start_time ASC").Find(&bookings).Error
		if err != nil {
			return nil, err
		}

		occupied := []models.TimeSlot{}
		for _, b := range bookings {
			occupied = append(occupied, models.TimeSlot{
				StartTime: b.StartTime,
				EndTime:   b.EndTime,
			})
		}

		available := calculateAvailableSlots(occupied)

		result = append(result, models.RoomOccupancy{
			RoomID:    room.ID,
			RoomName:  room.Name,
			Occupied:  occupied,
			Available: available,
		})
	}

	return result, nil
}

func calculateAvailableSlots(occupied []models.TimeSlot) []models.TimeSlot {
	if len(occupied) == 0 {
		return []models.TimeSlot{{StartTime: "00:00", EndTime: "23:59"}}
	}

	var available []models.TimeSlot
	dayStart, _ := ParseTime("2024-01-01", "00:00")
	dayEnd, _ := ParseTime("2024-01-01", "23:59")

	prevEnd := dayStart

	for _, slot := range occupied {
		slotStart, _ := ParseTime("2024-01-01", slot.StartTime)
		
		if slotStart.After(prevEnd) {
			available = append(available, models.TimeSlot{
				StartTime: FormatTime(prevEnd),
				EndTime:   FormatTime(slotStart),
			})
		}

		slotEnd, _ := ParseTime("2024-01-01", slot.EndTime)
		prevEnd = slotEnd
	}

	if prevEnd.Before(dayEnd) {
		available = append(available, models.TimeSlot{
			StartTime: FormatTime(prevEnd),
			EndTime:   "23:59",
		})
	}

	return available
}

func GetRoom(roomID uint) (*models.Room, error) {
	var room models.Room
	if err := database.DB.First(&room, roomID).Error; err != nil {
		return nil, errors.New("room not found")
	}
	return &room, nil
}

func CreateRoom(room *models.Room) error {
	return database.DB.Create(room).Error
}

func ListRooms() ([]models.Room, error) {
	var rooms []models.Room
	err := database.DB.Find(&rooms).Error
	return rooms, err
}

type CreateBookingRequest struct {
	RoomID            uint   `json:"room_id" binding:"required"`
	EmployeeID        string `json:"employee_id" binding:"required"`
	Title             string `json:"title"`
	Date              string `json:"date" binding:"required"`
	StartTime         string `json:"start_time" binding:"required"`
	DurationMinutes   int    `json:"duration_minutes" binding:"required,min=1"`
	IsRecurring       bool   `json:"is_recurring"`
	RecurringEndDate  string `json:"recurring_end_date"`
}

type BookingResponse struct {
	Bookings      []models.Booking      `json:"bookings"`
	SkippedWeeks  []models.SkippedWeek  `json:"skipped_weeks,omitempty"`
}

type ConflictError struct {
	Message         string
	ConflictBooking *models.Booking
}

func (e *ConflictError) Error() string {
	return e.Message
}

type DailyLimitError struct {
	Message       string
	AlreadyBooked int
}

func (e *DailyLimitError) Error() string {
	return e.Message
}

type AlreadyStartedError struct {
	Message string
}

func (e *AlreadyStartedError) Error() string {
	return e.Message
}

func init() {
	_ = strings.HasPrefix
}
