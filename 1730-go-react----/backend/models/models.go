package models

import "time"

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleVerifier  Role = "verifier"
	RoleViewer    Role = "viewer"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Role     Role   `json:"role"`
}

type EducationLevel string

const (
	LevelCollege    EducationLevel = "college"
	LevelBachelor   EducationLevel = "bachelor"
	LevelMaster     EducationLevel = "master"
	LevelDoctorate  EducationLevel = "doctorate"
)

type EducationStatus string

const (
	StatusValid   EducationStatus = "valid"
	StatusInvalid EducationStatus = "invalid"
)

type Diploma struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	IDCard          string          `json:"id_card"`
	School          string          `json:"school"`
	Level           EducationLevel  `json:"level"`
	Major           string          `json:"major"`
	StudyYears      int             `json:"study_years"`
	EnrollmentDate  string          `json:"enrollment_date"`
	GraduationDate  string          `json:"graduation_date"`
	Status          EducationStatus `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type OperationType string

const (
	OpCreateDiploma     OperationType = "create_diploma"
	OpUpdateDiploma     OperationType = "update_diploma"
	OpDeleteDiploma     OperationType = "delete_diploma"
	OpVerifyDiploma     OperationType = "verify_diploma"
	OpCreateUser        OperationType = "create_user"
	OpUpdateUser        OperationType = "update_user"
	OpDeleteUser        OperationType = "delete_user"
)

type VerificationResult string

const (
	ResultMatched    VerificationResult = "matched"
	ResultUnmatched  VerificationResult = "unmatched"
)

type OperationLog struct {
	ID              string          `json:"id"`
	OperatorID      string          `json:"operator_id"`
	OperatorName    string          `json:"operator_name"`
	OperatorRole    Role            `json:"operator_role"`
	OperationType   OperationType   `json:"operation_type"`
	Content         string          `json:"content"`
	VerificationResult *VerificationResult `json:"verification_result,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateDiplomaRequest struct {
	Name            string          `json:"name" binding:"required"`
	IDCard          string          `json:"id_card" binding:"required"`
	School          string          `json:"school" binding:"required"`
	Level           EducationLevel  `json:"level" binding:"required"`
	Major           string          `json:"major" binding:"required"`
	StudyYears      int             `json:"study_years" binding:"required,min=1"`
	EnrollmentDate  string          `json:"enrollment_date" binding:"required"`
	GraduationDate  string          `json:"graduation_date" binding:"required"`
}

type UpdateDiplomaRequest struct {
	Name            *string          `json:"name"`
	IDCard          *string          `json:"id_card"`
	School          *string          `json:"school"`
	Level           *EducationLevel  `json:"level"`
	Major           *string          `json:"major"`
	StudyYears      *int             `json:"study_years"`
	EnrollmentDate  *string          `json:"enrollment_date"`
	GraduationDate  *string          `json:"graduation_date"`
	Status          *EducationStatus `json:"status"`
}

type VerifyRequest struct {
	Name   string `json:"name" binding:"required"`
	IDCard string `json:"id_card" binding:"required"`
}

type VerifyResponse struct {
	Diplomas []Diploma `json:"diplomas"`
	Result   string    `json:"result"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     Role   `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	Password *string `json:"password"`
	Role     *Role   `json:"role"`
}

type LogFilter struct {
	OperationType *OperationType
	StartDate     *time.Time
	EndDate       *time.Time
	Page          int
	PageSize      int
}

type PagedResponse struct {
	Data       []OperationLog `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
