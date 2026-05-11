package common

type CreateArchiveRequest struct {
	Title        string              `json:"title"`
	Category     ArchiveCategory     `json:"category"`
	SecrecyLevel ArchiveSecrecyLevel `json:"secrecy_level"`
	ArchiveDate  string              `json:"archive_date"`
	Archiver     string              `json:"archiver"`
}

type ListArchivesRequest struct {
	Keyword  string             `json:"keyword,omitempty"`
	Category ArchiveCategory    `json:"category,omitempty"`
	UserRole UserRole           `json:"user_role"`
}

type GetArchiveRequest struct {
	ArchiveID string   `json:"archive_id"`
	UserRole  UserRole `json:"user_role"`
}

type ApplyBorrowRequest struct {
	ArchiveID      string `json:"archive_id"`
	Applicant      string `json:"applicant"`
	Reason         string `json:"reason"`
	ExpectedReturn string `json:"expected_return"`
}

type ApproveBorrowRequest struct {
	BorrowID   int64    `json:"borrow_id"`
	Approver   string   `json:"approver"`
	UserRole   UserRole `json:"user_role"`
	IsApproved bool     `json:"is_approved"`
}

type ReturnArchiveRequest struct {
	BorrowID  int64  `json:"borrow_id"`
	Returner  string `json:"returner"`
}

type DestroyArchiveRequest struct {
	ArchiveID string `json:"archive_id"`
	Destroyer string `json:"destroyer"`
}

type ListPendingDestroyRequest struct {
	CurrentDate string `json:"current_date,omitempty"`
}
