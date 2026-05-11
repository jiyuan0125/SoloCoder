package core

import (
	"math"
	"time"
)

const (
	daytimeBaseFare      = 13.0
	nighttimeBaseFare    = 16.0
	baseDistanceKm       = 3.0
	daytimePerKmRate     = 2.3
	nighttimePerKmRate   = 2.8
	lowSpeedFeePerMinute = 0.4
	nightStartHour       = 23
	nightEndHour         = 5
	largeOrderThreshold  = 200.0
)

type FareCalculator struct{}

func NewFareCalculator() *FareCalculator {
	return &FareCalculator{}
}

func isNightTime(hour int) bool {
	return hour >= nightStartHour || hour < nightEndHour
}

func (fc *FareCalculator) CalculateEstimatedFare(pickupTime time.Time, distanceKm float64, durationSeconds int64) float64 {
	startHour := pickupTime.Hour()
	estimatedEndTime := pickupTime.Add(time.Duration(durationSeconds) * time.Second)
	endHour := estimatedEndTime.Hour()

	if startHour == endHour {
		return fc.calculateSinglePeriod(isNightTime(startHour), distanceKm)
	}

	distanceSegments := splitTimeAndDistance(startHour, endHour, distanceKm, durationSeconds)
	totalFare := 0.0

	for _, seg := range distanceSegments {
		totalFare += fc.calculateSinglePeriod(seg.isNight, seg.distance)
	}

	return totalFare
}

func (fc *FareCalculator) calculateSinglePeriod(isNight bool, distanceKm float64) float64 {
	baseFare := daytimeBaseFare
	perKmRate := daytimePerKmRate
	if isNight {
		baseFare = nighttimeBaseFare
		perKmRate = nighttimePerKmRate
	}

	fare := baseFare
	if distanceKm > baseDistanceKm {
		extraDistance := distanceKm - baseDistanceKm
		fare += extraDistance * perKmRate
	}

	return fare
}

func (fc *FareCalculator) CalculateActualFare(startTime time.Time, endTime time.Time, totalDistanceKm float64, lowSpeedMinutes int) (float64, bool) {
	totalFare := 0.0
	startHour := startTime.Hour()
	endHour := endTime.Hour()
	totalDuration := endTime.Sub(startTime).Seconds()

	if totalDuration <= 0 {
		totalDuration = 1
	}

	segments := splitTimeAndDistance(startHour, endHour, totalDistanceKm, int64(totalDuration))

	for _, seg := range segments {
		totalFare += fc.calculateSinglePeriod(seg.isNight, seg.distance)
	}

	lowSpeedFee := float64(lowSpeedMinutes) * lowSpeedFeePerMinute
	totalFare += lowSpeedFee

	totalFare = math.Round(totalFare*100) / 100
	isLargeOrder := totalFare > largeOrderThreshold

	return totalFare, isLargeOrder
}

type distanceSegment struct {
	isNight  bool
	distance float64
}

func splitTimeAndDistance(startHour, endHour int, totalDistanceKm float64, totalDurationSeconds int64) []distanceSegment {
	segments := make([]distanceSegment, 0)

	if startHour == endHour {
		segments = append(segments, distanceSegment{
			isNight:  isNightTime(startHour),
			distance: totalDistanceKm,
		})
		return segments
	}

	currentHour := startHour
	elapsedSeconds := 0.0

	for {
		hourStart := currentHour
		hourEnd := currentHour + 1
		if hourEnd >= 24 {
			hourEnd = 0
		}

		hourStartTime := time.Date(2020, 1, 1, hourStart, 0, 0, 0, time.UTC)
		hourEndTime := time.Date(2020, 1, 1, hourEnd, 0, 0, 0, time.UTC)
		hourDuration := hourEndTime.Sub(hourStartTime).Seconds()

		remainingSeconds := float64(totalDurationSeconds) - elapsedSeconds
		segmentSeconds := hourDuration
		if segmentSeconds > remainingSeconds {
			segmentSeconds = remainingSeconds
		}

		segmentRatio := segmentSeconds / float64(totalDurationSeconds)
		segmentDistance := totalDistanceKm * segmentRatio

		segments = append(segments, distanceSegment{
			isNight:  isNightTime(currentHour),
			distance: segmentDistance,
		})

		elapsedSeconds += segmentSeconds
		if elapsedSeconds >= float64(totalDurationSeconds) {
			break
		}

		currentHour = hourEnd
		if currentHour == endHour {
			if elapsedSeconds < float64(totalDurationSeconds) {
				remainingSeconds = float64(totalDurationSeconds) - elapsedSeconds
				segmentRatio = remainingSeconds / float64(totalDurationSeconds)
				segmentDistance = totalDistanceKm * segmentRatio

				segments = append(segments, distanceSegment{
					isNight:  isNightTime(currentHour),
					distance: segmentDistance,
				})
			}
			break
		}
	}

	return segments
}
