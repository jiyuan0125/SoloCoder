package protocol

import (
	"encoding/json"
	"time"
)

type RequestType string

const (
	RequestTypeAddDomain     RequestType = "add_domain"
	RequestTypeRemoveDomain  RequestType = "remove_domain"
	RequestTypeListDomains   RequestType = "list_domains"
	RequestTypeCheckDomain   RequestType = "check_domain"
	RequestTypeCheckAll      RequestType = "check_all"
	RequestTypeGetHistory    RequestType = "get_history"
	RequestTypeGetStatus     RequestType = "get_status"
)

type ResponseType string

const (
	ResponseTypeSuccess ResponseType = "success"
	ResponseTypeError   ResponseType = "error"
)

type Request struct {
	Type    RequestType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Response struct {
	Type    ResponseType    `json:"type"`
	Message string          `json:"message,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type AddDomainPayload struct {
	Domain string `json:"domain"`
}

type RemoveDomainPayload struct {
	Domain string `json:"domain"`
}

type CheckDomainPayload struct {
	Domain   string `json:"domain"`
	WarnDays int    `json:"warn_days,omitempty"`
}

type CheckAllPayload struct {
	WarnDays int `json:"warn_days,omitempty"`
}

type GetHistoryPayload struct {
	Domain string `json:"domain"`
	Limit  int    `json:"limit,omitempty"`
}

type CertInfo struct {
	Domain           string    `json:"domain"`
	CommonName       string    `json:"common_name"`
	Issuer           string    `json:"issuer"`
	ValidFrom        time.Time `json:"valid_from"`
	ValidTo          time.Time `json:"valid_to"`
	RemainingDays    int       `json:"remaining_days"`
	SANs             []string  `json:"sans"`
	IsWildcard       bool      `json:"is_wildcard"`
	IsExpired        bool      `json:"is_expired"`
	Error            string    `json:"error,omitempty"`
	CheckedAt        time.Time `json:"checked_at"`
	RequestDomainOK  bool      `json:"request_domain_ok"`
}

type DomainStatus struct {
	Domain       string     `json:"domain"`
	LastCheck    *CertInfo  `json:"last_check,omitempty"`
	AddedAt      time.Time  `json:"added_at"`
	CheckHistory []CertInfo `json:"check_history,omitempty"`
}

type ListDomainsResponse struct {
	Domains []string `json:"domains"`
}

type CheckAllResponse struct {
	Results []CertInfo `json:"results"`
}

type GetStatusResponse struct {
	Status DomainStatus `json:"status"`
}

type GetHistoryResponse struct {
	History []CertInfo `json:"history"`
}
