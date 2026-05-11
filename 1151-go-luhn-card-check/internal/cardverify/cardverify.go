package cardverify

type CardType string
type CardOrganization string
type AccountType string

const (
	CardTypeDebit     CardType = "借记卡"
	CardTypeCredit    CardType = "贷记卡"
	CardTypeQuasiDebit CardType = "准贷记卡"
	CardTypePrepaid   CardType = "预付卡"
	CardTypeUnknown   CardType = "未知"
)

const (
	OrgVisa          CardOrganization = "Visa"
	OrgMasterCard     CardOrganization = "MasterCard"
	OrgUnionPay       CardOrganization = "银联"
	OrgJCB            CardOrganization = "JCB"
	OrgAmex           CardOrganization = "American Express"
	OrgUnknown         CardOrganization = "未知"
)

const (
	AccountDebit     AccountType = "借记卡"
	AccountCredit    AccountType = "贷记卡"
	AccountQuasiDebit AccountType = "准贷记卡"
	AccountPrepaid   AccountType = "预付卡"
	AccountUnknown   AccountType = "未知"
)

type CardInfo struct {
	RawNumber     string
	CleanNumber    string
	IsValid          bool
	ErrorReason      string
	CardOrganization CardOrganization
	BankName         string
	CardType         CardType
	BIN             string
}

type ValidateResult struct {
	CardInfo
}

type BINInfo struct {
	BIN         string
	BankName    string
	CardType    CardType
	AccountType AccountType
}
