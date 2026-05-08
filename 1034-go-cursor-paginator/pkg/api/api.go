package api

import "encoding/json"

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type PageRequest struct {
	Cursor       string    `json:"cursor,omitempty"`
	Previous     string    `json:"previous,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	SortField    string    `json:"sort_field,omitempty"`
	SortOrder    SortOrder `json:"sort_order,omitempty"`
}

type PageResponse struct {
	Data         []json.RawMessage `json:"data"`
	NextCursor   string            `json:"next_cursor,omitempty"`
	PrevCursor   string            `json:"prev_cursor,omitempty"`
	HasNext      bool              `json:"has_next"`
	HasPrev      bool              `json:"has_prev"`
	Total        int               `json:"total,omitempty"`
	Limit        int               `json:"limit"`
}

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
}

func (u *User) GetID() int {
	return u.ID
}

func (u *User) GetFieldValue(field string) interface{} {
	switch field {
	case "id":
		return u.ID
	case "name":
		return u.Name
	case "email":
		return u.Email
	case "created_at":
		return u.CreatedAt
	default:
		return u.ID
	}
}
