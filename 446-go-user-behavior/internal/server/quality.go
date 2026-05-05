package server

import (
	"regexp"
	"time"
	"userbehavior/internal/shared"
)

const (
	MaxURLReasonableLength = 2048
	MaxDurationHours       = 24
	MinButtonIDLength      = 1
	MaxButtonIDLength      = 100
)

var (
	urlRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+\-.]*://`)
)

type QualityMonitor struct {
}

func NewQualityMonitor() *QualityMonitor {
	return &QualityMonitor{}
}

func (m *QualityMonitor) ValidateAndFilter(behavior *shared.Behavior) (isValid bool, reason string) {
	if behavior.UserID == "" {
		return false, "user_id is empty"
	}

	if behavior.SessionID == "" {
		return false, "session_id is empty"
	}

	if behavior.Timestamp.IsZero() {
		return false, "timestamp is zero"
	}

	if behavior.Timestamp.After(time.Now().Add(1 * time.Hour)) {
		return false, "timestamp is in the future"
	}

	switch behavior.Type {
	case shared.BehaviorTypePageView:
		return m.validatePageView(behavior)
	case shared.BehaviorTypeButtonClick:
		return m.validateButtonClick(behavior)
	case shared.BehaviorTypeFeatureUse:
		return m.validateFeatureUse(behavior)
	default:
		return false, "unknown behavior type"
	}
}

func (m *QualityMonitor) validatePageView(behavior *shared.Behavior) (isValid bool, reason string) {
	if behavior.PageView == nil {
		return false, "page_view data is missing for page_view type"
	}

	if behavior.PageView.URL == "" {
		return false, "url is empty for page_view"
	}

	if len(behavior.PageView.URL) > MaxURLReasonableLength {
		return false, "url is too long"
	}

	if behavior.PageView.Duration < 0 {
		return false, "duration is negative"
	}

	if behavior.PageView.Duration > time.Duration(MaxDurationHours)*time.Hour {
		return false, "duration exceeds reasonable limit"
	}

	if behavior.PageView.PageKey == "" {
		return false, "page_key is empty for page_view"
	}

	return true, ""
}

func (m *QualityMonitor) validateButtonClick(behavior *shared.Behavior) (isValid bool, reason string) {
	if behavior.ButtonClick == nil {
		return false, "button_click data is missing for button_click type"
	}

	if behavior.ButtonClick.ButtonID == "" {
		return false, "button_id is empty for button_click"
	}

	if len(behavior.ButtonClick.ButtonID) < MinButtonIDLength {
		return false, "button_id is too short"
	}

	if len(behavior.ButtonClick.ButtonID) > MaxButtonIDLength {
		return false, "button_id is too long"
	}

	return true, ""
}

func (m *QualityMonitor) validateFeatureUse(behavior *shared.Behavior) (isValid bool, reason string) {
	if behavior.FeatureUse == nil {
		return false, "feature_use data is missing for feature_use type"
	}

	if behavior.FeatureUse.FeatureName == "" {
		return false, "feature_name is empty for feature_use"
	}

	if behavior.FeatureUse.Duration < 0 {
		return false, "duration is negative"
	}

	if behavior.FeatureUse.Duration > time.Duration(MaxDurationHours)*time.Hour {
		return false, "duration exceeds reasonable limit"
	}

	return true, ""
}

func (m *QualityMonitor) IsDuplicate(behavior *shared.Behavior, existing []*shared.Behavior) bool {
	if behavior.Type != shared.BehaviorTypePageView {
		return false
	}

	for _, existingBehavior := range existing {
		if existingBehavior.Type != shared.BehaviorTypePageView {
			continue
		}

		if existingBehavior.UserID != behavior.UserID {
			continue
		}

		if existingBehavior.IsFiltered {
			continue
		}

		timeDiff := behavior.Timestamp.Sub(existingBehavior.Timestamp)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		if timeDiff <= shared.DeduplicationWindow {
			if behavior.PageView != nil && existingBehavior.PageView != nil {
				if behavior.PageView.PageKey == existingBehavior.PageView.PageKey {
					return true
				}
			}
		}
	}

	return false
}
