package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ArchiveSummary struct {
	ArchiveID string `json:"archive_id"`
	Title     string `json:"title"`
}

type ListArchivesResponse struct {
	Archives []Archive `json:"archives"`
}

type GetArchiveResponse struct {
	Archive Archive `json:"archive"`
}

type CreateArchiveResponse struct {
	ArchiveID string `json:"archive_id"`
}

type BorrowRecordResponse struct {
	Records []BorrowRecord `json:"records"`
}

type DestroyRecordResponse struct {
	Records []DestroyRecord `json:"records"`
}

type PendingDestroyResponse struct {
	Archives []Archive `json:"archives"`
}

type OverdueReminder struct {
	BorrowID       int64  `json:"borrow_id"`
	ArchiveID      string `json:"archive_id"`
	ArchiveTitle   string `json:"archive_title"`
	Applicant      string `json:"applicant"`
	ExpectedReturn string `json:"expected_return"`
	OverdueDays    int    `json:"overdue_days"`
}

type OverdueRemindersResponse struct {
	Reminders []OverdueReminder `json:"reminders"`
}
