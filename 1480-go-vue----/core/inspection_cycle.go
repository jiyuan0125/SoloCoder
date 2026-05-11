package core

import (
	"math"
	"time"

	"vehicle-inspection/common"
)

type InspectionCycle struct {
	Months int
	Label  string
}

func GetInspectionCycle(vehicle *common.Vehicle, now time.Time) InspectionCycle {
	ageYears := now.Sub(vehicle.RegisterDate).Hours() / (24 * 365.25)

	if ageYears < 6 {
		return InspectionCycle{Months: 24, Label: "2年（6年内新车）"}
	} else if ageYears < 15 {
		return InspectionCycle{Months: 12, Label: "1年（6-15年）"}
	}
	return InspectionCycle{Months: 6, Label: "半年（15年以上）"}
}

func CalculateNextInspectionDate(vehicle *common.Vehicle, now time.Time) time.Time {
	cycle := GetInspectionCycle(vehicle, now)
	
	baseDate := vehicle.LastInspectionDate
	if baseDate.IsZero() {
		baseDate = vehicle.RegisterDate
	}
	
	nextDate := baseDate.AddDate(0, cycle.Months, 0)
	return nextDate
}

func IsDueForInspection(vehicle *common.Vehicle, now time.Time) bool {
	nextDate := CalculateNextInspectionDate(vehicle, now)
	daysUntilDue := int(math.Ceil(nextDate.Sub(now).Hours() / 24))
	
	return daysUntilDue <= 30
}

func CalculateDaysUntilDue(vehicle *common.Vehicle, now time.Time) int {
	nextDate := CalculateNextInspectionDate(vehicle, now)
	days := int(math.Ceil(nextDate.Sub(now).Hours() / 24))
	return days
}

func CanBookAppointment(vehicle *common.Vehicle, now time.Time) (bool, string) {
	nextDate := CalculateNextInspectionDate(vehicle, now)
	daysUntilDue := int(math.Ceil(nextDate.Sub(now).Hours() / 24))
	
	if daysUntilDue > 30 {
		return false, "距离下次检测还有超过30天，暂不可预约"
	}
	
	if daysUntilDue < 0 {
		return true, "车辆已过检，请尽快预约检测"
	}
	
	return true, ""
}
