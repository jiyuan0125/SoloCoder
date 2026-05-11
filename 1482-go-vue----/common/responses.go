package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type GetOrdersResponse struct {
	Orders []Order `json:"orders"`
}

type GetOrderResponse struct {
	Order Order `json:"order"`
}

type GetPartsResponse struct {
	Parts []Part `json:"parts"`
}

type GetPartResponse struct {
	Part Part `json:"part"`
}

type GetReplenishmentsResponse struct {
	Todos []ReplenishmentTodo `json:"todos"`
}
