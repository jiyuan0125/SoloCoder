package ical

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(cal *Calendar, w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	fmt.Fprintln(bw, "BEGIN:VCALENDAR")
	fmt.Fprintln(bw, "VERSION:2.0")

	if cal.ProdID == "" {
		cal.ProdID = "-//iCalTool//EN"
	}
	fmt.Fprintf(bw, "PRODID:%s\n", cal.ProdID)

	for tzID := range cal.Timezones {
		fmt.Fprintln(bw, "BEGIN:VTIMEZONE")
		fmt.Fprintf(bw, "TZID:%s\n", tzID)
		fmt.Fprintln(bw, "END:VTIMEZONE")
	}

	for _, event := range cal.Events {
		g.writeEvent(bw, event)
	}

	fmt.Fprintln(bw, "END:VCALENDAR")
	return nil
}

func (g *Generator) writeEvent(w io.Writer, event *Event) {
	fmt.Fprintln(w, "BEGIN:VEVENT")

	if event.UID != "" {
		fmt.Fprintf(w, "UID:%s\n", event.UID)
	} else {
		fmt.Fprintf(w, "UID:%s\n", fmt.Sprintf("%d", time.Now().UnixNano()))
	}

	if event.Summary != "" {
		fmt.Fprintf(w, "SUMMARY:%s\n", escapeText(event.Summary))
	}

	if event.Description != "" {
		fmt.Fprintf(w, "DESCRIPTION:%s\n", escapeText(event.Description))
	}

	if event.Location != "" {
		fmt.Fprintf(w, "LOCATION:%s\n", escapeText(event.Location))
	}

	if event.Organizer != nil {
		g.writeOrganizer(w, event.Organizer)
	}

	if !event.DTStart.IsZero() {
		fmt.Fprintf(w, "DTSTART:%s\n", formatDateTime(event.DTStart))
	}

	if !event.DTEnd.IsZero() {
		fmt.Fprintf(w, "DTEND:%s\n", formatDateTime(event.DTEnd))
	}

	if !event.DTStamp.IsZero() {
		fmt.Fprintf(w, "DTSTAMP:%s\n", formatDateTime(event.DTStamp))
	}

	if event.RRULE != nil {
		fmt.Fprintf(w, "RRULE:%s\n", formatRRULE(event.RRULE))
	}

	for _, rdate := range event.RDATEs {
		fmt.Fprintf(w, "RDATE:%s\n", formatDateTime(rdate))
	}

	for _, exdate := range event.EXDATEs {
		fmt.Fprintf(w, "EXDATE:%s\n", formatDateTime(exdate))
	}

	fmt.Fprintln(w, "END:VEVENT")
}

func (g *Generator) writeOrganizer(w io.Writer, org *Organizer) {
	var parts []string
	if org.CN != "" {
		parts = append(parts, fmt.Sprintf("CN=%s", org.CN))
	}

	mail := org.Mail
	if !strings.HasPrefix(mail, "mailto:") {
		mail = "mailto:" + mail
	}

	if len(parts) > 0 {
		fmt.Fprintf(w, "ORGANIZER;%s:%s\n", strings.Join(parts, ";"), mail)
	} else {
		fmt.Fprintf(w, "ORGANIZER:%s\n", mail)
	}
}

func formatDateTime(t time.Time) string {
	if t.Location() == time.UTC {
		return t.Format("20060102T150405Z")
	}
	return t.Format("20060102T150405")
}

func formatDate(t time.Time) string {
	return t.Format("20060102")
}

func formatRRULE(r *RRULE) string {
	var parts []string

	if r.FREQ != "" {
		parts = append(parts, fmt.Sprintf("FREQ=%s", r.FREQ))
	}

	if r.INTERVAL != 0 && r.INTERVAL != 1 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", r.INTERVAL))
	}

	if r.COUNT > 0 {
		parts = append(parts, fmt.Sprintf("COUNT=%d", r.COUNT))
	}

	if !r.UNTIL.IsZero() {
		parts = append(parts, fmt.Sprintf("UNTIL=%s", formatDateTime(r.UNTIL)))
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

	return strings.Join(parts, ";")
}

func formatBYDAY(days []WeekdayPos) string {
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

func escapeText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, ";", "\\;")
	return s
}
