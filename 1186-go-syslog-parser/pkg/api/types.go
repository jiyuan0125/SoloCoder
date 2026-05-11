package api

import "time"

type ParseRequest struct {
	Message string `json:"message"`
}

type ParseResponse struct {
	Success bool       `json:"success"`
	Message *SyslogMsg `json:"message,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type GenerateRequest struct {
	Facility        int          `json:"facility"`
	Severity        int          `json:"severity"`
	Version         int          `json:"version"`
	Timestamp       time.Time    `json:"timestamp"`
	Hostname        string       `json:"hostname"`
	AppName         string       `json:"app_name"`
	ProcID          string       `json:"proc_id"`
	MsgID           string       `json:"msg_id"`
	StructuredData  []SDElement  `json:"structured_data"`
	Message         string       `json:"message"`
}

type GenerateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BatchParseRequest struct {
	Messages []string `json:"messages"`
}

type BatchParseResponse struct {
	Success  bool              `json:"success"`
	Messages []*ParsedSyslogMsg `json:"messages,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type ParsedSyslogMsg struct {
	Raw     string    `json:"raw"`
	Parsed  *SyslogMsg `json:"parsed,omitempty"`
	Error   string    `json:"error,omitempty"`
}

type SyslogMsg struct {
	PRI             int         `json:"pri"`
	Facility        int         `json:"facility"`
	FacilityName    string      `json:"facility_name"`
	Severity        int         `json:"severity"`
	SeverityName    string      `json:"severity_name"`
	Version         int         `json:"version"`
	Timestamp       time.Time   `json:"timestamp"`
	TimestampFormat string      `json:"timestamp_format"`
	Hostname        string      `json:"hostname"`
	AppName         string      `json:"app_name"`
	ProcID          string      `json:"proc_id"`
	MsgID           string      `json:"msg_id"`
	StructuredData  []SDElement `json:"structured_data"`
	Message         string      `json:"message"`
}

type SDElement struct {
	ID     string          `json:"id"`
	Params []SDParam       `json:"params"`
}

type SDParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FacilitySeverityResponse struct {
	Success  bool              `json:"success"`
	Facility []FacilityInfo    `json:"facility,omitempty"`
	Severity []SeverityInfo    `json:"severity,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type FacilityInfo struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

type SeverityInfo struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}
