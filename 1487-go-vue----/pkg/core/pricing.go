package core

import "math"

const (
	BaseFeeCents         int64 = 28000
	DistanceFactor       float64 = 1.4
	FreeDistanceKM       float64 = 10.0
	MediumDistanceRate   int64 = 500
	LongDistanceRate     int64 = 800
	MediumDistanceMax    float64 = 30.0
	PianoFeeCents        int64 = 20000
	DisassemblyFeeCents  int64 = 5000
	BoxVolumeM3          float64 = 0.06
)

var furnitureVolume = map[ItemSize]float64{
	SizeSmall:  0.5,
	SizeMedium: 1.0,
	SizeLarge:  2.0,
}

func EstimatePrice(items ItemList, distanceKM float64) int64 {
	total := BaseFeeCents
	total += calcDistanceFee(distanceKM)
	total += calcSpecialItemFee(items)
	return total
}

func calcDistanceFee(distanceKM float64) int64 {
	actualDistance := distanceKM * DistanceFactor
	var fee int64 = 0

	if actualDistance <= FreeDistanceKM {
		return 0
	}

	mediumDistance := math.Min(actualDistance-FreeDistanceKM, MediumDistanceMax-FreeDistanceKM)
	fee += int64(mediumDistance * float64(MediumDistanceRate))

	if actualDistance > MediumDistanceMax {
		longDistance := actualDistance - MediumDistanceMax
		fee += int64(longDistance * float64(LongDistanceRate))
	}

	return fee
}

func calcSpecialItemFee(items ItemList) int64 {
	var fee int64 = 0

	for _, appliance := range items.Appliances {
		if appliance.NeedDisassemble {
			fee += DisassemblyFeeCents
		}
	}

	for _, special := range items.Specials {
		if special == "钢琴" {
			fee += PianoFeeCents
		}
	}

	return fee
}

func CalculateCancellationFee(estimate int64, hoursUntilMove float64) int64 {
	if hoursUntilMove >= 24 {
		return 0
	}

	fee := int64(float64(estimate) * 0.3)
	const minCancellationFee int64 = 5000
	if fee < minCancellationFee {
		fee = minCancellationFee
	}

	return fee
}
