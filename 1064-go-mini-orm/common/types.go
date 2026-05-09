package common

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type InsertRequest struct {
	Model string                 `json:"model"`
	Data  map[string]interface{} `json:"data"`
}

type InsertResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetByIDRequest struct {
	Model string      `json:"model"`
	ID    interface{} `json:"id"`
}

type GetByIDResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type UpdateRequest struct {
	Model string                 `json:"model"`
	ID    interface{}            `json:"id"`
	Data  map[string]interface{} `json:"data"`
}

type UpdateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Model string      `json:"model"`
	ID    interface{} `json:"id"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListRequest struct {
	Model   string                 `json:"model"`
	Where   map[string]interface{} `json:"where,omitempty"`
	OrderBy string                 `json:"order_by,omitempty"`
	Limit   int                    `json:"limit,omitempty"`
	Offset  int                    `json:"offset,omitempty"`
}

type ListResponse struct {
	Success bool                     `json:"success"`
	Data    []map[string]interface{} `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}
