package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"obd-platform/pkg/api"
)

type OBDRecord struct {
	DeviceID          string
	Timestamp         time.Time
	Latitude          float64
	Longitude         float64
	Speed             float64
	EngineRPM         int
	FuelConsumption   float64
	HardAccelerations int
	HardBrakes        int
	HardTurns         int
	PrevSpeed         float64
	IsSevereBrake     bool
}

func newRecordFromAPI(p *api.OBDPacket) *OBDRecord {
	return &OBDRecord{
		DeviceID:          p.DeviceID,
		Timestamp:         p.Timestamp,
		Latitude:          p.Latitude,
		Longitude:         p.Longitude,
		Speed:             p.Speed,
		EngineRPM:         p.EngineRPM,
		FuelConsumption:   p.FuelConsumption,
		HardAccelerations: p.HardAccelerations,
		HardBrakes:        p.HardBrakes,
		HardTurns:         p.HardTurns,
	}
}

func recordToAPIDailyStats(r *DailyVehicleStats) *api.VehicleDailyStats {
	return &api.VehicleDailyStats{
		Date:        r.Date,
		DeviceID:    r.DeviceID,
		Mileage:     r.TotalMileage,
		AvgFuel:     r.AvgFuelConsumption,
		SafetyScore: r.SafetyScore,
		TotalFuel:   r.TotalFuel,
	}
}

func recordsToFleetStats(r *DailyFleetStats) *api.FleetDailyStats {
	return &api.FleetDailyStats{
		Date:               r.Date,
		TotalMileage:       r.TotalMileage,
		AvgFuelConsumption: r.AvgFuelConsumption,
		AbnormalEvents:     r.TotalAbnormalEvents,
	}
}

type DailyVehicleStats struct {
	Date                time.Time
	DeviceID            string
	TotalMileage        float64
	TotalFuel           float64
	AvgFuelConsumption  float64
	HardAccelerations   int
	HardBrakes          int
	HardTurns           int
	SevereBrakes        int
	SafetyScore         int
	Packets             int
	LastRecord          *OBDRecord
	HasData             bool
}

type DailyFleetStats struct {
	Date                   time.Time
	TotalMileage           float64
	AvgFuelConsumption     float64
	TotalAbnormalEvents    int
	TotalVehicles          int
}

type AlertType string

const (
	AlertTypeSevereBrake   AlertType = "severe_brake"
	AlertTypeHighBraking   AlertType = "high_braking"
	AlertTypeDrivingHabit  AlertType = "driving_habit"
)

const (
	AlertStatusPending    = "pending"
	AlertStatusResolved   = "resolved"
)

type Alert struct {
	ID        string
	DeviceID  string
	Type      AlertType
	Date      time.Time
	Message   string
	Status    string
	CreatedAt time.Time
}

func (a *Alert) toAPI() *api.Alert {
	return &api.Alert{
		ID:        a.ID,
		DeviceID:  a.DeviceID,
		Type:      string(a.Type),
		Date:      a.Date,
		Message:   a.Message,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
	}
}

func recordHash(deviceID string, t time.Time) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d", deviceID, t.Unix())))
	return hex.EncodeToString(h[:])
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func isSameDay(t1, t2 time.Time) bool {
	return truncateToDate(t1).Equal(truncateToDate(t2))
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func distance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	radLat1 := lat1 * math.Pi / 180
	radLat2 := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func calculateSafetyScore(hardAccels, hardBrakes, hardTurns int) int {
	score := 100
	score -= hardAccels * 3
	score -= hardBrakes * 5
	score -= hardTurns * 2
	if score < 0 {
		score = 0
	}
	return score
}
