package common

type CreateTagRequest struct {
	Operator     string       `json:"operator"`
	Name         string       `json:"name"`
	Type         TagType      `json:"type"`
	Description  string       `json:"description"`
	GroupID      string       `json:"group_id"`
	EnumConfig   *EnumConfig  `json:"enum_config,omitempty"`
	NumberConfig *NumberConfig `json:"number_config,omitempty"`
	TextConfig   *TextConfig  `json:"text_config,omitempty"`
}

type UpdateTagRequest struct {
	Operator     string       `json:"operator"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	GroupID      string       `json:"group_id"`
	EnumConfig   *EnumConfig  `json:"enum_config,omitempty"`
	NumberConfig *NumberConfig `json:"number_config,omitempty"`
	TextConfig   *TextConfig  `json:"text_config,omitempty"`
}

type DeleteTagRequest struct {
	Operator string `json:"operator"`
}

type CreateTagGroupRequest struct {
	Operator    string `json:"operator"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ApplyTagToUserRequest struct {
	Operator  string      `json:"operator"`
	UserID    string      `json:"user_id"`
	TagValue  interface{} `json:"tag_value"`
}

type BatchApplyTagRequest struct {
	Operator  string        `json:"operator"`
	UserIDs   []string      `json:"user_ids"`
	TagValue  interface{}   `json:"tag_value"`
}

type FilterCondition struct {
	TagID    string      `json:"tag_id"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type FilterGroup struct {
	Logic     string            `json:"logic"`
	Conditions []FilterCondition `json:"conditions,omitempty"`
	Groups     []FilterGroup     `json:"groups,omitempty"`
}

type FilterUsersRequest struct {
	Filter FilterGroup `json:"filter"`
}

type GetTagImpactRequest struct {
}
