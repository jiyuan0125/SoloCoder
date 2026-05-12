package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProtocolStatus string

const (
	ProtocolPreparing ProtocolStatus = "筹备中"
	ProtocolOngoing   ProtocolStatus = "进行中"
	ProtocolCompleted ProtocolStatus = "已完成"
	ProtocolTerminated ProtocolStatus = "已终止"
)

type TrialPhase string

const (
	PhaseI   TrialPhase = "I期"
	PhaseII  TrialPhase = "II期"
	PhaseIII TrialPhase = "III期"
	PhaseIV  TrialPhase = "IV期"
)

type SubjectStatus string

const (
	SubjectScreening SubjectStatus = "筛选中"
	SubjectEnrolled  SubjectStatus = "入组"
	SubjectTreating  SubjectStatus = "治疗中"
	SubjectFollowup  SubjectStatus = "随访中"
	SubjectCompleted SubjectStatus = "已完成"
	SubjectWithdrawn SubjectStatus = "退出"
)

type Protocol struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProtocolNumber      string         `gorm:"uniqueIndex;not null" json:"protocol_number"`
	DrugName            string         `gorm:"not null" json:"drug_name"`
	Indication          string         `gorm:"not null" json:"indication"`
	TrialPhase          TrialPhase     `gorm:"not null" json:"trial_phase"`
	PlannedEnrollment   int            `gorm:"not null" json:"planned_enrollment"`
	StartDate           time.Time      `gorm:"not null" json:"start_date"`
	EndDate             time.Time      `gorm:"not null" json:"end_date"`
	InclusionCriteria   string         `gorm:"type:text" json:"inclusion_criteria"`
	ExclusionCriteria   string         `gorm:"type:text" json:"exclusion_criteria"`
	Status              ProtocolStatus `gorm:"not null" json:"status"`
	GroupRatio          string         `gorm:"not null;default:'1:1'" json:"group_ratio"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`

	Sites      []Site      `gorm:"foreignKey:ProtocolID;constraint:OnDelete:CASCADE" json:"sites,omitempty"`
	Visits     []Visit     `gorm:"foreignKey:ProtocolID;constraint:OnDelete:CASCADE" json:"visits,omitempty"`
	Subjects   []Subject   `gorm:"foreignKey:ProtocolID" json:"subjects,omitempty"`
}

func (p *Protocol) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Site struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProtocolID        uuid.UUID `gorm:"type:uuid;index;not null" json:"protocol_id"`
	SiteCode          string    `gorm:"size:10;not null" json:"site_code"`
	SiteName          string    `gorm:"not null" json:"site_name"`
	PrincipalInvestigator string `gorm:"not null" json:"principal_investigator"`
	PlannedEnrollment int       `gorm:"not null" json:"planned_enrollment"`
	CreatedAt         time.Time `json:"created_at"`

	Protocol *Protocol `gorm:"foreignKey:ProtocolID" json:"-"`
	Subjects []Subject `gorm:"foreignKey:SiteID" json:"-"`
}

func (s *Site) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type Visit struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProtocolID    uuid.UUID `gorm:"type:uuid;index;not null" json:"protocol_id"`
	VisitName     string    `gorm:"not null" json:"visit_name"`
	VisitOrder    int       `gorm:"not null" json:"visit_order"`
	WindowDays    int       `gorm:"not null;default:0" json:"window_days"`
	WindowTolerance int     `gorm:"not null;default:0" json:"window_tolerance"`
	CreatedAt     time.Time `json:"created_at"`

	Protocol *Protocol `gorm:"foreignKey:ProtocolID" json:"-"`
	Records  []VisitRecord `gorm:"foreignKey:VisitID" json:"-"`
}

func (v *Visit) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type Subject struct {
	ID                uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	ProtocolID        uuid.UUID     `gorm:"type:uuid;index;not null" json:"protocol_id"`
	SiteID            uuid.UUID     `gorm:"type:uuid;index;not null" json:"site_id"`
	ScreeningNumber   string        `gorm:"not null" json:"screening_number"`
	RandomizationID   string        `gorm:"uniqueIndex;not null" json:"randomization_id"`
	NameInitials      string        `gorm:"not null" json:"name_initials"`
	Gender            string        `gorm:"not null" json:"gender"`
	BirthDate         time.Time     `gorm:"not null" json:"birth_date"`
	EnrollmentDate    *time.Time    `json:"enrollment_date"`
	GroupAssignment   string        `json:"group_assignment"`
	Status            SubjectStatus `gorm:"not null" json:"status"`
	SeqNumber         int           `gorm:"not null;default:0" json:"seq_number"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`

	Protocol *Protocol   `gorm:"foreignKey:ProtocolID" json:"-"`
	Site     *Site       `gorm:"foreignKey:SiteID" json:"-"`
	Records  []VisitRecord `gorm:"foreignKey:SubjectID" json:"-"`
	AEs      []AdverseEvent `gorm:"foreignKey:SubjectID" json:"-"`
}

