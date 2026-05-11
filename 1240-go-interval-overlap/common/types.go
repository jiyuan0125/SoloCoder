package common

type AddBookingRequest struct {
	Resource string `json:"resource"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Booker   string `json:"booker"`
}

type AddBookingResponse struct {
	Success   bool        `json:"success"`
	BookingID string      `json:"booking_id,omitempty"`
	Conflicts []BookingInfo `json:"conflicts,omitempty"`
	Message   string      `json:"message,omitempty"`
}

type ListBookingsRequest struct {
	Resource string `json:"resource"`
	Start    string `json:"start,omitempty"`
	End      string `json:"end,omitempty"`
}

type ListBookingsResponse struct {
	Bookings []BookingInfo `json:"bookings"`
}

type CheckConflictRequest struct {
	Resource string `json:"resource"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

type CheckConflictResponse struct {
	HasConflict bool          `json:"has_conflict"`
	Conflicts   []BookingInfo `json:"conflicts,omitempty"`
}

type DeleteBookingRequest struct {
	BookingID string `json:"booking_id"`
	Operator  string `json:"operator"`
}

type DeleteBookingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type BookingInfo struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Booker   string `json:"booker"`
}
