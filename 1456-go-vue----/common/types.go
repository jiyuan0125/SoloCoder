package common

type ArchiveCategory string
type ArchiveSecrecyLevel string
type BorrowStatus string
type ApprovalStatus string
type UserRole string

const (
	ArchiveCategoryPersonnel   ArchiveCategory = "人事"
	ArchiveCategoryFinance     ArchiveCategory = "财务"
	ArchiveCategoryContract    ArchiveCategory = "合同"
	ArchiveCategoryTechDoc     ArchiveCategory = "技术文档"

	ArchiveSecrecyLevelPublic   ArchiveSecrecyLevel = "公开"
	ArchiveSecrecyLevelInternal ArchiveSecrecyLevel = "内部"
	ArchiveSecrecyLevelConf     ArchiveSecrecyLevel = "机密"
	ArchiveSecrecyLevelTopSecret ArchiveSecrecyLevel = "绝密"

	BorrowStatusPending    BorrowStatus = "待审批"
	BorrowStatusApproved   BorrowStatus = "已通过"
	BorrowStatusRejected   BorrowStatus = "已拒绝"
	BorrowStatusBorrowed   BorrowStatus = "借阅中"
	BorrowStatusOverdue    BorrowStatus = "逾期"
	BorrowStatusReturned   BorrowStatus = "已归还"

	ApprovalStatusPending  ApprovalStatus = "待审批"
	ApprovalStatusApproved ApprovalStatus = "已通过"
	ApprovalStatusRejected ApprovalStatus = "已拒绝"

	UserRoleAdmin    UserRole = "管理员"
	UserRoleManager  UserRole = "部门经理"
	UserRoleEmployee UserRole = "普通员工"
)

type Archive struct {
	ArchiveID       string               `json:"archive_id"`
	Title           string               `json:"title"`
	Category        ArchiveCategory      `json:"category"`
	SecrecyLevel    ArchiveSecrecyLevel  `json:"secrecy_level"`
	ArchiveDate     string               `json:"archive_date"`
	Archiver        string               `json:"archiver"`
	IsDestroyed     bool                 `json:"is_destroyed"`
}

type BorrowRecord struct {
	ID              int64         `json:"id"`
	ArchiveID       string        `json:"archive_id"`
	Applicant       string        `json:"applicant"`
	Reason          string        `json:"reason"`
	ExpectedReturn  string        `json:"expected_return"`
	ActualReturn    string        `json:"actual_return,omitempty"`
	Status          BorrowStatus  `json:"status"`
	ManagerApproval ApprovalStatus `json:"manager_approval"`
	AdminApproval   ApprovalStatus `json:"admin_approval"`
	ApplyDate       string        `json:"apply_date"`
}

type DestroyRecord struct {
	ID          int64  `json:"id"`
	ArchiveID   string `json:"archive_id"`
	Title       string `json:"title"`
	Destroyer   string `json:"destroyer"`
	DestroyDate string `json:"destroy_date"`
}
