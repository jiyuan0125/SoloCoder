package common

const (
	OperationTypeCreate   = "create"
	OperationTypeUpdate   = "update"
	OperationTypeDelete   = "delete"
	OperationTypeLogin    = "login"
	OperationTypeLogout   = "logout"
	OperationTypeExport   = "export"
	OperationTypeImport   = "import"
	OperationTypeQuery    = "query"
)

const (
	ResultSuccess = "success"
	ResultFailed  = "failed"
)

const (
	MaxOperationsPerMinute   = 100
	MaxLoginFailures         = 5
	LockDurationMinutes      = 30
	RetentionDays            = 90
	StorageAlertThreshold    = 0.8
)

var SensitiveOperations = map[string]bool{
	OperationTypeExport: true,
	OperationTypeDelete: true,
}

var RequireApprovalOperations = map[string]bool{
	OperationTypeExport: true,
}
