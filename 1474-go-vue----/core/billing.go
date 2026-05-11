package core

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	InitialFee            = 1.50
	PerFifteenMinutesFee  = 1.00
	DailyMaxFee           = 25.00
	ForbiddenDispatchFee  = 5.00
	OtherDispatchFee      = 2.00
)

type BillingService struct {
	store *Store
	mu    sync.Mutex
}

func NewBillingService(store *Store) *BillingService {
	return &BillingService{store: store}
}

func (bs *BillingService) StartRide(bikeID string, startLat, startLng float64) (*Ride, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bike := bs.store.GetBike(bikeID)
	if bike == nil {
		return nil, ErrBikeNotFound
	}
	if bike.Status != BikeStatusIdle {
		return nil, ErrBikeNotIdle
	}

	bike.Status = BikeStatusInUse
	bike.Lat = startLat
	bike.Lng = startLng
	bike.LastUsedAt = time.Now()
	bs.store.saveBike(bike)

	ride := &Ride{
		ID:        uuid.New().String(),
		BikeID:    bikeID,
		StartTime: time.Now(),
		StartLat:  startLat,
		StartLng:  startLng,
	}
	bs.store.saveRide(ride)

	return ride, nil
}

func (bs *BillingService) EndRide(rideID string, endLat, endLng float64) (*Ride, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	ride := bs.store.GetRide(rideID)
	if ride == nil {
		return nil, errors.New("ride not found")
	}
	if ride.EndTime != nil {
		return nil, errors.New("ride already ended")
	}

	bike := bs.store.GetBike(ride.BikeID)
	if bike == nil {
		return nil, ErrBikeNotFound
	}

	endTime := time.Now()
	ride.EndTime = &endTime
	ride.EndLat = &endLat
	ride.EndLng = &endLng

	duration := endTime.Sub(ride.StartTime)
	minutes := int(duration.Minutes())
	if minutes < 0 {
		minutes = 0
	}

	ride.BaseFee = calculateBaseFee(minutes)
	ride.DispatchFee = calculateDispatchFee(bs.store, endLat, endLng)
	ride.TotalFee = roundToCents(ride.BaseFee + ride.DispatchFee)

	bike.Status = BikeStatusIdle
	bike.Lat = endLat
	bike.Lng = endLng
	bike.LastUsedAt = time.Now()

	fence := bs.store.GetFenceByPoint(endLat, endLng)
	if fence != nil {
		bike.FenceID = fence.ID
	} else {
		bike.FenceID = ""
	}

	bs.store.saveBike(bike)
	bs.store.saveRide(ride)

	return ride, nil
}

func calculateBaseFee(minutes int) float64 {
	if minutes <= 15 {
		return InitialFee
	}

	extraMinutes := minutes - 15
	extraBlocks := int(math.Ceil(float64(extraMinutes) / 15.0))
	fee := InitialFee + float64(extraBlocks)*PerFifteenMinutesFee

	if fee > DailyMaxFee {
		return DailyMaxFee
	}
	return roundToCents(fee)
}

func calculateDispatchFee(store *Store, lat, lng float64) float64 {
	fence := store.GetFenceByPoint(lat, lng)

	if fence == nil {
		return OtherDispatchFee
	}

	switch fence.Type {
	case FenceTypeForbidden:
		return ForbiddenDispatchFee
	case FenceTypeRecommended:
		return 0.00
	default:
		return OtherDispatchFee
	}
}

func roundToCents(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func (bs *BillingService) GetActiveRides() []*Ride {
	rides := bs.store.GetRides()
	result := make([]*Ride, 0)
	for _, r := range rides {
		if r.EndTime == nil {
			result = append(result, r)
		}
	}
	return result
}
