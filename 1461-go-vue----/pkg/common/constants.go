package common

type BatchStatus string
type ProcessResult string
type InspectionType string
type InspectionResult string

const (
	BatchStatusCreated    BatchStatus = "created"
	BatchStatusInProgress BatchStatus = "in_progress"
	BatchStatusPaused     BatchStatus = "paused"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusAbnormal   BatchStatus = "abnormal"
	BatchStatusScrapped   BatchStatus = "scrapped"
)

const (
	ProcessResultPass ProcessResult = "pass"
	ProcessResultFail ProcessResult = "fail"
)

const (
	InspectionTypeIncoming InspectionType = "incoming"
	InspectionTypeProcess  InspectionType = "process"
	InspectionTypeFinal    InspectionType = "final"
)

const (
	InspectionResultPass InspectionResult = "pass"
	InspectionResultFail InspectionResult = "fail"
)

const MaxReworkCount = 2
const DeviationThreshold = 0.10
