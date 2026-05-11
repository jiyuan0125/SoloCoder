package common

import (
	"encoding/json"
	"net/http"
)

type Granularity string

const (
	GranularitySecond Granularity = "second"
	GranularityMinute Granularity = "minute"
	GranularityHour   Granularity = "hour"
)

func (g Granularity) Duration() int64 {
	switch g {
	case GranularitySecond:
		return 1
	case GranularityMinute:
		return 60
	case GranularityHour:
		return 3600
	default:
		return 1
	}
}

type Rule struct {
	Granularity Granularity `json:"granularity"`
	Quota       int         `json:"quota"`
}

type RuleResult struct {
	Granularity Granularity `json:"granularity"`
	Quota       int         `json:"quota"`
	Current     int         `json:"current"`
	Allowed     bool        `json:"allowed"`
}

type AllowRequest struct {
	Path        string `json:"path"`
	ClientID    string `json:"client_id,omitempty"`
	HasClientID bool   `json:"-"`
}

func (r *AllowRequest) UnmarshalJSON(data []byte) error {
	type alias AllowRequest
	var aux struct {
		*alias
		ClientID *string `json:"client_id"`
	}
	aux.alias = (*alias)(r)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.HasClientID = aux.ClientID != nil
	if aux.ClientID != nil {
		r.ClientID = *aux.ClientID
	}
	return nil
}

type AllowResponse struct {
	Allowed       bool          `json:"allowed"`
	RejectedRules []RuleResult  `json:"rejected_rules,omitempty"`
	AllRules      []RuleResult  `json:"all_rules,omitempty"`
}

type ConfigGetRequest struct {
	Path string `json:"path"`
}

type ConfigGetResponse struct {
	Path  string `json:"path"`
	Rules []Rule `json:"rules"`
}

type ConfigSetRequest struct {
	Path  string `json:"path"`
	Rules []Rule `json:"rules"`
}

type ConfigSetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsRequest struct {
	Path        string `json:"path"`
	ClientID    string `json:"client_id,omitempty"`
	HasClientID bool   `json:"-"`
}

func (r *StatsRequest) UnmarshalJSON(data []byte) error {
	type alias StatsRequest
	var aux struct {
		*alias
		ClientID *string `json:"client_id"`
	}
	aux.alias = (*alias)(r)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.HasClientID = aux.ClientID != nil
	if aux.ClientID != nil {
		r.ClientID = *aux.ClientID
	}
	return nil
}

type StatsResponse struct {
	Path     string       `json:"path"`
	ClientID string       `json:"client_id,omitempty"`
	Stats    []RuleResult `json:"stats"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
