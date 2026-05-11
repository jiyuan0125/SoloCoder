package ical

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(r io.Reader) (*Calendar, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	cal := &Calendar{
		Timezones: make(map[string]*Timezone),
	}

	var currentSection string
	var currentEvent *Event
	var currentTZ *Timezone

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line == "BEGIN:VCALENDAR" {
			currentSection = "VCALENDAR"
			continue
		}

		if line == "END:VCALENDAR" {
			currentSection = ""
			continue
		}

		if line == "BEGIN:VEVENT" {
			currentSection = "VEVENT"
			currentEvent = &Event{}
			continue
		}

		if line == "END:VEVENT" {
			if currentEvent != nil {
				cal.Events = append(cal.Events, currentEvent)
			}
			currentSection = "VCALENDAR"
			currentEvent = nil
			continue
		}

		if line == "BEGIN:VTIMEZONE" {
			currentSection = "VTIMEZONE"
			currentTZ = &Timezone{}
			continue
		}

		if line == "END:VTIMEZONE" {
			if currentTZ != nil && currentTZ.TZID != "" {
				cal.Timezones[currentTZ.TZID] = currentTZ
			}
			currentSection = "VCALENDAR"
			currentTZ = nil
			continue
		}

		if currentSection == "VTIMEZONE" {
			p.parseVTIMEZONELine(currentTZ, line)
			continue
		}

		if currentSection == "VEVENT" {
			if err := p.parseVEVENTLine(currentEvent, line); err != nil {
				return nil, err
			}
			continue
		}

		if currentSection == "VCALENDAR" {
			p.parseCalendarLine(cal, line)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return cal, nil
}

func (p *Parser) parseCalendarLine(cal *Calendar, line string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return
	}

	key := strings.ToUpper(parts[0])
	value := parts[1]

	switch key {
	case "VERSION":
		cal.Version = value
	case "PRODID":
		cal.ProdID = value
	}
}

func (p *Parser) parseVTIMEZONELine(tz *Timezone, line string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return
	}

	key := strings.ToUpper(parts[0])
	value := parts[1]

	paramIdx := strings.Index(key, ";")
	if paramIdx != -1 {
		params := key[paramIdx+1:]
		key = key[:paramIdx]
		if key == "TZID" {
			if strings.HasPrefix(params, "VALUE=") {
			} else {
			}
		}
	}

	if key == "TZID" {
		tz.TZID = value
	}
}

func (p *Parser) parseVEVENTLine(event *Event, line string) error {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return nil
	}

	key := strings.ToUpper(parts[0])
	value := parts[1]

	var tzID string
	paramIdx := strings.Index(key, ";")
	if paramIdx != -1 {
		paramsStr := key[paramIdx+1:]
		key = key[:paramIdx]
		params := parseParams(paramsStr)
		if tz, ok := params["TZID"]; ok {
			tzID = tz
		}
	}

	switch key {
	case "UID":
		event.UID = value
	case "SUMMARY":
		event.Summary = value
	case "DESCRIPTION":
		event.Description = value
	case "LOCATION":
		event.Location = value
	case "ORGANIZER":
		event.Organizer = parseOrganizer(parts[0], value)
	case "DTSTART":
		t, err := parseDateTime(value, tzID)
		if err != nil {
			return err
		}
		event.DTStart = t
	case "DTEND":
		t, err := parseDateTime(value, tzID)
		if err != nil {
			return err
		}
		event.DTEnd = t
	case "DTSTAMP":
		t, err := parseDateTime(value, "")
		if err != nil {
			return err
		}
		event.DTStamp = t
	case "RRULE":
		rrule, err := parseRRULE(value)
		if err != nil {
			return err
		}
		event.RRULE = rrule
	case "RDATE":
		t, err := parseDateTime(value, tzID)
		if err != nil {
			return err
		}
		event.RDATEs = append(event.RDATEs, t)
	case "EXDATE":
		t, err := parseDateTime(value, tzID)
		if err != nil {
			return err
		}
		event.EXDATEs = append(event.EXDATEs, t)
	}

	return nil
}

func parseParams(s string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(s, ";")
	for _, p := range parts {
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else if len(kv) == 1 {
			result[kv[0]] = ""
		}
	}
	return result
}

