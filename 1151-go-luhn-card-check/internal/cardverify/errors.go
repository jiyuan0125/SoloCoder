package cardverify

import "errors"

var (
	ErrEmptyCardNumber    = errors.New("银行卡号不能为空")
	ErrInvalidCharacter  = errors.New("卡号包含无效字符，仅允许数字和指定分隔符(空格、横杠、下划线)")
	ErrInvalidLength     = errors.New("银行卡号长度不符合标准长度(13-19位)")
	ErrLuhnCheckFailed   = errors.New("Luhn校验失败，卡号无效")
	ErrInvalidCardNumber  = errors.New("银行卡号无效")
)
