package mask

type DataType string

const (
	Phone       DataType = "phone"
	IDCard      DataType = "idcard"
	BankCard    DataType = "bankcard"
	Email       DataType = "email"
	Name        DataType = "name"
)

type MaskRule struct {
	PrefixKeep int `json:"prefix_keep"`
	SuffixKeep int `json:"suffix_keep"`
}

type DefaultRules struct {
	Phone    *MaskRule `json:"phone"`
	IDCard   *MaskRule `json:"idcard"`
	BankCard *MaskRule `json:"bankcard"`
	Email    *MaskRule `json:"email"`
	Name     *MaskRule `json:"name"`
}

func NewDefaultRules() *DefaultRules {
	return &DefaultRules{
		Phone:    &MaskRule{PrefixKeep: 3, SuffixKeep: 4},
		IDCard:   &MaskRule{PrefixKeep: 3, SuffixKeep: 4},
		BankCard: &MaskRule{PrefixKeep: 0, SuffixKeep: 4},
		Email:    &MaskRule{PrefixKeep: 1, SuffixKeep: 0},
		Name:     &MaskRule{PrefixKeep: 1, SuffixKeep: 0},
	}
}

func (r *DefaultRules) GetRule(dataType DataType) *MaskRule {
	switch dataType {
	case Phone:
		return r.Phone
	case IDCard:
		return r.IDCard
	case BankCard:
		return r.BankCard
	case Email:
		return r.Email
	case Name:
		return r.Name
	default:
		return nil
	}
}
