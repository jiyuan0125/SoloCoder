package common

type SubmitCommentResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	CommentID string        `json:"comment_id,omitempty"`
	Status    CommentStatus `json:"status,omitempty"`
}

type EditCommentResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Status  CommentStatus `json:"status,omitempty"`
}

type ReportCommentResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ReportCount int    `json:"report_count,omitempty"`
}

type BatchActionResponse struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	Processed    int      `json:"processed,omitempty"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

type AddSensitiveWordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RemoveSensitiveWordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetPendingCommentsResponse struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	Comments  []*Comment `json:"comments,omitempty"`
	Total     int        `json:"total,omitempty"`
	HasMore   bool       `json:"has_more,omitempty"`
}

type ViewRejectedCommentResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RejectReason string `json:"reject_reason,omitempty"`
}

type RegisterModeratorResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ModeratorID string `json:"moderator_id,omitempty"`
}

type GetSensitiveWordsResponse struct {
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	Words    []*SensitiveWord `json:"words,omitempty"`
}

type GetAuditLogsResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Logs    []*AuditLog `json:"logs,omitempty"`
}

type GetModeratorStatsResponse struct {
	Success  bool                    `json:"success"`
	Message  string                  `json:"message"`
	Moderators map[string]ModeratorStats `json:"moderators,omitempty"`
}

type ModeratorStats struct {
	Name       string `json:"name"`
	DailyCount int    `json:"daily_count"`
	TotalCount int    `json:"total_count"`
}
