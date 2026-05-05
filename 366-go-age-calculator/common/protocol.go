package common

import "agecalculator/agecalc"

type AgeRequest struct {
	BirthDate   string       `json:"birth_date"`
	CurrentDate string       `json:"current_date,omitempty"`
	Gender      agecalc.Gender `json:"gender,omitempty"`
	TargetAge   int          `json:"target_age,omitempty"`
}

type AgeResponse struct {
	Age struct {
		Years  int `json:"years"`
		Months int `json:"months"`
		Days   int `json:"days"`
	} `json:"age"`
	AgeYears           int    `json:"age_years"`
	IsAdult            bool   `json:"is_adult"`
	IsRetired          bool   `json:"is_retired,omitempty"`
	HasReachedTargetAge bool  `json:"has_reached_target_age,omitempty"`
	DaysUntilBirthday  int    `json:"days_until_birthday"`
	Error              string `json:"error,omitempty"`
}

type DaysBetweenRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type DaysBetweenResponse struct {
	Days  int    `json:"days"`
	Error string `json:"error,omitempty"`
}

const DateFormat = "2006-01-02"
