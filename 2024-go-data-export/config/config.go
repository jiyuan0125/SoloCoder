package config

const (
	DBPath           = "./data_export.db"
	ExportDir        = "./exports"
	FileRetentionDays = 7
	ServerPort       = "9800"
)

const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

const (
	FormatCSV   = "csv"
	FormatExcel = "excel"
)
