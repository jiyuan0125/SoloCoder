package common

import "time"

type VehicleType int

const (
	Passenger VehicleType = iota + 1
	Truck
)

type PassengerClass int

const (
	PassengerClass1 PassengerClass = iota + 1
	PassengerClass2
	PassengerClass3
	PassengerClass4
)

type TruckClass int

const (
	TruckClass1 TruckClass = iota + 1
	TruckClass2
	TruckClass3
	TruckClass4
	TruckClass5
)

type AccountStatus string

const (
	StatusNormal    AccountStatus = "normal"
	StatusOwe       AccountStatus = "owe"
	StatusBlacklist AccountStatus = "blacklist"
)

type PaymentStatus string

const (
	PaymentSuccess PaymentStatus = "success"
	PaymentOwe     PaymentStatus = "owe"
	PaymentFree    PaymentStatus = "free"
)

type Account struct {
	ID          string        `json:"id"`
	LicensePlate string       `json:"license_plate"`
	BankCard    string        `json:"bank_card"`
	Balance     float64       `json:"balance"`
	Status      AccountStatus `json:"status"`
	OweAmount   float64       `json:"owe_amount"`
	OweDays     int           `json:"owe_days"`
	CreatedAt   time.Time     `json:"created_at"`
	VehicleType VehicleType   `json:"vehicle_type"`
	Seats       int           `json:"seats,omitempty"`
	LoadWeight  float64       `json:"load_weight,omitempty"`
}

type PassRecord struct {
	ID           string        `json:"id"`
	LicensePlate string        `json:"license_plate"`
	EntryStation string        `json:"entry_station"`
	ExitStation  string        `json:"exit_station"`
	PassDate     time.Time     `json:"pass_date"`
	Mileage      float64       `json:"mileage"`
	Fee          float64       `json:"fee"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	CreatedAt    time.Time     `json:"created_at"`
}

type CreateAccountRequest struct {
	LicensePlate string      `json:"license_plate"`
	BankCard     string      `json:"bank_card"`
	Balance      float64     `json:"balance"`
	VehicleType  VehicleType `json:"vehicle_type"`
	Seats        int         `json:"seats,omitempty"`
	LoadWeight   float64     `json:"load_weight,omitempty"`
}

type CreateAccountResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Account *Account `json:"account,omitempty"`
}

type GetAccountRequest struct {
	LicensePlate string `json:"license_plate"`
}

type GetAccountResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Account *Account `json:"account,omitempty"`
}

type RechargeRequest struct {
	LicensePlate string  `json:"license_plate"`
	Amount       float64 `json:"amount"`
}

type RechargeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Balance float64 `json:"balance,omitempty"`
}

type PassRequest struct {
	LicensePlate string    `json:"license_plate"`
	EntryStation string    `json:"entry_station"`
	ExitStation  string    `json:"exit_station"`
	PassDate     time.Time `json:"pass_date"`
	Mileage      float64   `json:"mileage"`
	IsFree       bool      `json:"is_free,omitempty"`
}

type PassResponse struct {
	Success       bool          `json:"success"`
	Message       string        `json:"message,omitempty"`
	Record        *PassRecord   `json:"record,omitempty"`
	PaymentStatus PaymentStatus `json:"payment_status,omitempty"`
}

type ExportRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type ExportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	CSVData string `json:"csv_data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
