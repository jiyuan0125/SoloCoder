package common

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type AnnouncementDetailResponse struct {
	Announcement Announcement           `json:"announcement"`
	ViewStats    ViewStatistics         `json:"view_stats"`
}

type ViewStatistics struct {
	ViewCount     int     `json:"view_count"`
	TotalAudience int     `json:"total_audience"`
	ReadRate      float64 `json:"read_rate"`
	ReadUserIDs   []string `json:"read_user_ids"`
}

type AnnouncementListResponse struct {
	Announcements []Announcement `json:"announcements"`
	Total         int            `json:"total"`
}

type ApprovalListResponse struct {
	Approvals []ApprovalRecord `json:"approvals"`
	Total     int               `json:"total"`
}

type UserInfoResponse struct {
	User User `json:"user"`
}

type DepartmentListResponse struct {
	Departments []Department `json:"departments"`
}
