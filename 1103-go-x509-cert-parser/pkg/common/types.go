package common

import "time"

type ParseRequest struct {
	PEM string `json:"pem"`
}

type SubjectInfo struct {
	CommonName         string   `json:"commonName"`
	Organization       []string `json:"organization,omitempty"`
	OrganizationalUnit []string `json:"organizationalUnit,omitempty"`
	Country            []string `json:"country,omitempty"`
	Locality           []string `json:"locality,omitempty"`
	Province           []string `json:"province,omitempty"`
	StreetAddress      []string `json:"streetAddress,omitempty"`
	PostalCode         []string `json:"postalCode,omitempty"`
}

type IssuerInfo struct {
	CommonName         string   `json:"commonName"`
	Organization       []string `json:"organization,omitempty"`
	OrganizationalUnit []string `json:"organizationalUnit,omitempty"`
	Country            []string `json:"country,omitempty"`
	Locality           []string `json:"locality,omitempty"`
	Province           []string `json:"province,omitempty"`
	StreetAddress      []string `json:"streetAddress,omitempty"`
	PostalCode         []string `json:"postalCode,omitempty"`
}

type SANInfo struct {
	DNSNames []string `json:"dnsNames,omitempty"`
	IPAddresses []string `json:"ipAddresses,omitempty"`
}

type CertificateInfo struct {
	Subject      SubjectInfo `json:"subject"`
	Issuer       IssuerInfo  `json:"issuer"`
	SerialNumber string      `json:"serialNumber"`
	NotBefore    time.Time  `json:"notBefore"`
	NotAfter     time.Time  `json:"notAfter"`
	SANs         SANInfo  `json:"sans,omitempty"`
}

type ParseResponse struct {
	Success bool              `json:"success"`
	Certificate *CertificateInfo `json:"certificate,omitempty"`
	Error string            `json:"error,omitempty"`
}

type VerifyRequest struct {
	PEMChain string `json:"pemChain"`
}

type CertificateVerifyInfo struct {
	Index     int               `json:"index"`
	Info      *CertificateInfo `json:"info"`
	IsExpired bool             `json:"isExpired"`
	IsNotYetValid bool         `json:"isNotYetValid"`
	SignatureValid bool        `json:"signatureValid"`
}

type VerifyResponse struct {
	Success    bool                     `json:"success"`
	Valid      bool                     `json:"valid"`
	Error      string                   `json:"error,omitempty"`
	Certificates []*CertificateVerifyInfo `json:"certificates,omitempty"`
	Expired    bool                     `json:"expired"`
	NotYetValid bool                    `json:"notYetValid"`
}
