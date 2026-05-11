package core

import "errors"

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrProjectAlreadyPublished  = errors.New("project already published")
	ErrProjectNotDraft          = errors.New("project is not in draft status")
	ErrOpenTimeTooEarly         = errors.New("open time must be at least 2 hours after bid deadline")
	ErrInviteRequiresSuppliers  = errors.New("invite tender requires invited suppliers list")
	ErrBidDeadlinePassed        = errors.New("bid deadline has passed")
	ErrBidNotFound              = errors.New("bid not found")
	ErrProjectNotPublished      = errors.New("project is not published")
	ErrInvalidAmount            = errors.New("amount must be greater than zero")
	ErrSupplierNotInvited       = errors.New("supplier is not invited")
	ErrAlreadySubmitted         = errors.New("bid already submitted for this project by this supplier")
	ErrProjectAlreadyOpened     = errors.New("project already opened")
	ErrOpenTimeNotReached       = errors.New("open time not reached yet")
	ErrNoBids                   = errors.New("no valid bids for this project")
)
