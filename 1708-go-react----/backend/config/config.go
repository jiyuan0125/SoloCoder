package config

import "os"

func GetPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func GetDBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "./blood.db"
}

const (
	ABO_A     = "A"
	ABO_B     = "B"
	ABO_AB    = "AB"
	ABO_O     = "O"
	RhPositive = "Positive"
	RhNegative = "Negative"

	CollectionWhole200   = "Whole_200ml"
	CollectionWhole400   = "Whole_400ml"
	CollectionPlatelets  = "Platelets"

	StatusPendingTest    = "Pending_Test"
	StatusTesting        = "Testing"
	StatusQualified      = "Qualified"
	StatusDisqualified   = "Disqualified"
	StatusScrapped       = "Scrapped"
	StatusPendingStorage = "Pending_Storage"
	StatusInStock        = "In_Stock"
	StatusFrozen         = "Frozen"
	StatusIssued         = "Issued"
	StatusExpired        = "Expired"

	ProductWholeBlood     = "Whole_Blood"
	ProductRBC            = "Red_Blood_Cells"
	ProductFFP            = "Fresh_Frozen_Plasma"
	ProductPlatelets      = "Platelets"

	ResultNegative = "Negative"
	ResultPositive = "Positive"
	ResultInvalid  = "Invalid"

	TestRound1 = 1
	TestRound2 = 2
	TestRound3 = 3

	UrgencyRegular = "Regular"
	UrgencyUrgent  = "Urgent"
	UrgencySpecial = "Special_Urgent"

	ReqStatusPending  = "Pending"
	ReqStatusMatched  = "Matched"
	ReqStatusQueued   = "Queued"
	ReqStatusIssued   = "Issued"
	ReqStatusRejected = "Rejected"

	StatusDraft        = "Draft"
	StatusSubmitted    = "Submitted"
	StatusApproved     = "Approved"
	StatusPublished    = "Published"
	StatusRejected     = "Rejected"
)

var BloodTypeCombinations = []string{
	"A_Positive", "A_Negative",
	"B_Positive", "B_Negative",
	"AB_Positive", "AB_Negative",
	"O_Positive", "O_Negative",
}

var ProductTypes = []string{
	ProductWholeBlood,
	ProductRBC,
	ProductFFP,
	ProductPlatelets,
}

var TestItems = []string{
	"ABO_Front",
	"ABO_Back",
	"RhD",
	"ALT",
	"HBsAg",
	"Anti_HCV",
	"Anti_HIV",
	"Anti_Syphilis",
	"NAT",
}
