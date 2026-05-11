package common

import "errors"

var (
	ErrActivityNotFound          = errors.New("activity not found")
	ErrParticipantNotFound       = errors.New("participant not found")
	ErrRegistrationNotFound      = errors.New("registration not found")
	ErrFeedbackNotFound          = errors.New("feedback not found")
	ErrPhoneAlreadyRegistered    = errors.New("phone already registered")
	ErrRoomNumberAlreadyInUse    = errors.New("room number already in use")
	ErrInvalidActivityType       = errors.New("invalid activity type")
	ErrInvalidFrequency          = errors.New("invalid frequency")
	ErrRegistrationClosed        = errors.New("registration is closed")
	ErrActivityFull              = errors.New("activity is full")
	ErrAlreadyRegistered         = errors.New("participant already registered for this activity")
	ErrCancellationPeriodPassed  = errors.New("cancellation period has passed")
	ErrActivityAlreadyStarted    = errors.New("activity has already started")
	ErrActivityAlreadyCompleted  = errors.New("activity has already completed")
	ErrActivityAlreadyCancelled  = errors.New("activity is already cancelled")
	ErrCheckInWindowClosed       = errors.New("check-in window is closed")
	ErrCheckInNotYetOpen         = errors.New("check-in is not yet open")
	ErrPhoneMismatch             = errors.New("phone last 4 digits mismatch")
	ErrFeedbackPeriodPassed      = errors.New("feedback collection period has passed")
	ErrActivityNotEnded          = errors.New("activity has not ended yet")
	ErrAlreadySubmittedFeedback  = errors.New("feedback already submitted")
	ErrInvalidRating             = errors.New("rating must be between 1 and 5")
	ErrPrerequisitesNotMet       = errors.New("activity prerequisites not met")
	ErrConcurrentRegistration    = errors.New("concurrent registration conflict, please try again")
	ErrInvalidPhoneNumber        = errors.New("invalid phone number")
	ErrInvalidRoomNumber         = errors.New("invalid room number")
	ErrMissingRequiredField      = errors.New("missing required field")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
