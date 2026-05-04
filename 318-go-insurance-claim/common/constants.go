package common

type PolicyType string

const (
	PolicyTypeMedical  PolicyType = "medical"
	PolicyTypeAuto     PolicyType = "auto"
	PolicyTypeAccident PolicyType = "accident"
)

type ClaimStatus string

const (
	ClaimStatusPending    ClaimStatus = "pending"
	ClaimStatusApproved   ClaimStatus = "approved"
	ClaimStatusRejected   ClaimStatus = "rejected"
	ClaimStatusPaid       ClaimStatus = "paid"
	ClaimStatusProcessing ClaimStatus = "processing"
)

type LiabilityRatio string

const (
	LiabilityFull       LiabilityRatio = "full"
	LiabilityMain       LiabilityRatio = "main"
	LiabilityEqual      LiabilityRatio = "equal"
	LiabilitySecondary  LiabilityRatio = "secondary"
	LiabilityNone       LiabilityRatio = "none"
)

func (l LiabilityRatio) ToPercentage() float64 {
	switch l {
	case LiabilityFull:
		return 1.0
	case LiabilityMain:
		return 0.7
	case LiabilityEqual:
		return 0.5
	case LiabilitySecondary:
		return 0.3
	case LiabilityNone:
		return 0.0
	default:
		return 0.0
	}
}

type PolicyStatus string

const (
	PolicyStatusActive  PolicyStatus = "active"
	PolicyStatusExhausted PolicyStatus = "exhausted"
)

const (
	MaxReasonLength = 500
)
