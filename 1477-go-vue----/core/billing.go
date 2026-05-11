package core

import (
	"math"
	"time"
)

const (
	FastRatePerKWH   = 120.0 
	SlowRatePerKWH   = 60.0  
	OvertimeMultiplier = 1.5
	ReservationGraceMinutes = 30
	PauseTimeoutHours = 2
)

func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

func CalculateEnergyUsed(startMeter, endMeter float64) float64 {
	return roundToTwoDecimals(endMeter - startMeter)
}

func CalculateChargingAmount(chargerType ChargerType, energyUsed float64, isOvertime bool) int64 {
	var rate float64
	
	if isOvertime {
		if chargerType == FastCharger {
			rate = FastRatePerKWH * OvertimeMultiplier
		} else {
			rate = SlowRatePerKWH * OvertimeMultiplier
		}
	} else {
		if chargerType == FastCharger {
			rate = FastRatePerKWH
		} else {
			rate = SlowRatePerKWH
		}
	}
	
	return int64(math.Round(energyUsed * rate))
}

func CalculateSessionAmount(session *ChargingSession, chargerType ChargerType, reservationEndTime time.Time) (int64, float64) {
	if session.EndTime.IsZero() {
		return 0, 0
	}
	
	energyUsed := CalculateEnergyUsed(session.StartMeter, session.EndMeter)
	
	actualDuration := session.EndTime.Sub(session.StartTime) - session.TotalPauseDuration
	reservationDuration := reservationEndTime.Sub(session.StartTime)
	
	if actualDuration <= 0 {
		return 0, energyUsed
	}
	
	if actualDuration <= reservationDuration {
		return CalculateChargingAmount(chargerType, energyUsed, false), energyUsed
	}
	
	reservedEnergyRatio := float64(reservationDuration) / float64(actualDuration)
	reservedEnergy := roundToTwoDecimals(energyUsed * reservedEnergyRatio)
	overtimeEnergy := roundToTwoDecimals(energyUsed - reservedEnergy)
	
	reservedAmount := CalculateChargingAmount(chargerType, reservedEnergy, false)
	overtimeAmount := CalculateChargingAmount(chargerType, overtimeEnergy, true)
	
	return reservedAmount + overtimeAmount, energyUsed
}

func IsOverGracePeriod(reservationStartTime, checkTime time.Time) bool {
	return checkTime.After(reservationStartTime.Add(ReservationGraceMinutes * time.Minute))
}

func IsPauseTimedOut(pauseStartTime, checkTime time.Time) bool {
	return checkTime.After(pauseStartTime.Add(PauseTimeoutHours * time.Hour))
}

func ValidateTimeSlot(startTime, endTime time.Time) bool {
	if startTime.After(endTime) || startTime.Equal(endTime) {
		return false
	}
	
	duration := endTime.Sub(startTime)
	if duration < 15*time.Minute {
		return false
	}
	
	if startTime.Minute()%15 != 0 || startTime.Second() != 0 || startTime.Nanosecond() != 0 {
		return false
	}
	if endTime.Minute()%15 != 0 || endTime.Second() != 0 || endTime.Nanosecond() != 0 {
		return false
	}
	
	return true
}

func DoTimeSlotsOverlap(start1, end1, start2, end2 time.Time) bool {
	if start1.After(end2) || start1.Equal(end2) {
		return false
	}
	if start2.After(end1) || start2.Equal(end1) {
		return false
	}
	return true
}
