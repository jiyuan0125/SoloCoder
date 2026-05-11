package model

type ParseRequest struct {
	XML string `json:"xml"`
}

type Attribute struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type ParseResponse struct {
	Success    bool                   `json:"success"`
	Attributes []Attribute            `json:"attributes,omitempty"`
	Signature  SignatureInfo          `json:"signature,omitempty"`
	Subject    SubjectInfo            `json:"subject,omitempty"`
	Conditions ConditionsInfo         `json:"conditions,omitempty"`
	Raw        map[string]interface{} `json:"raw,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

type SignatureInfo struct {
	Algorithm      string `json:"algorithm"`
	DigestValue    string `json:"digestValue"`
	SignatureValue string `json:"signatureValue"`
	HasSignature   bool   `json:"hasSignature"`
}

type SubjectConfirmationDataInfo struct {
	Recipient    string `json:"recipient"`
	NotOnOrAfter string `json:"notOnOrAfter"`
	InResponseTo string `json:"inResponseTo"`
}

type SubjectConfirmationInfo struct {
	Method string                        `json:"method"`
	Data   SubjectConfirmationDataInfo   `json:"data"`
}

type SubjectInfo struct {
	NameID             string                   `json:"nameID"`
	SubjectConfirmation SubjectConfirmationInfo `json:"subjectConfirmation"`
}

type AudienceRestrictionInfo struct {
	Audiences []string `json:"audiences"`
}

type ConditionsInfo struct {
	NotBefore           string                   `json:"notBefore"`
	NotOnOrAfter        string                   `json:"notOnOrAfter"`
	AudienceRestriction AudienceRestrictionInfo  `json:"audienceRestriction"`
}

type ValidateRequest struct {
	XML       string `json:"xml"`
	Audience  string `json:"audience,omitempty"`
	Recipient string `json:"recipient,omitempty"`
}

type ValidateResponse struct {
	Success   bool     `json:"success"`
	Valid     bool     `json:"valid,omitempty"`
	Validated bool     `json:"validated,omitempty"`
	Errors    []string `json:"errors,omitempty"`
	Parse     ParseResponse `json:"parse,omitempty"`
}

type ExtractRequest struct {
	XML          string `json:"xml"`
	AttributeName string `json:"attributeName"`
}

type ExtractResponse struct {
	Success bool     `json:"success"`
	Values  []string `json:"values,omitempty"`
	Error   string   `json:"error,omitempty"`
}
