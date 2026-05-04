package common

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	MaxPurchaseLimit = 10
	MaxEventNameLength = 100
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateEventID() string {
	timestamp := time.Now().UnixMilli()
	random := rand.Intn(10000)
	return fmt.Sprintf("EVT%d%04d", timestamp, random)
}

func GenerateTicketNumber() string {
	timestamp := time.Now().UnixMilli()
	random := rand.Intn(1000000)
	return fmt.Sprintf("TKT%d%06d", timestamp, random)
}

func ValidateCreateEventRequest(req *CreateEventRequest) error {
	if req.Name == "" {
		return ErrEventNameEmpty
	}
	if len(req.Name) > MaxEventNameLength {
		return ErrEventNameTooLong
	}
	if req.Location == "" {
		return ErrLocationEmpty
	}
	if req.Time.Before(time.Now()) {
		return ErrInvalidTime
	}
	if len(req.Tiers) == 0 {
		return ErrNoTiers
	}
	for _, tier := range req.Tiers {
		if tier.Name == "" {
			return ErrTierNameEmpty
		}
		if tier.Price <= 0 {
			return ErrInvalidPrice
		}
		if tier.Capacity <= 0 {
			return ErrInvalidCapacity
		}
	}
	return nil
}

func ValidatePurchaseRequest(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if quantity > MaxPurchaseLimit {
		return ErrPurchaseLimit
	}
	return nil
}
