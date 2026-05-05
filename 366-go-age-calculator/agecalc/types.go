package agecalc

import "errors"

var (
	ErrBirthDateAfterCurrent = errors.New("出生日期不能晚于当前日期")
)

type Age struct {
	Years  int
	Months int
	Days   int
}

type Gender int

const (
	Male Gender = iota
	Female
)

func (g Gender) String() string {
	switch g {
	case Male:
		return "男"
	case Female:
		return "女"
	default:
		return "未知"
	}
}
