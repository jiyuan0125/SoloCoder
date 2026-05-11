package models

type CardOrganization string
type CardType string

const (
	CardOrgVisa          CardOrganization = "Visa"
	CardOrgMasterCard     CardOrganization = "MasterCard"
	CardOrgUnionPay       CardOrganization = "银联"
	CardOrgJCB            CardOrganization = "JCB"
	CardOrgAmex           CardOrganization = "American Express"
	CardOrgUnknown         CardOrganization = "未知"
)

const (
	CardTypeDebit     CardType = "借记卡"
	CardTypeCredit    CardType = "贷记卡"
	CardTypeQuasiDebit CardType = "准贷记卡"
	CardTypePrepaid   CardType = "预付卡"
	CardTypeUnknown   CardType = "未知"
)

type ValidateRequest struct {
	CardNumber string `json:"card_number"`
}

type ValidateResponse struct {
	Success          bool             `json:"success"`
	CardNumber       string           `json:"card_number"`
	CleanNumber      string           `json:"clean_number,omitempty"`
	IsValid          bool             `json:"is_valid"`
	ErrorReason      string           `json:"error_reason,omitempty"`
	CardOrganization CardOrganization `json:"card_organization,omitempty"`
	BankName         string           `json:"bank_name,omitempty"`
	CardType         CardType         `json:"card_type,omitempty"`
	BIN             string           `json:"bin,omitempty"`
}

type BatchValidateRequest struct {
	CardNumbers []string `json:"card_numbers"`
}

type BatchValidateResponse struct {
	Success bool               `json:"success"`
	Results []ValidateResponse `json:"results"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
