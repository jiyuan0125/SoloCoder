package common

type TagType string

const (
	TagTypeEnum   TagType = "enum"
	TagTypeNumber TagType = "number"
	TagTypeText   TagType = "text"
)

type Tag struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        TagType   `json:"type"`
	Description string    `json:"description"`
	GroupID     string    `json:"group_id"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
	EnumConfig  *EnumConfig  `json:"enum_config,omitempty"`
	NumberConfig *NumberConfig `json:"number_config,omitempty"`
	TextConfig  *TextConfig   `json:"text_config,omitempty"`
}

type EnumConfig struct {
	Values []string `json:"values"`
}

type NumberConfig struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type TextConfig struct {
	MaxLength int `json:"max_length"`
}

type TagGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

type UserTag struct {
	UserID     string      `json:"user_id"`
	TagID      string      `json:"tag_id"`
	TagValue   interface{} `json:"tag_value"`
	CreatedAt  int64       `json:"created_at"`
}

type AuditLog struct {
	ID        string `json:"id"`
	Operator  string `json:"operator"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Details   string `json:"details"`
	Timestamp int64  `json:"timestamp"`
}

type TagUsageStats struct {
	TagID    string `json:"tag_id"`
	TagName  string `json:"tag_name"`
	UserCount int64 `json:"user_count"`
}

type TagTrend struct {
	TagID        string `json:"tag_id"`
	TagName      string `json:"tag_name"`
	NewUserCount int64  `json:"new_user_count"`
}
