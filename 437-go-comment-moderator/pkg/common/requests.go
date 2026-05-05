package common

type SubmitCommentRequest struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

type EditCommentRequest struct {
	CommentID string `json:"comment_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
}

type ReportCommentRequest struct {
	CommentID string `json:"comment_id"`
	UserID    string `json:"user_id"`
	Reason    string `json:"reason"`
}

type BatchApproveRequest struct {
	ModeratorID string   `json:"moderator_id"`
	CommentIDs  []string `json:"comment_ids"`
}

type BatchRejectRequest struct {
	ModeratorID string   `json:"moderator_id"`
	CommentIDs  []string `json:"comment_ids"`
	Reason      string   `json:"reason"`
}

type AddSensitiveWordRequest struct {
	AdminID string              `json:"admin_id"`
	Word    string              `json:"word"`
	Level   SensitiveWordLevel  `json:"level"`
}

type RemoveSensitiveWordRequest struct {
	AdminID string `json:"admin_id"`
	Word    string `json:"word"`
}

type GetPendingCommentsRequest struct {
	ModeratorID string `json:"moderator_id"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type ViewRejectedCommentRequest struct {
	UserID    string `json:"user_id"`
	CommentID string `json:"comment_id"`
}

type RegisterModeratorRequest struct {
	AdminID string `json:"admin_id"`
	Name    string `json:"name"`
}
