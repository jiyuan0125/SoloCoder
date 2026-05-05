package main

import (
	"encoding/json"
	"net/http"
	"time"

	"agecalculator/agecalc"
	"agecalculator/common"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/age", s.handleAge)
	mux.HandleFunc("/api/days-between", s.handleDaysBetween)

	return http.ListenAndServe(s.addr, mux)
}

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation(common.DateFormat, dateStr, time.Local)
}

func (s *Server) handleAge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	birthDate, err := parseDate(req.BirthDate)
	if err != nil {
		http.Error(w, "Invalid birth_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	currentDate, err := parseDate(req.CurrentDate)
	if err != nil {
		http.Error(w, "Invalid current_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var resp common.AgeResponse

	age, err := agecalc.CalculateAge(birthDate, currentDate)
	if err != nil {
		resp.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Age.Years = age.Years
	resp.Age.Months = age.Months
	resp.Age.Days = age.Days

	ageYears, err := agecalc.CalculateAgeYears(birthDate, currentDate)
	if err != nil {
		resp.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.AgeYears = ageYears

	isAdult, err := agecalc.IsAdult(birthDate, currentDate)
	if err != nil {
		resp.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.IsAdult = isAdult

	daysUntil, err := agecalc.DaysUntilNextBirthday(birthDate, currentDate)
	if err != nil {
		resp.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.DaysUntilBirthday = daysUntil

	if req.Gender == agecalc.Male || req.Gender == agecalc.Female {
		isRetired, err := agecalc.HasReachedRetirementAge(birthDate, currentDate, req.Gender)
		if err != nil {
			resp.Error = err.Error()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp.IsRetired = isRetired
	}

	if req.TargetAge > 0 {
		hasReached, err := agecalc.HasReachedAge(birthDate, currentDate, req.TargetAge)
		if err != nil {
			resp.Error = err.Error()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp.HasReachedTargetAge = hasReached
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDaysBetween(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DaysBetweenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		http.Error(w, "Invalid end_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var resp common.DaysBetweenResponse
	resp.Days = agecalc.DaysBetween(startDate, endDate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
