package main

import (
	"business-hours/businesshours"
	"business-hours/protocol"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	bh *businesshours.BusinessHours
}

func NewHandler(bh *businesshours.BusinessHours) *Handler {
	return &Handler{bh: bh}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Time.IsZero() {
		req.Time = time.Now()
	}

	result := h.bh.Check(req.Time)

	resp := protocol.CheckResponse{
		IsOpen:       result.IsOpen,
		NextOpenTime: result.NextOpenTime,
		TimeToOpen:   result.TimeToOpen,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CalculateHours(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.CalculateHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := h.bh.CalculateHours(req.StartTime, req.EndTime)

	slots := make([]protocol.TimeSlot, len(result.Slots))
	for i, slot := range result.Slots {
		slots[i] = protocol.TimeSlot{
			Start: slot.Start,
			End:   slot.End,
		}
	}

	resp := protocol.CalculateHoursResponse{
		TotalHours: result.TotalHours,
		Slots:      slots,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Config(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newBh := businesshours.NewBusinessHours()

	for weekdayStr, daySchedule := range req.Weekly {
		weekday, err := parseWeekday(weekdayStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		schedule := &businesshours.DailySchedule{
			OpenSlots:  make([]businesshours.TimeSlot, len(daySchedule.OpenSlots)),
			BreakSlots: make([]businesshours.TimeSlot, len(daySchedule.BreakSlots)),
		}

		for i, slot := range daySchedule.OpenSlots {
			schedule.OpenSlots[i] = businesshours.TimeSlot{
				Start: businesshours.CreateTime(slot.StartHour, slot.StartMinute),
				End:   businesshours.CreateTime(slot.EndHour, slot.EndMinute),
			}
		}

		for i, slot := range daySchedule.BreakSlots {
			schedule.BreakSlots[i] = businesshours.TimeSlot{
				Start: businesshours.CreateTime(slot.StartHour, slot.StartMinute),
				End:   businesshours.CreateTime(slot.EndHour, slot.EndMinute),
			}
		}

		newBh.SetDailySchedule(weekday, schedule)
	}

	for _, holiday := range req.Holidays {
		newBh.AddHoliday(holiday)
	}

	h.bh = newBh

	resp := protocol.ConfigResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func initDefaultBusinessHours() *businesshours.BusinessHours {
	bh := businesshours.NewBusinessHours()

	weekdaySchedule := &businesshours.DailySchedule{
		OpenSlots: []businesshours.TimeSlot{
			{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(18, 0)},
		},
		BreakSlots: []businesshours.TimeSlot{
			{Start: businesshours.CreateTime(12, 0), End: businesshours.CreateTime(13, 0)},
		},
	}

	for weekday := time.Monday; weekday <= time.Friday; weekday++ {
		bh.SetDailySchedule(weekday, weekdaySchedule)
	}

	return bh
}

func parseWeekday(s string) (time.Weekday, error) {
	switch s {
	case "Sunday", "sunday":
		return time.Sunday, nil
	case "Monday", "monday":
		return time.Monday, nil
	case "Tuesday", "tuesday":
		return time.Tuesday, nil
	case "Wednesday", "wednesday":
		return time.Wednesday, nil
	case "Thursday", "thursday":
		return time.Thursday, nil
	case "Friday", "friday":
		return time.Friday, nil
	case "Saturday", "saturday":
		return time.Saturday, nil
	default:
		return 0, nil
	}
}
