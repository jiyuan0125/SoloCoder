package server

import (
	"encoding/json"
	"regexp"
	"time"

	"event-collector/common"
)

type Validator struct {
	eventNameRegex *regexp.Regexp
}

func NewValidator() *Validator {
	return &Validator{
		eventNameRegex: regexp.MustCompile(common.EventNamePattern),
	}
}

func (v *Validator) ValidateEvent(event *common.Event) error {
	if event.Name == "" {
		return common.NewAPIError(common.ErrCodeInvalidEventName, "Event name is required")
	}

	if len(event.Name) > common.MaxEventNameLength {
		return common.NewAPIError(common.ErrCodeInvalidEventName, "Event name exceeds maximum length")
	}

	if !v.eventNameRegex.MatchString(event.Name) {
		return common.NewAPIError(common.ErrCodeInvalidEventName, "Event name must contain only alphanumeric characters and underscores")
	}

	now := time.Now().Unix()
	if event.Timestamp < now-common.TimestampPastLimit {
		return common.NewAPIError(common.ErrCodeInvalidTimestamp, "Timestamp is too far in the past")
	}

	if event.Timestamp > now+common.TimestampFutureLimit {
		return common.NewAPIError(common.ErrCodeInvalidTimestamp, "Timestamp is too far in the future")
	}

	propsJSON, err := json.Marshal(event.Properties)
	if err != nil {
		return common.NewAPIError(common.ErrCodeInvalidRequest, "Invalid properties JSON")
	}

	if len(propsJSON) > common.MaxPropertiesSize {
		return common.NewAPIError(common.ErrCodePropertiesTooLarge, "Properties JSON exceeds 10KB limit")
	}

	if event.UserID == "" {
		return common.NewAPIError(common.ErrCodeMissingUserID, "User ID is required")
	}

	if event.DeviceID == "" {
		return common.NewAPIError(common.ErrCodeMissingDeviceID, "Device ID is required")
	}

	return nil
}
