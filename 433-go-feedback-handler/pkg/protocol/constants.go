package protocol

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

const (
	SevenDaysInHours   = 24 * 7
	TwentyFourHours    = 24
	SimilarityThreshold = 0.8
	MaxInvalidCloses   = 5
	ReviewDurationDays = 7
)

const (
	NotificationTypeReminder     = "reminder"
	NotificationTypeEscalation   = "escalation"
	NotificationTypeHighPriority = "high_priority"
	NotificationTypeReopened     = "reopened"
)

const (
	SystemTagReproduced      = "已复现"
	SystemTagConfirming      = "需求确认中"
	SystemTagNextVersionFix  = "下版本修复"
	SystemTagDuplicate       = "重复反馈"
	SystemTagWontFix         = "暂不处理"
)

var SystemTags = []struct {
	Name  string
	Color string
}{
	{SystemTagReproduced, "#3498db"},
	{SystemTagConfirming, "#f39c12"},
	{SystemTagNextVersionFix, "#27ae60"},
	{SystemTagDuplicate, "#95a5a6"},
	{SystemTagWontFix, "#e74c3c"},
}

var ValidStatusTransitions = map[FeedbackStatus][]FeedbackStatus{
	StatusPending:    {StatusProcessing, StatusClosed},
	StatusProcessing: {StatusResolved, StatusClosed},
	StatusResolved:   {StatusClosed, StatusProcessing},
	StatusClosed:     {StatusPending},
}

func IsValidStatusTransition(from, to FeedbackStatus) bool {
	if validTos, ok := ValidStatusTransitions[from]; ok {
		for _, validTo := range validTos {
			if validTo == to {
				return true
			}
		}
	}
	return false
}

func GetPriorityForFeedback(ftype FeedbackType) FeedbackPriority {
	switch ftype {
	case FeedbackTypeComplaint:
		return PriorityUrgent
	case FeedbackTypeBug:
		return PriorityHigh
	default:
		return PriorityMedium
	}
}

func IsHighPriority(priority FeedbackPriority) bool {
	return priority == PriorityHigh || priority == PriorityUrgent
}
