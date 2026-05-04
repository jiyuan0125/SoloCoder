package common

import "regexp"

const (
	LuggageTagLength = 10

	StageCheckIn    = "check_in"
	StageSecurity   = "security"
	StageSorting    = "sorting"
	StageLoading    = "loading"
	StageArrival    = "arrival"
	StageConveyor   = "conveyor"

	StatusNormal      = "normal"
	StatusAnomaly     = "anomaly"
	StatusFlightCancel = "flight_cancel"
	StatusCompleted   = "completed"

	AnomalyTimeoutMinutes = 30

	RoleStaff     = "staff"
	RoleTraveler  = "traveler"
	RoleAdmin     = "admin"
)

var StageOrder = map[string]int{
	StageCheckIn:  1,
	StageSecurity: 2,
	StageSorting:  3,
	StageLoading:  4,
	StageArrival:  5,
	StageConveyor: 6,
}

var StageName = map[string]string{
	StageCheckIn:  "值机",
	StageSecurity: "安检",
	StageSorting:  "分拣",
	StageLoading:  "装机",
	StageArrival:  "到达",
	StageConveyor: "传送带",
}

func ValidateLuggageTag(tag string) bool {
	if len(tag) != LuggageTagLength {
		return false
	}
	match, _ := regexp.MatchString(`^[A-Za-z0-9]+$`, tag)
	return match
}

func ValidateStage(stage string) bool {
	_, exists := StageOrder[stage]
	return exists
}
