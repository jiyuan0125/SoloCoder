package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/icaltool/api"
	"github.com/example/icaltool/internal/ical"
	"github.com/example/icaltool/internal/rrule"
)

type Server struct {
	ical   *ical.ICal
	expander *rrule.Expander
}

func NewServer() *Server {
	return &Server{
		ical:    ical.New(),
		expander: rrule.NewExpander(),
	}
}

func (s *Server) parseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	cal, err := s.ical.Parse(strings.NewReader(req.ICalData))
	if err != nil {
		sendError(w, http.StatusBadRequest, "parse error: "+err.Error())
		return
	}

	resp := api.ParseResponse{
		Events: make([]api.EventDTO, 0, len(cal.Events)),
	}

	for _, e := range cal.Events {
		resp.Events = append(resp.Events, eventToDTO(e))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	cal := &ical.Calendar{
		Version: "2.0",
		ProdID:  "-//iCalServer//EN",
		Events:  make([]*ical.Event, 0, len(req.Events)),
	}

	for _, dto := range req.Events {
		event, err := dtoToEvent(&dto)
		if err != nil {
			sendError(w, http.StatusBadRequest, "event error: "+err.Error())
			return
		}
		cal.Events = append(cal.Events, event)
	}

	var buf bytes.Buffer
	if err := s.ical.Generate(cal, &buf); err != nil {
		sendError(w, http.StatusInternalServerError, "generate error: "+err.Error())
		return
	}

	resp := api.GenerateResponse{
		ICalData: buf.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) expandRulesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ExpandRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	parsedRRULE, err := ical.ParseRRULE(req.RRULE)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid RRULE: "+err.Error())
		return
	}

	dtStart, err := parseDateTimeWithTZ(req.DTStart, req.TZID)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid DTSTART: "+err.Error())
		return
	}

	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid start time: "+err.Error())
		return
	}

	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid end time: "+err.Error())
		return
	}

	opts := rrule.ExpandOptions{
		Start: start,
		End:   end,
	}

	instances := s.expander.ExpandRRULE(parsedRRULE, dtStart, opts)

	resp := api.ExpandRulesResponse{
		Instances: make([]string, 0, len(instances)),
	}

	for _, inst := range instances {
		resp.Instances = append(resp.Instances, inst.Format(time.RFC3339))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: msg})
}

func eventToDTO(e *ical.Event) api.EventDTO {
	dto := api.EventDTO{
		UID:         e.UID,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		DTStart:     formatTime(e.DTStart),
	}

	if !e.DTEnd.IsZero() {
		dto.DTEnd = formatTime(e.DTEnd)
	}

	if !e.DTStamp.IsZero() {
		dto.DTStamp = formatTime(e.DTStamp)
	}

	if e.Organizer != nil {
		dto.Organizer = &api.OrganizerDTO{
			CN:   e.Organizer.CN,
			Mail: e.Organizer.Mail,
		}
	}

	if e.RRULE != nil {
		ruleStr, _ := formatRRULEString(e.RRULE)
		dto.RRULE = ruleStr
	}

	if len(e.RDATEs) > 0 {
		dto.RDATEs = make([]string, 0, len(e.RDATEs))
		for _, rd := range e.RDATEs {
			dto.RDATEs = append(dto.RDATEs, formatTime(rd))
		}
	}

	if len(e.EXDATEs) > 0 {
		dto.EXDATEs = make([]string, 0, len(e.EXDATEs))
		for _, ed := range e.EXDATEs {
			dto.EXDATEs = append(dto.EXDATEs, formatTime(ed))
		}
	}

	return dto
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatRRULEString(r *ical.RRULE) (string, error) {
	var parts []string

	if r.FREQ != "" {
		parts = append(parts, fmt.Sprintf("FREQ=%s", r.FREQ))
	}

	if r.INTERVAL != 1 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", r.INTERVAL))
	}

	if r.COUNT > 0 {
		parts = append(parts, fmt.Sprintf("COUNT=%d", r.COUNT))
	}

	if !r.UNTIL.IsZero() {
		parts = append(parts, fmt.Sprintf("UNTIL=%s", formatUTCTime(r.UNTIL)))
	}

	if len(r.BYDAY) > 0 {
		parts = append(parts, "BYDAY="+formatBYDAY(r.BYDAY))
	}

	if len(r.BYMONTH) > 0 {
		parts = append(parts, "BYMONTH="+intsToCSV(r.BYMONTH))
	}

	if len(r.BYMONTHDAY) > 0 {
		parts = append(parts, "BYMONTHDAY="+intsToCSV(r.BYMONTHDAY))
	}

	if len(r.BYSETPOS) > 0 {
		parts = append(parts, "BYSETPOS="+intsToCSV(r.BYSETPOS))
	}

	return strings.Join(parts, ";"), nil
}

func formatUTCTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

func formatBYDAY(days []ical.WeekdayPos) string {
	parts := make([]string, 0, len(days))
	for _, d := range days {
		dayStr := weekdayToString(d.Day)
		if d.Pos != 0 {
			parts = append(parts, fmt.Sprintf("%d%s", d.Pos, dayStr))
		} else {
			parts = append(parts, dayStr)
		}
	}
	return strings.Join(parts, ",")
}

func weekdayToString(d time.Weekday) string {
	switch d {
	case time.Sunday:
		return "SU"
	case time.Monday:
		return "MO"
	case time.Tuesday:
		return "TU"
	case time.Wednesday:
		return "WE"
	case time.Thursday:
		return "TH"
	case time.Friday:
		return "FR"
	case time.Saturday:
		return "SA"
	default:
		return ""
	}
}

func intsToCSV(nums []int) string {
	parts := make([]string, 0, len(nums))
	for _, n := range nums {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ",")
}

func dtoToEvent(dto *api.EventDTO) (*ical.Event, error) {
	event := &ical.Event{
		UID:         dto.UID,
		Summary:     dto.Summary,
		Description: dto.Description,
		Location:    dto.Location,
	}

	if dto.DTStart != "" {
		t, err := time.Parse(time.RFC3339, dto.DTStart)
		if err != nil {
			return nil, err
		}
		event.DTStart = t
	}

	if dto.DTEnd != "" {
		t, err := time.Parse(time.RFC3339, dto.DTEnd)
		if err != nil {
			return nil, err
		}
		event.DTEnd = t
	}

	if dto.DTStamp != "" {
		t, err := time.Parse(time.RFC3339, dto.DTStamp)
		if err != nil {
			return nil, err
		}
		event.DTStamp = t
	}

	if dto.Organizer != nil {
		event.Organizer = &ical.Organizer{
			CN:   dto.Organizer.CN,
			Mail: dto.Organizer.Mail,
		}
	}

	if dto.RRULE != "" {
		rule, err := ical.ParseRRULE(dto.RRULE)
		if err != nil {
			return nil, err
		}
		event.RRULE = rule
	}

	if len(dto.RDATEs) > 0 {
		event.RDATEs = make([]time.Time, 0, len(dto.RDATEs))
		for _, rd := range dto.RDATEs {
			t, err := time.Parse(time.RFC3339, rd)
			if err != nil {
				return nil, err
			}
			event.RDATEs = append(event.RDATEs, t)
		}
	}

	if len(dto.EXDATEs) > 0 {
		event.EXDATEs = make([]time.Time, 0, len(dto.EXDATEs))
		for _, ed := range dto.EXDATEs {
			t, err := time.Parse(time.RFC3339, ed)
			if err != nil {
				return nil, err
			}
			event.EXDATEs = append(event.EXDATEs, t)
		}
	}

	return event, nil
}

func parseDateTimeWithTZ(dtStr, tzID string) (time.Time, error) {
	if tzID != "" {
		loc, err := time.LoadLocation(tzID)
		if err != nil {
			loc = time.Local
		}
		if strings.HasSuffix(dtStr, "Z") {
			t, err := time.Parse("20060102T150405Z", dtStr)
			if err == nil {
				return t, nil
			}
		}
		t, err := time.ParseInLocation("20060102T150405", dtStr, loc)
		if err == nil {
			return t, nil
		}
	}

	if strings.HasSuffix(dtStr, "Z") {
		return time.Parse("20060102T150405Z", dtStr)
	}

	if t, err := time.Parse("20060102T150405", dtStr); err == nil {
		return t, nil
	}

	return time.Parse(time.RFC3339, dtStr)
}

func getPort() string {
	port := "8104"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "HTTP server port")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	return port
}

func main() {
	port := getPort()

	server := NewServer()

	http.HandleFunc("/parse", server.parseHandler)
	http.HandleFunc("/generate", server.generateHandler)
	http.HandleFunc("/expand-rules", server.expandRulesHandler)

	addr := ":" + port
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
