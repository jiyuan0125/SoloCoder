package api

type CreateGroupRequest struct {
	GroupID string `json:"group_id"`
	Topic   string `json:"topic"`
}

type CreateGroupResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type JoinGroupRequest struct {
	GroupID     string `json:"group_id"`
	ConsumerID  string `json:"consumer_id"`
}

type JoinGroupResponse struct {
	AssignedOffset int64  `json:"assigned_offset"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

type CommitOffsetRequest struct {
	GroupID    string `json:"group_id"`
	ConsumerID string `json:"consumer_id"`
	Offset     int64  `json:"offset"`
}

type CommitOffsetResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetOffsetRequest struct {
	GroupID    string `json:"group_id"`
	ConsumerID string `json:"consumer_id"`
}

type GetOffsetResponse struct {
	Offset  int64  `json:"offset"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
