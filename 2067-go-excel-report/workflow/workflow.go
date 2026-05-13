package workflow

import "fmt"

const (
	StatusDraft      = "draft"
	StatusPending    = "pending"
	StatusApproved   = "approved"
	StatusExecuting  = "executing"
	StatusConfirmed  = "confirmed"
	StatusClosed     = "closed"
	StatusRejected   = "rejected"
)

const (
	ActionSubmit     = "submit"
	ActionApprove    = "approve"
	ActionReject     = "reject"
	ActionExecute    = "execute"
	ActionConfirm    = "confirm"
	ActionClose      = "close"
	ActionRollback   = "rollback"
)

var transitions = map[string][]string{
	StatusDraft:      {StatusPending},
	StatusPending:    {StatusApproved, StatusRejected, StatusDraft},
	StatusApproved:   {StatusExecuting, StatusPending},
	StatusRejected:   {StatusDraft},
	StatusExecuting:  {StatusConfirmed, StatusApproved},
	StatusConfirmed:  {StatusClosed, StatusExecuting},
	StatusClosed:     {},
}

func CanTransition(current, target string) bool {
	allowed, ok := transitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

func GetNextStatus(current, action string) (string, error) {
	switch action {
	case ActionSubmit:
		if current == StatusDraft {
			return StatusPending, nil
		}
	case ActionApprove:
		if current == StatusPending {
			return StatusApproved, nil
		}
	case ActionReject:
		if current == StatusPending {
			return StatusRejected, nil
		}
	case ActionExecute:
		if current == StatusApproved {
			return StatusExecuting, nil
		}
	case ActionConfirm:
		if current == StatusExecuting {
			return StatusConfirmed, nil
		}
	case ActionClose:
		if current == StatusConfirmed {
			return StatusClosed, nil
		}
	case ActionRollback:
		switch current {
		case StatusPending:
			return StatusDraft, nil
		case StatusRejected:
			return StatusDraft, nil
		case StatusApproved:
			return StatusPending, nil
		case StatusExecuting:
			return StatusApproved, nil
		case StatusConfirmed:
			return StatusExecuting, nil
		}
	}
	return "", fmt.Errorf("invalid action %q for status %q", action, current)
}

func IsClosed(status string) bool {
	return status == StatusClosed
}

func StatusDisplayName(status string) string {
	names := map[string]string{
		StatusDraft:     "草稿",
		StatusPending:   "待审批",
		StatusApproved:  "已审批",
		StatusExecuting: "执行中",
		StatusConfirmed: "已确认",
		StatusClosed:    "已关闭",
		StatusRejected:  "已拒绝",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}
