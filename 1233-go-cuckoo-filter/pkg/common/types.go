package common

type CreateRequest struct {
	Name     string `json:"name"`
	Capacity uint64 `json:"capacity"`
}

type CreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type InsertRequest struct {
	Name     string   `json:"name"`
	Element  string   `json:"element"`
	Elements []string `json:"elements"`
}

type InsertResponse struct {
	Success bool   `json:"success"`
	Inserted int    `json:"inserted"`
	Failed   int    `json:"failed"`
	Message  string `json:"message,omitempty"`
}

type LookupRequest struct {
	Name     string `json:"name"`
	Element  string `json:"element"`
}

type LookupResponse struct {
	Name     string `json:"name"`
	Element  string `json:"element"`
	Exists   bool   `json:"exists"`
}

type DeleteRequest struct {
	Name    string `json:"name"`
	Element string `json:"element"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type InfoRequest struct {
	Name string `json:"name"`
}

type InfoResponse struct {
	Name       string  `json:"name"`
	Exists     bool    `json:"exists"`
	UsedSlots  uint64  `json:"used_slots"`
	TotalSlots uint64  `json:"total_slots"`
	LoadRate   float64 `json:"load_rate"`
}
