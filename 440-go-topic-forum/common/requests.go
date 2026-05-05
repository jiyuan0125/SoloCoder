package common

type CreateUserRequest struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
}

type CreatePostRequest struct {
	UserID   string   `json:"user_id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Tags     []string `json:"tags,omitempty"`
}

type UpdatePostRequest struct {
	UserID   string   `json:"user_id"`
	Title    string   `json:"title,omitempty"`
	Content  string   `json:"content,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type CreateReplyRequest struct {
	UserID   string `json:"user_id"`
	PostID   string `json:"post_id"`
	Content  string `json:"content"`
	ParentID string `json:"parent_id,omitempty"`
}

type LikeRequest struct {
	UserID     string `json:"user_id"`
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
}

type SetBestReplyRequest struct {
	UserID  string `json:"user_id"`
	PostID  string `json:"post_id"`
	ReplyID string `json:"reply_id"`
}

type SetTopRequest struct {
	AdminID string `json:"admin_id"`
	PostID  string `json:"post_id"`
	IsTop   bool   `json:"is_top"`
}

type SetEssenceRequest struct {
	AdminID   string `json:"admin_id"`
	PostID    string `json:"post_id"`
	IsEssence bool   `json:"is_essence"`
}

type ReportRequest struct {
	ReporterID string `json:"reporter_id"`
	PostID     string `json:"post_id"`
	Reason     string `json:"reason"`
}

type ReviewReportRequest struct {
	AdminID string `json:"admin_id"`
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
}

type SearchRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	Category string `json:"category,omitempty"`
	Tag      string `json:"tag,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

type ViewPostRequest struct {
	UserID string `json:"user_id"`
	PostID string `json:"post_id"`
}

type DeletePostRequest struct {
	UserID string `json:"user_id"`
	PostID string `json:"post_id"`
}
