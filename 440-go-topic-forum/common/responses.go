package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type UserResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	IsAdmin     bool   `json:"is_admin"`
	PostCount   int    `json:"post_count"`
	ReplyCount  int    `json:"reply_count"`
	LikeCount   int    `json:"like_count"`
	CreatedAt   string `json:"created_at"`
}

type PostResponse struct {
	ID             string         `json:"id"`
	UserID         string         `json:"user_id"`
	Username       string         `json:"username"`
	Title          string         `json:"title"`
	Content        string         `json:"content"`
	Category       string         `json:"category"`
	Tags           []string       `json:"tags"`
	ViewCount      int            `json:"view_count"`
	ReplyCount     int            `json:"reply_count"`
	LikeCount      int            `json:"like_count"`
	IsTop          bool           `json:"is_top"`
	IsEssence      bool           `json:"is_essence"`
	BestReplyID    string         `json:"best_reply_id,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type PostHistoryResponse struct {
	ID       string   `json:"id"`
	PostID   string   `json:"post_id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	EditedAt string   `json:"edited_at"`
	EditorID string   `json:"editor_id"`
}

type ReplyResponse struct {
	ID        string           `json:"id"`
	PostID    string           `json:"post_id"`
	UserID    string           `json:"user_id"`
	Username  string           `json:"username"`
	Content   string           `json:"content"`
	ParentID  string           `json:"parent_id"`
	Floor     int              `json:"floor"`
	LikeCount int              `json:"like_count"`
	CreatedAt string           `json:"created_at"`
	Children  []ReplyResponse  `json:"children,omitempty"`
	IsBest    bool             `json:"is_best,omitempty"`
}

type ReportResponse struct {
	ID          string `json:"id"`
	PostID      string `json:"post_id"`
	ReporterID  string `json:"reporter_id"`
	Reason      string `json:"reason"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	ReviewedAt  string `json:"reviewed_at,omitempty"`
	ReviewerID  string `json:"reviewer_id,omitempty"`
}

type TagCloudResponse struct {
	Tags []TagUsage `json:"tags"`
}

type HotPostsResponse struct {
	Date  string              `json:"date"`
	Posts []HotPostWithDetail `json:"posts"`
}

type HotPostWithDetail struct {
	PostID   string  `json:"post_id"`
	Title    string  `json:"title"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
}

type LeaderboardResponse struct {
	Users []UserRank `json:"users"`
}

type UserRank struct {
	Rank      int    `json:"rank"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	LikeCount int    `json:"like_count"`
}

type SearchResponse struct {
	Total int             `json:"total"`
	Posts []*PostResponse `json:"posts"`
}

type PostDetailResponse struct {
	Post     PostResponse          `json:"post"`
	Replies  []ReplyResponse       `json:"replies"`
	History  []*PostHistoryResponse `json:"history"`
}
