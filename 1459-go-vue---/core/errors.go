package core

import "errors"

var (
	ErrContractNotFound          = errors.New("contract not found")
	ErrMilestoneNotFound         = errors.New("milestone not found")
	ErrPaymentNotFound           = errors.New("payment not found")
	ErrMilestonePercentageSum    = errors.New("milestone percentages must sum to 100%")
	ErrInvalidPercentage         = errors.New("invalid percentage value")
	ErrInvalidAmount             = errors.New("invalid amount")
	ErrMilestoneAlreadyCompleted = errors.New("milestone already completed")
	ErrMilestoneNotApproved      = errors.New("milestone not approved")
	ErrPaymentAlreadyPaid        = errors.New("payment already paid")
	ErrContractAlreadyCompleted  = errors.New("contract already completed")
)