func parseDateTime(value string, tzID string) (time.Time, error) {
	value = strings.TrimSpace(value)
	isUTC := strings.HasSuffix(value, "Z")
	format := "20060102T150405"
	dateOnlyFormat := "20060102"

	if isUTC {
		value = strings.TrimSuffix(value, "Z")
		t, err := time.Parse(format, value)
		if err != nil {
			if t2, err2 := time.Parse(dateOnlyFormat, value); err2 == nil {
				return t2.UTC(), nil
			}
			return time.Time{}, err
		}
		return t.UTC(), nil
	}

	if len(value) == 8 {
		t, err := time.ParseInLocation(dateOnlyFormat, value, time.Local)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}

	if tzID != "" {
		loc, err := loadLocation(tzID)
		if err != nil {
			return time.Time{}, err
		}
		t, err := time.ParseInLocation(format, value, loc)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}

	t, err := time.Parse(format, value)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func loadLocation(tzID string) (*time.Location, error) {
	if tzID == "" {
		return time.Local, nil
	}
	loc, err := time.LoadLocation(tzID)
	if err != nil {
		return time.Local, nil
	}
	return loc, nil
}

func parseOrganizer(paramsStr, value string) *Organizer {
	params := parseParams(paramsStr)
	org := &Organizer{
		Mail: strings.TrimPrefix(value, "mailto:"),
	}
	if cn, ok := params["CN"]; ok {
		org.CN = cn
	}
	return org
}

func parseRRULE(value string) (*RRULE, error) {
	rrule := &RRULE{
		INTERVAL: 1,
	}

	parts := strings.Split(value, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.ToUpper(kv[0])
		val := kv[1]

		switch key {
		case "FREQ":
			rrule.FREQ = Frequency(strings.ToUpper(val))
		case "INTERVAL":
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			rrule.INTERVAL = n
		case "COUNT":
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			rrule.COUNT = n
		case "UNTIL":
			t, err := parseDateTime(val, "")
			if err != nil {
				return nil, err
			}
			rrule.UNTIL = t
		case "BYDAY":
			wd, err := parseBYDAY(val)
			if err != nil {
				return nil, err
			}
			rrule.BYDAY = wd
		case "BYMONTH":
			vals := strings.Split(val, ",")
			for _, v := range vals {
				n, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				rrule.BYMONTH = append(rrule.BYMONTH, n)
			}
		case "BYMONTHDAY":
			vals := strings.Split(val, ",")
			for _, v := range vals {
				n, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				rrule.BYMONTHDAY = append(rrule.BYMONTHDAY, n)
			}
		case "BYSETPOS":
			vals := strings.Split(val, ",")
			for _, v := range vals {
				n, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				rrule.BYSETPOS = append(rrule.BYSETPOS, n)
			}
		}
	}

	return rrule, nil
}

func parseBYDAY(val string) ([]WeekdayPos, error) {
	vals := strings.Split(val, ",")
	result := make([]WeekdayPos, 0, len(vals))

	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		var pos int = 0
		dayStr := v

		if len(v) > 2 {
			idx := 0
			if v[0] == '-' {
				idx = 1
			}
			for idx < len(v) && v[idx] >= '0' && v[idx] <= '9' {
				idx++
			}
			if idx > 0 {
				n, err := strconv.Atoi(v[:idx])
				if err == nil {
					pos = n
					dayStr = v[idx:]
				}
			}
		}

		wd, err := parseWeekday(dayStr)
		if err != nil {
			return nil, err
		}

		result = append(result, WeekdayPos{
			Day: wd,
			Pos: pos,
		})
	}

	return result, nil
}

func parseWeekday(s string) (time.Weekday, error) {
	switch strings.ToUpper(s) {
	case "SU", "SUN", "SUNDAY":
		return time.Sunday, nil
	case "MO", "MON", "MONDAY":
		return time.Monday, nil
	case "TU", "TUE", "TUESDAY":
		return time.Tuesday, nil
	case "WE", "WED", "WEDNESDAY":
		return time.Wednesday, nil
	case "TH", "THU", "THURSDAY":
		return time.Thursday, nil
	case "FR", "FRI", "FRIDAY":
		return time.Friday, nil
	case "SA", "SAT", "SATURDAY":
		return time.Saturday, nil
	default:
		return time.Sunday, fmt.Errorf("unknown weekday: %s", s)
	}
}
