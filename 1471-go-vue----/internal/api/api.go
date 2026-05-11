package api

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Schedule struct {
	ScheduleNo       string `json:"schedule_no"`
	Date             string `json:"date"`
	DepartureStation string `json:"departure_station"`
	ArrivalStation   string `json:"arrival_station"`
	DepartureTime    string `json:"departure_time"`
	ArrivalTime      string `json:"arrival_time"`
	BusType          string `json:"bus_type"`
	Price            int    `json:"price"`
	TotalSeats       int    `json:"total_seats"`
	SoldSeats        int    `json:"sold_seats"`
	Status           string `json:"status"`
	IsAlmostSoldOut  bool   `json:"is_almost_sold_out"`
}

type CreateScheduleRequest struct {
	ScheduleNo       string `json:"schedule_no"`
	Date             string `json:"date"`
	DepartureStation string `json:"departure_station"`
	ArrivalStation   string `json:"arrival_station"`
	DepartureTime    string `json:"departure_time"`
	ArrivalTime      string `json:"arrival_time"`
	BusType          string `json:"bus_type"`
	Price            int    `json:"price"`
	TotalSeats       int    `json:"total_seats"`
}

type UpdateScheduleStatusRequest struct {
	ScheduleNo string `json:"schedule_no"`
	Date       string `json:"date"`
	Status     string `json:"status"`
}

type CancelScheduleRequest struct {
	ScheduleNo string `json:"schedule_no"`
	Date       string `json:"date"`
}

type SearchSchedulesRequest struct {
	Departure string `json:"departure"`
	Arrival   string `json:"arrival"`
	Date      string `json:"date"`
}

type Seat struct {
	SeatNo int  `json:"seat_no"`
	IsSold bool `json:"is_sold"`
}

type PurchaseTicketsRequest struct {
	ScheduleNo    string `json:"schedule_no"`
	Date          string `json:"date"`
	SeatNos       []int  `json:"seat_nos"`
	PassengerName string `json:"passenger_name"`
}

type Ticket struct {
	TicketNo      string `json:"ticket_no"`
	ScheduleNo    string `json:"schedule_no"`
	Date          string `json:"date"`
	SeatNo        int    `json:"seat_no"`
	PassengerName string `json:"passenger_name"`
	Price         int    `json:"price"`
	Status        string `json:"status"`
}

type PurchaseTicketsResponse struct {
	Tickets     []*Ticket `json:"tickets"`
	TotalAmount int       `json:"total_amount"`
}

type CheckInRequest struct {
	TicketNo string `json:"ticket_no"`
}

type RefundRequest struct {
	TicketNo string `json:"ticket_no"`
}

type RefundResult struct {
	RequestID    string `json:"request_id"`
	TicketNo     string `json:"ticket_no"`
	RefundAmount int    `json:"refund_amount"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type ProcessRefundRequest struct {
	RequestID string `json:"request_id"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason"`
}

func NewSuccessResponse(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func NewErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}
