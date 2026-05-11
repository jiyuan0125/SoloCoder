package core

import (
	"math"
	"time"
)

const (
	FreeMinutes       = 30
	HourlyRate        = 4.0
	DailyCap          = 40.0
	MinutesPerHour    = 60
	MinutesPerDay     = 24 * 60
)

func CalculateFee(checkIn, checkOut time.Time) float64 {
	totalMinutes := checkOut.Sub(checkIn).Minutes()
	if totalMinutes <= FreeMinutes {
		return 0.0
	}

	chargeableMinutes := totalMinutes - FreeMinutes
	if chargeableMinutes <= 0 {
		return 0.0
	}

	return calculateFeeByDay(checkIn, checkOut, chargeableMinutes)
}

func calculateFeeByDay(checkIn, checkOut time.Time, chargeableMinutes float64) float64 {
	var totalFee float64 = 0.0

	checkInDate := truncateToDate(checkIn)
	checkOutDate := truncateToDate(checkOut)

	if checkInDate.Equal(checkOutDate) {
		hours := math.Ceil(chargeableMinutes / MinutesPerHour)
		totalFee = hours * HourlyRate
		if totalFee > DailyCap {
			totalFee = DailyCap
		}
		return roundToTwoDecimals(totalFee)
	}

	dayStart := checkInDate.Add(24 * time.Hour)
	remainingMinutes := chargeableMinutes

	for d := checkInDate; !d.After(checkOutDate); d = d.AddDate(0, 0, 1) {
		var dayMinutes float64

		if d.Equal(checkInDate) {
			endOfDay := d.Add(24 * time.Hour)
			elapsedOnFirstDay := endOfDay.Sub(checkIn).Minutes()
			if elapsedOnFirstDay > FreeMinutes {
				dayMinutes = elapsedOnFirstDay - FreeMinutes
			} else {
				dayMinutes = 0
			}
		} else if d.Equal(checkOutDate) {
			startOfDay := d
			elapsedOnLastDay := checkOut.Sub(startOfDay).Minutes()
			dayMinutes = math.Min(remainingMinutes, elapsedOnLastDay)
		} else {
			dayMinutes = MinutesPerDay
		}

		if dayMinutes > remainingMinutes {
			dayMinutes = remainingMinutes
		}

		if dayMinutes > 0 {
			hours := math.Ceil(dayMinutes / MinutesPerHour)
			dayFee := hours * HourlyRate
			if dayFee > DailyCap {
				dayFee = DailyCap
			}
			totalFee += dayFee
		}

		remainingMinutes -= dayMinutes
		if remainingMinutes <= 0 {
			break
		}

		_ = dayStart
	}

	return roundToTwoDecimals(totalFee)
}

func truncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func roundToTwoDecimals(f float64) float64 {
	return math.Round(f*100) / 100
}

func GetFreeMinutesUsed(checkIn, checkOut time.Time) float64 {
	totalMinutes := checkOut.Sub(checkIn).Minutes()
	if totalMinutes <= FreeMinutes {
		return totalMinutes
	}
	return float64(FreeMinutes)
}
