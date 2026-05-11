package common

type Severity string

const (
	SeverityGeneral Severity = "general"
	SeveritySerious Severity = "serious"
	SeverityMajor   Severity = "major"
)

type AcceptanceStatus string

const (
	AcceptanceStatusPending   AcceptanceStatus = "pending"
	AcceptanceStatusPassed    AcceptanceStatus = "passed"
	AcceptanceStatusFailed    AcceptanceStatus = "failed"
	AcceptanceStatusRectified AcceptanceStatus = "rectified"
	AcceptanceStatusReported  AcceptanceStatus = "reported"
)

type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusResolved   IssueStatus = "resolved"
	IssueStatusClosed     IssueStatus = "closed"
	IssueStatusOverdue    IssueStatus = "overdue"
)
