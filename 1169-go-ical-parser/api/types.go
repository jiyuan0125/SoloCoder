package api

type ParseRequest struct {
	ICalData string `json:"ical_data"`
}

type ParseResponse struct {
	Events []EventDTO `json:"events"`
}

type GenerateRequest struct {
	Events []EventDTO `json:"events"`
}

type GenerateResponse struct {
	ICalData string `json:"ical_data"`
}

type ExpandRulesRequest struct {
	RRULE    string `json:"rrule"`
	DTStart  string `json:"dtstart"`
	TZID     string `json:"tzid,omitempty"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

type ExpandRulesResponse struct {
	Instances []string `json:"instances"`
}

type EventDTO struct {
	UID         string        `json:"uid,omitempty"`
	Summary     string        `json:"summary,omitempty"`
	Description string        `json:"description,omitempty"`
	Location    string        `json:"location,omitempty"`
	Organizer   *OrganizerDTO `json:"organizer,omitempty"`
	DTStart     string        `json:"dtstart"`
	DTEnd       string        `json:"dtend,omitempty"`
	DTStamp     string        `json:"dtstamp,omitempty"`
	RRULE       string        `json:"rrule,omitempty"`
	RDATEs      []string      `json:"rdates,omitempty"`
	EXDATEs     []string      `json:"exdates,omitempty"`
}

type OrganizerDTO struct {
	CN   string `json:"cn,omitempty"`
	Mail string `json:"mail,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
