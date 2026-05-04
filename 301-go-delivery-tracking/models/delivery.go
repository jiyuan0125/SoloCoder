package models

type DeliveryStatus struct {
	ID          int64   `json:"id"`
	OrderNo     string  `json:"order_no"`
	StatusName  string  `json:"status_name"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	OccurredAt  int64   `json:"occurred_at"`
	ReceivedAt  int64   `json:"received_at"`
}

type DeliveryStatusRequest struct {
	OrderNo    string  `json:"order_no"`
	StatusName string  `json:"status_name"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	OccurredAt int64   `json:"occurred_at"`
}

type DeliveryStatusBatchRequest struct {
	Statuses []DeliveryStatusRequest `json:"statuses"`
}

type DeliveryTrajectoryResponse struct {
	OrderNo  string           `json:"order_no"`
	Statuses []DeliveryStatus `json:"statuses"`
}
