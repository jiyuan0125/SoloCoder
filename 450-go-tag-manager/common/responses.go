package common

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateTagResponse struct {
	TagID string `json:"tag_id"`
}

type CreateTagGroupResponse struct {
	GroupID string `json:"group_id"`
}

type GetTagImpactResponse struct {
	ImpactedUserCount int64 `json:"impacted_user_count"`
}

type ListTagsResponse struct {
	Tags []Tag `json:"tags"`
}

type ListTagGroupsResponse struct {
	Groups []TagGroup `json:"groups"`
}

type GetTagStatsResponse struct {
	Stats []TagUsageStats `json:"stats"`
}

type GetTagTrendResponse struct {
	Trends []TagTrend `json:"trends"`
}

type ListAuditLogsResponse struct {
	Logs []AuditLog `json:"logs"`
}

type FilterUsersResponse struct {
	UserIDs []string `json:"user_ids"`
}

type GetUserTagsResponse struct {
	Tags []UserTagWithInfo `json:"tags"`
}

type UserTagWithInfo struct {
	TagID      string      `json:"tag_id"`
	TagName    string      `json:"tag_name"`
	TagType    TagType     `json:"tag_type"`
	TagValue   interface{} `json:"tag_value"`
	CreatedAt  int64       `json:"created_at"`
}