func (s *Subject) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type VisitRecord struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SubjectID     uuid.UUID `gorm:"type:uuid;index:idx_subject_visit,unique;not null" json:"subject_id"`
	VisitID       uuid.UUID `gorm:"type:uuid;index:idx_subject_visit,unique;not null" json:"visit_id"`
	ActualDate    time.Time `gorm:"not null" json:"actual_date"`
	IsOutOfWindow bool      `gorm:"not null;default:false" json:"is_out_of_window"`
	VitalSigns    string    `gorm:"type:text" json:"vital_signs"`
	LabTests      string    `gorm:"type:text" json:"lab_tests"`
	OtherData     string    `gorm:"type:text" json:"other_data"`
	CreatedAt     time.Time `json:"created_at"`

	Subject *Subject `gorm:"foreignKey:SubjectID" json:"-"`
	Visit   *Visit   `gorm:"foreignKey:VisitID" json:"-"`
}

func (r *VisitRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type AESeverity string

const (
	AEMild     AESeverity = "轻度"
	AEModerate AESeverity = "中度"
	AEComplete AESeverity = "重度"
)

type AERelationship string

const (
	RelDefinitely AERelationship = "肯定有关"
	RelProbably   AERelationship = "可能有关"
	RelPossibly   AERelationship = "可能无关"
	RelUnrelated  AERelationship = "无关"
)

type AEOutcome string

const (
	OutcomeRecovered AEOutcome = "恢复"
	OutcomeImproved  AEOutcome = "好转"
	OutcomeOngoing   AEOutcome = "未恢复"
	OutcomeDeath     AEOutcome = "死亡"
	OutcomeWithdrawn AEOutcome = "导致退出试验"
)

type AdverseEvent struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	SubjectID        uuid.UUID       `gorm:"type:uuid;index;not null" json:"subject_id"`
	EventName        string          `gorm:"not null" json:"event_name"`
	StartDate        time.Time       `gorm:"not null" json:"start_date"`
	EndDate          *time.Time      `json:"end_date"`
	Severity         AESeverity      `gorm:"not null" json:"severity"`
	Relationship     AERelationship  `gorm:"not null" json:"relationship"`
	IsSAE            bool            `gorm:"not null;default:false" json:"is_sae"`
	Treatment        string          `gorm:"type:text" json:"treatment"`
	Outcome          AEOutcome       `json:"outcome"`
	ReportSubmitted  bool            `gorm:"not null;default:false" json:"report_submitted"`
	ReportDeadline   *time.Time      `json:"report_deadline"`
	CreatedAt        time.Time       `json:"created_at"`

	Subject *Subject `gorm:"foreignKey:SubjectID" json:"-"`
}

func (a *AdverseEvent) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type TodoStatus string

const (
	TodoPending    TodoStatus = "待办"
	TodoCompleted  TodoStatus = "已完成"
	TodoOverdue    TodoStatus = "逾期"
)

type Todo struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	AeID        uuid.UUID  `gorm:"type:uuid;index" json:"ae_id"`
	Status      TodoStatus `gorm:"not null" json:"status"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (t *Todo) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type Department struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

func (d *Department) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type Budget struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DepartmentID  uuid.UUID `gorm:"type:uuid;index;not null" json:"department_id"`
	Period        string    `gorm:"not null" json:"period"`
	TotalAmount   float64   `gorm:"not null" json:"total_amount"`
	UsedAmount    float64   `gorm:"not null;default:0" json:"used_amount"`
	MonthlyUsed   float64   `gorm:"not null;default:0" json:"monthly_used"`
	CreatedAt     time.Time `json:"created_at"`

	Department *Department `gorm:"foreignKey:DepartmentID" json:"-"`
	Items      []BudgetItem `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BudgetItemStatus string

const (
	ItemPending   BudgetItemStatus = "待审批"
	ItemApproved  BudgetItemStatus = "已通过"
	ItemRejected  BudgetItemStatus = "已拒绝"
	ItemCompleted BudgetItemStatus = "已完成"
)

type BudgetItem struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	BudgetID    uuid.UUID        `gorm:"type:uuid;index;not null" json:"budget_id"`
	Name        string           `gorm:"not null" json:"name"`
	Amount      float64          `gorm:"not null" json:"amount"`
	UsedAmount  float64          `gorm:"not null;default:0" json:"used_amount"`
	Status      BudgetItemStatus `gorm:"not null;default:待审批" json:"status"`
	CreatedAt   time.Time        `json:"created_at"`

	Budget *Budget `gorm:"foreignKey:BudgetID" json:"-"`
}

func (i *BudgetItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

type BudgetAlert struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	BudgetID     uuid.UUID `gorm:"type:uuid;index;not null" json:"budget_id"`
	DepartmentID uuid.UUID `gorm:"type:uuid;index;not null" json:"department_id"`
	Message      string    `gorm:"not null" json:"message"`
	AlertType    string    `gorm:"not null" json:"alert_type"`
	CreatedAt    time.Time `json:"created_at"`
}

func (a *BudgetAlert) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type SystemA struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name       string    `gorm:"not null" json:"name"`
	ConfigInfo string    `gorm:"type:text" json:"config_info"`
	CreatedAt  time.Time `json:"created_at"`

	Items []SystemB `gorm:"foreignKey:SystemAID" json:"items,omitempty"`
}

func (a *SystemA) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type SystemB struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SystemAID uuid.UUID `gorm:"type:uuid;index;not null" json:"system_a_id"`
	Name      string    `gorm:"not null" json:"name"`
	Details   string    `gorm:"type:text" json:"details"`
	CreatedAt time.Time `json:"created_at"`

	SystemA *SystemA `gorm:"foreignKey:SystemAID" json:"-"`
}

func (b *SystemB) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
