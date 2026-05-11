package common

import "time"

type ActivityType string

const (
	ActivityTypeArtPerformance ActivityType = "art_performance"
	ActivityTypeHealthLecture  ActivityType = "health_lecture"
	ActivityTypeParentChild    ActivityType = "parent_child"
	ActivityTypeSports         ActivityType = "sports"
	ActivityTypeVolunteer      ActivityType = "volunteer"
	ActivityTypeFestival       ActivityType = "festival"
	ActivityTypeOther          ActivityType = "other"
)

type ActivityFrequency string

const (
	ActivityFrequencySingle    ActivityFrequency = "single"
	ActivityFrequencyRecurring ActivityFrequency = "recurring"
)

type ActivityStatus string

const (
	ActivityStatusDraft    ActivityStatus = "draft"
	ActivityStatusActive   ActivityStatus = "active"
	ActivityStatusCancelled ActivityStatus = "cancelled"
	ActivityStatusCompleted ActivityStatus = "completed"
)

type RegistrationStatus string

const (
	RegistrationStatusRegistered RegistrationStatus = "registered"
	RegistrationStatusCancelled  RegistrationStatus = "cancelled"
	RegistrationStatusWaitlist   RegistrationStatus = "waitlist"
	RegistrationStatusCheckedIn  RegistrationStatus = "checked_in"
	RegistrationStatusNoShow     RegistrationStatus = "no_show"
)

type CheckInStatus string

const (
	CheckInStatusNotYet   CheckInStatus = "not_yet"
	CheckInStatusSuccess  CheckInStatus = "success"
	CheckInStatusLate     CheckInStatus = "late"
	CheckInStatusNotCheckedIn CheckInStatus = "not_checked_in"
)

type Participant struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	PhoneLast4  string `json:"phone_last4"`
	RoomNumber  string `json:"room_number"`
	Name        string `json:"name"`
	Age         int    `json:"age"`
	CreateTime  time.Time `json:"create_time"`
}

type Activity struct {
	ID                string             `json:"id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	StartTime         time.Time          `json:"start_time"`
	EndTime           time.Time          `json:"end_time"`
	Location          string             `json:"location"`
	MaxParticipants   int                `json:"max_participants"`
	RegistrationDeadline time.Time       `json:"registration_deadline"`
	ActivityType      ActivityType       `json:"activity_type"`
	Tags              []string           `json:"tags"`
	Frequency         ActivityFrequency  `json:"frequency"`
	RecurringRule     string             `json:"recurring_rule,omitempty"`
	SeriesID          string             `json:"series_id,omitempty"`
	EpisodeNumber     int                `json:"episode_number,omitempty"`
	Status            ActivityStatus     `json:"status"`
	CreateTime        time.Time          `json:"create_time"`
	UpdateTime        time.Time          `json:"update_time"`
}

type Registration struct {
	ID           string             `json:"id"`
	ActivityID   string             `json:"activity_id"`
	ParticipantID string            `json:"participant_id"`
	Status       RegistrationStatus `json:"status"`
	CheckInStatus CheckInStatus     `json:"checkin_status"`
	CheckInTime  *time.Time         `json:"checkin_time,omitempty"`
	CreateTime   time.Time          `json:"create_time"`
	UpdateTime   time.Time          `json:"update_time"`
}

type Feedback struct {
	ID            string    `json:"id"`
	ActivityID    string    `json:"activity_id"`
	ParticipantID string    `json:"participant_id"`
	Rating        int       `json:"rating"`
	Comments      string    `json:"comments"`
	CreateTime    time.Time `json:"create_time"`
}

type ActivitySummary struct {
	ActivityID       string    `json:"activity_id"`
	TotalRegistrations int     `json:"total_registrations"`
	CheckInCount     int       `json:"checkin_count"`
	CheckInRate      float64   `json:"checkin_rate"`
	AverageRating    float64   `json:"average_rating"`
	FeedbackCount    int       `json:"feedback_count"`
	GenerateTime     time.Time `json:"generate_time"`
}

type AnalysisTodo struct {
	ID           string    `json:"id"`
	ActivityID   string    `json:"activity_id"`
	CheckInRate  float64   `json:"checkin_rate"`
	CreateTime   time.Time `json:"create_time"`
	Resolved     bool      `json:"resolved"`
}

type ActivityDetailResponse struct {
	Activity
	RegisteredCount  int `json:"registered_count"`
	WaitlistCount    int `json:"waitlist_count"`
	AvailableSlots   int `json:"available_slots"`
}

type CreateActivityRequest struct {
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	StartTime            time.Time     `json:"start_time"`
	EndTime              time.Time     `json:"end_time"`
	Location             string        `json:"location"`
	MaxParticipants      int           `json:"max_participants"`
	RegistrationDeadline time.Time     `json:"registration_deadline"`
	ActivityType         ActivityType  `json:"activity_type"`
	Tags                 []string      `json:"tags"`
	Frequency            ActivityFrequency `json:"frequency"`
	RecurringRule        string        `json:"recurring_rule,omitempty"`
	EpisodeCount         int           `json:"episode_count,omitempty"`
}

type UpdateActivityRequest struct {
	Title                *string        `json:"title,omitempty"`
	Description          *string        `json:"description,omitempty"`
	StartTime            *time.Time     `json:"start_time,omitempty"`
	EndTime              *time.Time     `json:"end_time,omitempty"`
	Location             *string        `json:"location,omitempty"`
	MaxParticipants      *int           `json:"max_participants,omitempty"`
	RegistrationDeadline *time.Time     `json:"registration_deadline,omitempty"`
	ActivityType         *ActivityType  `json:"activity_type,omitempty"`
	Tags                 *[]string      `json:"tags,omitempty"`
}

type RegisterParticipantRequest struct {
	Phone      string `json:"phone"`
	RoomNumber string `json:"room_number"`
	Name       string `json:"name"`
	Age        int    `json:"age"`
}

type RegisterActivityRequest struct {
	ParticipantID string `json:"participant_id"`
	ActivityID    string `json:"activity_id"`
	AdultsCount   int    `json:"adults_count"`
	ChildrenCount int    `json:"children_count"`
}

type CancelRegistrationRequest struct {
	RegistrationID string `json:"registration_id"`
}

type CheckInRequest struct {
	ActivityID   string `json:"activity_id"`
	PhoneLast4   string `json:"phone_last4"`
}

type SubmitFeedbackRequest struct {
	ActivityID    string `json:"activity_id"`
	ParticipantID string `json:"participant_id"`
	Rating        int    `json:"rating"`
	Comments      string `json:"comments"`
}

type CancelActivityRequest struct {
	ActivityID string `json:"activity_id"`
}

type ListActivitiesRequest struct {
	Status   *ActivityStatus `json:"status,omitempty"`
	Type     *ActivityType   `json:"type,omitempty"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
