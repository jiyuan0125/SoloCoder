package core

import (
	"math"
	"time"

	"insurance-claim/pkg/common"
)

type PayoutCalculator struct{}

func NewPayoutCalculator() *PayoutCalculator {
	return &PayoutCalculator{}
}

func (c *PayoutCalculator) GetLiabilityRatio(liability common.LiabilityType) float64 {
	switch liability {
	case common.LiabilityFull:
		return 1.0
	case common.LiabilityMajor:
		return 0.70
	case common.LiabilityEqual:
		return 0.50
	case common.LiabilityMinor:
		return 0.30
	case common.LiabilityNone:
		return 0.0
	default:
		return 0.0
	}
}

func (c *PayoutCalculator) CalculatePayout(
	assessmentAmount int64,
	policyLimit int64,
	liability common.LiabilityType,
	policyClaimCount int,
	accidentTime time.Time,
	reportTime time.Time,
	hasValidReason bool,
) *common.Payout {
	liabilityRatio := c.GetLiabilityRatio(liability)
	if liabilityRatio == 0 {
		return &common.Payout{
			StandardAmount:   0,
			FinalAmount:      0,
			LiabilityRatio:   0,
			PolicyClaimCount: policyClaimCount,
			CalculatedAt:     time.Now(),
		}
	}

	standardAmount := assessmentAmount
	if standardAmount > policyLimit {
		standardAmount = policyLimit
	}

	hasTimeDiscount := false
	timeDiscountRatio := 1.0
	timeDiff := reportTime.Sub(accidentTime)
	if timeDiff > time.Duration(common.MaxHoursForFullPayout)*time.Hour && !hasValidReason {
		hasTimeDiscount = true
		timeDiscountRatio = common.TimeDiscountRatio
	}

	hasFreqDiscount := false
	freqDiscountRatio := 1.0
	if policyClaimCount > common.MaxClaimsPerYear {
		hasFreqDiscount = true
		excessClaims := policyClaimCount - common.MaxClaimsPerYear
		freqDiscountRatio = math.Pow(common.FreqDiscountRatio, float64(excessClaims))
	}

	rawAmount := float64(standardAmount) * liabilityRatio * timeDiscountRatio * freqDiscountRatio
	finalAmount := int64(math.Round(rawAmount))

	return &common.Payout{
		StandardAmount:    standardAmount,
		FinalAmount:       finalAmount,
		LiabilityRatio:    liabilityRatio,
		PolicyClaimCount:  policyClaimCount,
		HasTimeDiscount:   hasTimeDiscount,
		TimeDiscountRatio: timeDiscountRatio,
		HasFreqDiscount:   hasFreqDiscount,
		FreqDiscountRatio: freqDiscountRatio,
		CalculatedAt:      time.Now(),
	}
}
