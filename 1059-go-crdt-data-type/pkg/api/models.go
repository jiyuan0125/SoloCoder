package api

import "github.com/example/crdt/pkg/crdt"

type CRDTType string

const (
	TypeGCounter  CRDTType = "gcounter"
	TypePNCounter CRDTType = "pncounter"
	TypeGSet      CRDTType = "gset"
)

type CreateInstanceRequest struct {
	Type CRDTType `json:"type"`
}

type CreateInstanceResponse struct {
	ID         string      `json:"id"`
	Type       CRDTType    `json:"type"`
	State      interface{} `json:"state"`
	CreatedAt  string      `json:"created_at"`
}

type OperationRequest struct {
	Operation string      `json:"operation"`
	NodeID    string      `json:"node_id,omitempty"`
	Element   string      `json:"element,omitempty"`
	Delta     int         `json:"delta,omitempty"`
}

type OperationResponse struct {
	ID       string      `json:"id"`
	Type     CRDTType    `json:"type"`
	Value    int         `json:"value,omitempty"`
	Elements []string    `json:"elements,omitempty"`
	State    interface{} `json:"state"`
}

type MergeRequest struct {
	TargetID string `json:"target_id"`
	SourceID string `json:"source_id"`
}

type MergeResponse struct {
	TargetID  string      `json:"target_id"`
	SourceID  string      `json:"source_id"`
	Value     int         `json:"value,omitempty"`
	Elements  []string    `json:"elements,omitempty"`
	State     interface{} `json:"state"`
}

type GetInstanceResponse struct {
	ID        string      `json:"id"`
	Type      CRDTType    `json:"type"`
	Value     int         `json:"value,omitempty"`
	Elements  []string    `json:"elements,omitempty"`
	State     interface{} `json:"state"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func GCounterToState(gc *crdt.GCounter) map[string]int {
	result := make(map[string]int)
	for k, v := range gc.Counters {
		result[k] = v
	}
	return result
}

func PNCounterToState(pn *crdt.PNCounter) map[string]interface{} {
	return map[string]interface{}{
		"positive": GCountersToMap(pn.Positive),
		"negative": GCountersToMap(pn.Negative),
	}
}

func GCountersToMap(gc *crdt.GCounter) map[string]int {
	result := make(map[string]int)
	for k, v := range gc.Counters {
		result[k] = v
	}
	return result
}

func GSetToState(gset *crdt.GSet) []string {
	return gset.ElementsList()
}
