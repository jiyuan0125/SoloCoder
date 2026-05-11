package syslog

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	PRI             int
	Facility        Facility
	Severity        Severity
	Version         int
	Timestamp       time.Time
	TimestampFormat string
	Hostname        string
	AppName         string
	ProcID          string
	MsgID           string
	StructuredData  StructuredData
	Message         string
}

func ParseMessage(raw string) (*Message, error) {
	msg := &Message{}

	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	if raw[0] != '<' {
		return nil, fmt.Errorf("missing PRI section")
	}

	closeIdx := strings.Index(raw, ">")
	if closeIdx == -1 {
		return nil, fmt.Errorf("invalid PRI format")
	}

	priStr := raw[1:closeIdx]
	pri, err := strconv.Atoi(priStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PRI value: %s", priStr)
	}

	msg.PRI = pri
	msg.Facility = Facility(pri / 8)
	msg.Severity = Severity(pri % 8)

	afterPRI := raw[closeIdx+1:]

	spaceIdx := strings.Index(afterPRI, " ")
	if spaceIdx == -1 {
		return nil, fmt.Errorf("missing fields after PRI")
	}

	versionStr := afterPRI[:spaceIdx]
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", versionStr)
	}
	msg.Version = version

	rest := afterPRI[spaceIdx+1:]

	var fields []string
	var remaining string

	ts, tsFormat, tsEndIdx, err := parseTimestampFromStart(rest)
	if err != nil {
		return nil, err
	}
	msg.Timestamp = ts
	msg.TimestampFormat = tsFormat

	afterTS := rest[tsEndIdx:]
	if strings.HasPrefix(afterTS, " ") {
		afterTS = afterTS[1:]
	}

	parts := strings.SplitN(afterTS, " ", 5)
	if len(parts) < 4 {
		return nil, fmt.Errorf("not enough fields after timestamp")
	}

	fields = append(fields, "")
	fields = append(fields, parts[:4]...)

	if len(parts) >= 5 {
		remaining = parts[4]
	} else {
		remaining = ""
	}

	msg.Hostname = parseNilValue(fields[1])
	msg.AppName = parseNilValue(fields[2])
	msg.ProcID = parseNilValue(fields[3])
	msg.MsgID = parseNilValue(fields[4])

	if len(remaining) == 0 {
		msg.StructuredData = nil
		return msg, nil
	}

	if remaining[0] == '-' {
		msg.StructuredData = nil
		if len(remaining) > 1 {
			if remaining[1] == ' ' {
				msg.Message = remaining[2:]
			} else {
				msg.Message = remaining[1:]
			}
		}
		return msg, nil
	}

	if remaining[0] != '[' {
		return nil, fmt.Errorf("invalid structured data section")
	}

	sdEnd, msgContent := extractAllSDAndMsg(remaining)
	if sdEnd == 0 {
		return nil, fmt.Errorf("invalid structured data")
	}

	sdStr := remaining[:sdEnd]
	sd, err := ParseStructuredData(sdStr)
	if err != nil {
		return nil, err
	}
	msg.StructuredData = sd

	if msgContent != "" && strings.HasPrefix(msgContent, " ") {
		msgContent = msgContent[1:]
	}
	msg.Message = msgContent

	return msg, nil
}

func parseTimestampFromStart(s string) (time.Time, string, int, error) {
	rfc3339Formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999999Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999Z",
	}

	spaceIdx := strings.Index(s, " ")
	if spaceIdx == -1 {
		spaceIdx = len(s)
	}

	firstToken := s[:spaceIdx]
	for _, format := range rfc3339Formats {
		if t, err := time.Parse(format, firstToken); err == nil {
			return t, "rfc3339", spaceIdx, nil
		}
	}

	normalized := normalizeRFC3339Timestamp(firstToken)
	for _, format := range rfc3339Formats {
		if t, err := time.Parse(format, normalized); err == nil {
			return t, "rfc3339", spaceIdx, nil
		}
	}

	if len(s) >= 15 {
		rfc3164TS := s[:15]
		t, err := parseRFC3164Timestamp(rfc3164TS)
		if err == nil {
			return t, "rfc3164", 15, nil
		}
	}

	return time.Time{}, "", 0, fmt.Errorf("unable to parse timestamp from: %s", s)
}

func extractAllSDAndMsg(s string) (int, string) {
	depth := 0
	inQuote := false
	escaped := false
	lastSDEnd := 0

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inQuote {
			escaped = true
			continue
		}

		if c == '"' {
			inQuote = !inQuote
			continue
		}

		if !inQuote {
			if c == '[' {
				depth++
			} else if c == ']' {
				depth--
				if depth == 0 {
					lastSDEnd = i + 1
					if i+1 < len(s) {
						nextIdx := i + 1
						for nextIdx < len(s) && s[nextIdx] == ' ' {
							nextIdx++
						}
						if nextIdx < len(s) && s[nextIdx] == '[' {
							i = nextIdx - 1
							continue
						}
						return lastSDEnd, s[i+1:]
					}
					return lastSDEnd, ""
				}
			}
		}
	}

	return lastSDEnd, ""
}

func parseNilValue(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

func (m *Message) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("<%d>%d", m.PRI, m.Version))

	if m.Timestamp.IsZero() {
		buf.WriteString(" -")
	} else {
		buf.WriteString(" ")
		buf.WriteString(FormatTimestamp(m.Timestamp, m.TimestampFormat))
	}

	buf.WriteString(" ")
	if m.Hostname == "" {
		buf.WriteString("-")
	} else {
		buf.WriteString(m.Hostname)
	}

	buf.WriteString(" ")
	if m.AppName == "" {
		buf.WriteString("-")
	} else {
		buf.WriteString(m.AppName)
	}

	buf.WriteString(" ")
	if m.ProcID == "" {
		buf.WriteString("-")
	} else {
		buf.WriteString(m.ProcID)
	}

	buf.WriteString(" ")
	if m.MsgID == "" {
		buf.WriteString("-")
	} else {
		buf.WriteString(m.MsgID)
	}

	buf.WriteString(" ")
	if m.StructuredData == nil || len(m.StructuredData) == 0 {
		buf.WriteString("-")
	} else {
		buf.WriteString(m.StructuredData.String())
	}

	buf.WriteString(" ")
	if m.Message != "" {
		buf.WriteString(m.Message)
	}

	return buf.String()
}

func NewMessage(facility Facility, severity Severity, hostname, appName, procID, msgID, message string, ts time.Time, sd StructuredData) *Message {
	return &Message{
		PRI:             CalculatePRI(facility, severity),
		Facility:        facility,
		Severity:        severity,
		Version:         1,
		Timestamp:       ts,
		TimestampFormat: "rfc3339",
		Hostname:        hostname,
		AppName:         appName,
		ProcID:          procID,
		MsgID:           msgID,
		StructuredData:  sd,
		Message:         message,
	}
}
