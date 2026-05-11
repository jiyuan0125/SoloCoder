package common

import (
	"time"

	"bikeshare/core"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateFenceRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	MinLat   float64 `json:"min_lat"`
	MaxLat   float64 `json:"max_lat"`
	MinLng   float64 `json:"min_lng"`
	MaxLng   float64 `json:"max_lng"`
	Capacity int     `json:"capacity"`
}

type FenceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	MinLat    float64   `json:"min_lat"`
	MaxLat    float64   `json:"max_lat"`
	MinLng    float64   `json:"min_lng"`
	MaxLng    float64   `json:"max_lng"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}

type AddBikeRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type BikeResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	FenceID    string    `json:"fence_id"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type UpdateBikeLocationRequest struct {
	BikeID string  `json:"bike_id"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
}

type StartRideRequest struct {
	BikeID   string  `json:"bike_id"`
	StartLat float64 `json:"start_lat"`
	StartLng float64 `json:"start_lng"`
}

type EndRideRequest struct {
	RideID string  `json:"ride_id"`
	EndLat float64 `json:"end_lat"`
	EndLng float64 `json:"end_lng"`
}

type RideResponse struct {
	ID          string     `json:"id"`
	BikeID      string     `json:"bike_id"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	StartLat    float64    `json:"start_lat"`
	StartLng    float64    `json:"start_lng"`
	EndLat      *float64   `json:"end_lat,omitempty"`
	EndLng      *float64   `json:"end_lng,omitempty"`
	BaseFee     float64    `json:"base_fee,omitempty"`
	DispatchFee float64    `json:"dispatch_fee,omitempty"`
	TotalFee    float64    `json:"total_fee,omitempty"`
}

type DispatchTaskResponse struct {
	ID          string     `json:"id"`
	FromFenceID string     `json:"from_fence_id"`
	ToFenceID   string     `json:"to_fence_id"`
	BikeIDs     []string   `json:"bike_ids"`
	Reason      string     `json:"reason"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type FenceCapacityResponse struct {
	FenceID     string  `json:"fence_id"`
	FenceName   string  `json:"fence_name"`
	Current     int     `json:"current"`
	Capacity    int     `json:"capacity"`
	Ratio       float64 `json:"ratio"`
	Status      string  `json:"status"`
	NeedBikes   int     `json:"need_bikes"`
	ExcessBikes int     `json:"excess_bikes"`
}

func ToFenceResponse(f *core.Fence) *FenceResponse {
	return &FenceResponse{
		ID:        f.ID,
		Name:      f.Name,
		Type:      string(f.Type),
		MinLat:    f.MinLat,
		MaxLat:    f.MaxLat,
		MinLng:    f.MinLng,
		MaxLng:    f.MaxLng,
		Capacity:  f.Capacity,
		CreatedAt: f.CreatedAt,
	}
}

func ToBikeResponse(b *core.Bike) *BikeResponse {
	return &BikeResponse{
		ID:         b.ID,
		Status:     string(b.Status),
		Lat:        b.Lat,
		Lng:        b.Lng,
		FenceID:    b.FenceID,
		LastUsedAt: b.LastUsedAt,
		CreatedAt:  b.CreatedAt,
	}
}

func ToRideResponse(r *core.Ride) *RideResponse {
	return &RideResponse{
		ID:          r.ID,
		BikeID:      r.BikeID,
		StartTime:   r.StartTime,
		EndTime:     r.EndTime,
		StartLat:    r.StartLat,
		StartLng:    r.StartLng,
		EndLat:      r.EndLat,
		EndLng:      r.EndLng,
		BaseFee:     r.BaseFee,
		DispatchFee: r.DispatchFee,
		TotalFee:    r.TotalFee,
	}
}

func ToDispatchTaskResponse(t *core.DispatchTask) *DispatchTaskResponse {
	return &DispatchTaskResponse{
		ID:          t.ID,
		FromFenceID: t.FromFenceID,
		ToFenceID:   t.ToFenceID,
		BikeIDs:     t.BikeIDs,
		Reason:      t.Reason,
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
	}
}

func ToFenceCapacityResponse(s core.FenceCapacityStatus) *FenceCapacityResponse {
	return &FenceCapacityResponse{
		FenceID:     s.FenceID,
		FenceName:   s.FenceName,
		Current:     s.Current,
		Capacity:    s.Capacity,
		Ratio:       s.Ratio,
		Status:      s.Status,
		NeedBikes:   s.NeedBikes,
		ExcessBikes: s.ExcessBikes,
	}
}
