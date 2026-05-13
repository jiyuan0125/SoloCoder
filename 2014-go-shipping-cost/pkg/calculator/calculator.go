package calculator

import (
	"math"

	"shipping-cost/pkg/address"
	"shipping-cost/pkg/config"
)

type Input struct {
	Sender   string
	Receiver string
	Weight   float64
	Length   float64
	Width    float64
	Height   float64
}

type Result struct {
	Zone         string
	Weight       float64
	VolumeWeight float64
	ChargeWeight float64
	FirstWeight  int
	ContinueWeight int
	PricePerKG   int
	TotalFen     int
	TotalYuan    float64
}

func Calculate(in *Input) (*Result, error) {
	senderAddr, err := address.Resolve(in.Sender)
	if err != nil {
		return nil, err
	}

	receiverAddr, err := address.Resolve(in.Receiver)
	if err != nil {
		return nil, err
	}

	zone := address.DetermineZone(senderAddr, receiverAddr)

	volumeWeight := (in.Length * in.Width * in.Height) / 6000.0
	chargeWeight := in.Weight
	if volumeWeight > in.Weight {
		chargeWeight = volumeWeight
	}

	pricePerKG := getPricePerKG(zone)

	var firstWeight int
	var continueWeight int

	if chargeWeight <= 1.0 {
		firstWeight = 1
		continueWeight = 0
	} else {
		firstWeight = 1
		remaining := chargeWeight - 1.0
		continueUnits := math.Ceil(remaining / 0.5)
		continueWeight = int(continueUnits)
	}

	firstPriceFen := pricePerKG * firstWeight
	continuePriceFen := pricePerKG / 2 * continueWeight
	totalFen := firstPriceFen + continuePriceFen

	result := &Result{
		Zone:         zone.String(),
		Weight:       in.Weight,
		VolumeWeight: volumeWeight,
		ChargeWeight: chargeWeight,
		FirstWeight:  firstWeight,
		ContinueWeight: continueWeight,
		PricePerKG:   pricePerKG,
		TotalFen:     totalFen,
		TotalYuan:    float64(totalFen) / 100.0,
	}

	return result, nil
}

func getPricePerKG(zone address.Zone) int {
	cfg := config.Get()
	switch zone {
	case address.ZoneLocal:
		return cfg.ZoneRules.Local
	case address.ZoneIntra:
		return cfg.ZoneRules.Intra
	case address.ZoneNeighbor:
		return cfg.ZoneRules.Neighbor
	case address.ZoneInter:
		return cfg.ZoneRules.Inter
	case address.ZoneRemote:
		return cfg.ZoneRules.Remote
	default:
		return cfg.ZoneRules.Inter
	}
}
