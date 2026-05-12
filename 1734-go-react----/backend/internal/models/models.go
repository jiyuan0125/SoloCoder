package models

import "time"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleTeacher  Role = "teacher"
	RoleParent   Role = "parent"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Role      Role      `json:"role"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Parent struct {
	UserID   string `json:"user_id"`
	ChildIDs []string `json:"child_ids"`
}

type Student struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Grade       string `json:"grade"`
	Class       string `json:"class"`
	StudentNo   string `json:"student_no"`
}

type Teacher struct {
	UserID      string   `json:"user_id"`
	Subject     string   `json:"subject"`
	ClassIDs    []string `json:"class_ids"`
}

type ScopeType string

const (
	ScopeAll    ScopeType = "all"
	ScopeGrade  ScopeType = "grade"
	ScopeClass  ScopeType = "class"
)

type Urgency string

const (
	UrgencyNormal  Urgency = "normal"
	UrgencyImportant Urgency = "important"
	UrgencyUrgent  Urgency = "urgent"
)

type Announcement struct {
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	ScopeType    ScopeType   `json:"scope_type"`
	ScopeValue   string      `json:"scope_value"`
	Urgency      Urgency     `json:"urgency"`
	PublishedBy  string      `json:"published_by"`
	PublishedAt  time.Time   `json:"published_at"`
}

type AnnouncementRead struct {
	AnnouncementID string    `json:"announcement_id"`
	ParentID       string    `json:"parent_id"`
	ReadAt         time.Time `json:"read_at"`
}

type AnnouncementStats struct {
	AnnouncementID string `json:"announcement_id"`
	TotalParents   int    `json:"total_parents"`
	ReadCount      int    `json:"read_count"`
	UnreadCount    int    `json:"unread_count"`
}

type ExamScore struct {
	ID            string    `json:"id"`
	StudentID     string    `json:"student_id"`
	Grade         string    `json:"grade"`
	Class         string    `json:"class"`
	Subject       string    `json:"subject"`
	Semester      string    `json:"semester"`
	Score         *float64  `json:"score"`
	ExamDate      time.Time `json:"exam_date"`
	RecordedBy    string    `json:"recorded_by"`
	RecordedAt    time.Time `json:"recorded_at"`
}

type ScoreSegment struct {
	Range  string `json:"range"`
	Count  int    `json:"count"`
	Label  string `json:"label"`
}

type ClassScoreStats struct {
	Class         string        `json:"class"`
	Subject       string        `json:"subject"`
	Semester      string        `json:"semester"`
	Average       float64       `json:"average"`
	MaxScore      float64       `json:"max_score"`
	MinScore      float64       `json:"min_score"`
	TotalStudents int           `json:"total_students"`
	ValidScores   int           `json:"valid_scores"`
	Segments      []ScoreSegment `json:"segments"`
}

type StudentScoreHistory struct {
	StudentID    string      `json:"student_id"`
	Subject      string      `json:"subject"`
	Semester     string      `json:"semester"`
	LatestScore  *ExamScore  `json:"latest_score"`
}

type Message struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"sender_id"`
	ReceiverID string   `json:"receiver_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
}

type Conversation struct {
	ID              string    `json:"id"`
	Participant1ID  string    `json:"participant1_id"`
	Participant2ID  string    `json:"participant2_id"`
	LastMessageTime time.Time `json:"last_message_time"`
	LastMessage     string    `json:"last_message"`
}

type ConversationListItem struct {
	ConversationID string `json:"conversation_id"`
	OtherUserID    string `json:"other_user_id"`
	OtherUserName  string `json:"other_user_name"`
	LastMessage    string `json:"last_message"`
	LastMessageTime time.Time `json:"last_message_time"`
	UnreadCount    int    `json:"unread_count"`
}

type BatchScoreInput struct {
	Class     string         `json:"class"`
	Subject   string         `json:"subject"`
	Semester  string         `json:"semester"`
	ExamDate  time.Time      `json:"exam_date"`
	Scores    []StudentScore `json:"scores"`
}

type StudentScore struct {
	StudentID string   `json:"student_id"`
	Score     *float64 `json:"score"`
}
