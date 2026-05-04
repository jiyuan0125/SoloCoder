package common

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Rating int

func (r *Rating) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("评分不能为空")
	}

	str := string(data)
	if str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	parsed, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return errors.New("评分必须是1-5之间的整数")
	}

	*r = Rating(parsed)
	return nil
}

func (r Rating) Int() int {
	return int(r)
}

func (r Rating) Validate() error {
	if r < MinRating || r > MaxRating {
		return errors.New("评分必须在1-5之间")
	}
	return nil
}

type FeedbackType string

const (
	FeedbackTypeSuggestion FeedbackType = "suggestion"
	FeedbackTypeBug        FeedbackType = "bug"
	FeedbackTypeComplaint  FeedbackType = "complaint"
)

func IsValidFeedbackType(t string) bool {
	ft := FeedbackType(strings.ToLower(t))
	return ft == FeedbackTypeSuggestion || ft == FeedbackTypeBug || ft == FeedbackTypeComplaint
}

func (ft FeedbackType) String() string {
	return string(ft)
}

func (ft FeedbackType) DisplayName() string {
	switch ft {
	case FeedbackTypeSuggestion:
		return "功能建议"
	case FeedbackTypeBug:
		return "Bug报告"
	case FeedbackTypeComplaint:
		return "体验投诉"
	default:
		return "未知类型"
	}
}

type FeedbackStatus string

const (
	StatusPending   FeedbackStatus = "pending"
	StatusInProcess FeedbackStatus = "in_process"
	StatusClosed    FeedbackStatus = "closed"
)

func IsValidFeedbackStatus(s string) bool {
	fs := FeedbackStatus(strings.ToLower(s))
	return fs == StatusPending || fs == StatusInProcess || fs == StatusClosed
}

func (fs FeedbackStatus) String() string {
	return string(fs)
}

func (fs FeedbackStatus) DisplayName() string {
	switch fs {
	case StatusPending:
		return "待处理"
	case StatusInProcess:
		return "处理中"
	case StatusClosed:
		return "已关闭"
	default:
		return "未知状态"
	}
}

const (
	MinRating         = 1
	MaxRating         = 5
	MaxDescriptionLen = 500
	MaxNoteLen        = 500
	MaxDailyFeedbacks = 5
	DefaultPageSize   = 10
	MaxPageSize       = 100
	MinPageNum        = 1
)

func ValidateRating(rating interface{}) (int, error) {
	var r int

	switch v := rating.(type) {
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, errors.New("评分必须是1-5之间的整数")
		}
		r = parsed
	case int:
		r = v
	case int8:
		r = int(v)
	case int16:
		r = int(v)
	case int32:
		r = int(v)
	case int64:
		r = int(v)
	case float64:
		if v != float64(int(v)) {
			return 0, errors.New("评分必须是1-5之间的整数")
		}
		r = int(v)
	default:
		return 0, errors.New("评分格式无效")
	}

	if r < MinRating || r > MaxRating {
		return 0, errors.New("评分必须在1-5之间")
	}

	return r, nil
}

func ValidateDescription(desc string) error {
	if len(desc) == 0 {
		return errors.New("反馈描述不能为空")
	}
	if len(desc) > MaxDescriptionLen {
		return errors.New("反馈描述不能超过500个字符")
	}
	return nil
}

func ValidateNote(note string) error {
	if len(note) > MaxNoteLen {
		return errors.New("备注不能超过500个字符")
	}
	return nil
}

func ValidatePageParams(page, pageSize int) (int, int) {
	if page < MinPageNum {
		page = MinPageNum
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

type SubmitFeedbackRequest struct {
	UserID      string `json:"user_id"`
	Type        string `json:"type"`
	Rating      Rating `json:"rating"`
	Description string `json:"description"`
}

type SubmitFeedbackResponse struct {
	FeedbackID string `json:"feedback_id"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
}

type FeedbackListItem struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Type        FeedbackType   `json:"type"`
	Rating      int            `json:"rating"`
	Description string         `json:"description"`
	Status      FeedbackStatus `json:"status"`
	InternalNote string        `json:"internal_note,omitempty"`
	ProcessNote string         `json:"process_note,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ListFeedbackRequest struct {
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
}

type ListFeedbackResponse struct {
	Success     bool               `json:"success"`
	Feedbacks   []FeedbackListItem `json:"feedbacks"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalPages  int                `json:"total_pages"`
	Message     string             `json:"message,omitempty"`
}

type UpdateStatusRequest struct {
	FeedbackID  string `json:"feedback_id"`
	Status      string `json:"status"`
	ProcessNote string `json:"process_note,omitempty"`
}

type UpdateStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type AddNoteRequest struct {
	FeedbackID string `json:"feedback_id"`
	Note       string `json:"note"`
}

type AddNoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatisticsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ByType []TypeStats `json:"by_type"`
		Overall struct {
			TotalCount   int     `json:"total_count"`
			AverageRating float64 `json:"average_rating"`
		} `json:"overall"`
	} `json:"data"`
	Message string `json:"message,omitempty"`
}

type TypeStats struct {
	Type          FeedbackType `json:"type"`
	TypeName      string       `json:"type_name"`
	Count         int          `json:"count"`
	AverageRating float64      `json:"average_rating"`
}

type GetFeedbackRequest struct {
	FeedbackID string `json:"feedback_id"`
}

type GetFeedbackResponse struct {
	Success  bool               `json:"success"`
	Feedback *FeedbackListItem  `json:"feedback,omitempty"`
	Message  string             `json:"message,omitempty"`
}
