package api

type VehicleStatus string

const (
	StatusInStock   VehicleStatus = "in_stock"
	StatusReserved  VehicleStatus = "reserved"
	StatusSold      VehicleStatus = "sold"
)

type TransmissionType string

const (
	TransmissionManual    TransmissionType = "manual"
	TransmissionAutomatic TransmissionType = "automatic"
)

type Vehicle struct {
	ID           string           `json:"id"`
	Brand        string           `json:"brand"`
	Model        string           `json:"model"`
	Year         int              `json:"year"`
	Color        string           `json:"color"`
	Mileage      float64          `json:"mileage"`
	Displacement float64          `json:"displacement"`
	Transmission TransmissionType `json:"transmission"`
	Status       VehicleStatus    `json:"status"`
	CreatedAt    int64            `json:"created_at"`
}

type Condition struct {
	Exterior  int `json:"exterior"`
	Interior  int `json:"interior"`
	Engine    int `json:"engine"`
	Chassis   int `json:"chassis"`
}

type Evaluation struct {
	ReportID    string    `json:"report_id"`
	VehicleID   string    `json:"vehicle_id"`
	Evaluator   string    `json:"evaluator"`
	Condition   Condition `json:"condition"`
	ReferencePriceFen int64 `json:"reference_price_fen"`
	FinalPriceFen     int64 `json:"final_price_fen"`
	Description string    `json:"description"`
	CreatedAt   int64     `json:"created_at"`
	IsDuplicate *bool     `json:"is_duplicate,omitempty"`
}

type Deposit struct {
	ID              string `json:"id"`
	VehicleID       string `json:"vehicle_id"`
	AmountFen       int64  `json:"amount_fen"`
	CustomerName    string `json:"customer_name"`
	CreatedAt       int64  `json:"created_at"`
	TransactionType string `json:"transaction_type"`
}

type TransferRecord struct {
	ID           string `json:"id"`
	VehicleID    string `json:"vehicle_id"`
	FullPriceFen int64  `json:"full_price_fen"`
	CustomerName string `json:"customer_name"`
	CreatedAt    int64  `json:"created_at"`
}

type CreateVehicleRequest struct {
	Brand        string           `json:"brand"`
	Model        string           `json:"model"`
	Year         int              `json:"year"`
	Color        string           `json:"color"`
	Mileage      float64          `json:"mileage"`
	Displacement float64          `json:"displacement"`
	Transmission TransmissionType `json:"transmission"`
}

type CreateVehicleResponse struct {
	Success bool    `json:"success"`
	Vehicle Vehicle `json:"vehicle"`
	Error   string  `json:"error,omitempty"`
}

type ListVehiclesRequest struct {
	Status VehicleStatus `json:"status,omitempty"`
}

type ListVehiclesResponse struct {
	Success bool      `json:"success"`
	Vehicles []Vehicle `json:"vehicles"`
	Error   string    `json:"error,omitempty"`
}

type GetVehicleRequest struct {
	VehicleID string `json:"vehicle_id"`
}

type GetVehicleResponse struct {
	Success bool    `json:"success"`
	Vehicle Vehicle `json:"vehicle"`
	Error   string  `json:"error,omitempty"`
}

type EvaluateVehicleRequest struct {
	VehicleID    string    `json:"vehicle_id"`
	Evaluator    string    `json:"evaluator"`
	Condition    Condition `json:"condition"`
	Description  string    `json:"description"`
}

type EvaluateVehicleResponse struct {
	Success    bool       `json:"success"`
	Evaluation Evaluation `json:"evaluation"`
	Error      string     `json:"error,omitempty"`
}

type ListEvaluationsRequest struct {
	VehicleID string `json:"vehicle_id"`
}

type ListEvaluationsResponse struct {
	Success     bool         `json:"success"`
	Evaluations []Evaluation `json:"evaluations"`
	Error       string       `json:"error,omitempty"`
}

type PayDepositRequest struct {
	VehicleID    string `json:"vehicle_id"`
	ReportID     string `json:"report_id"`
	CustomerName string `json:"customer_name"`
	AmountFen    *int64 `json:"amount_fen,omitempty"`
}

type PayDepositResponse struct {
	Success bool    `json:"success"`
	Deposit Deposit `json:"deposit"`
	Error   string  `json:"error,omitempty"`
}

type PayFullRequest struct {
	VehicleID    string `json:"vehicle_id"`
	CustomerName string `json:"customer_name"`
}

type PayFullResponse struct {
	Success        bool           `json:"success"`
	TransferRecord TransferRecord `json:"transfer_record"`
	Error          string         `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
