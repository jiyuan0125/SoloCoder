package cardparser

import (
	"strconv"
)

type BINInfo struct {
	StartBIN string
	EndBIN   string
	BankName string
	CardType CardType
}

var binData = []BINInfo{
	// 工商银行
	{StartBIN: "622200", EndBIN: "622202", BankName: "工商银行", CardType: CardTypeDebit},
	{StartBIN: "622208", EndBIN: "622208", BankName: "工商银行", CardType: CardTypeDebit},
	{StartBIN: "621225", EndBIN: "621226", BankName: "工商银行", CardType: CardTypeDebit},
	{StartBIN: "621721", EndBIN: "621723", BankName: "工商银行", CardType: CardTypeDebit},
	{StartBIN: "427010", EndBIN: "427029", BankName: "工商银行", CardType: CardTypeCredit},
	{StartBIN: "427030", EndBIN: "427039", BankName: "工商银行", CardType: CardTypeDebit},
	{StartBIN: "458060", EndBIN: "458069", BankName: "工商银行", CardType: CardTypeCredit},
	{StartBIN: "530970", EndBIN: "530999", BankName: "工商银行", CardType: CardTypeCredit},
	{StartBIN: "370246", EndBIN: "370248", BankName: "工商银行", CardType: CardTypeCredit},
	{StartBIN: "370249", EndBIN: "370249", BankName: "工商银行", CardType: CardTypeCredit},

	// 建设银行
	{StartBIN: "622280", EndBIN: "622280", BankName: "建设银行", CardType: CardTypeDebit},
	{StartBIN: "621700", EndBIN: "621700", BankName: "建设银行", CardType: CardTypeDebit},
	{StartBIN: "623668", EndBIN: "623668", BankName: "建设银行", CardType: CardTypeDebit},
	{StartBIN: "620060", EndBIN: "620060", BankName: "建设银行", CardType: CardTypeCredit},
	{StartBIN: "532450", EndBIN: "532458", BankName: "建设银行", CardType: CardTypeCredit},
	{StartBIN: "436742", EndBIN: "436742", BankName: "建设银行", CardType: CardTypeDebit},
	{StartBIN: "436745", EndBIN: "436748", BankName: "建设银行", CardType: CardTypeCredit},
	{StartBIN: "526410", EndBIN: "526410", BankName: "建设银行", CardType: CardTypeCredit},
	{StartBIN: "552801", EndBIN: "552801", BankName: "建设银行", CardType: CardTypeCredit},
	{StartBIN: "491031", EndBIN: "491031", BankName: "建设银行", CardType: CardTypeCredit},

	// 农业银行
	{StartBIN: "622848", EndBIN: "622849", BankName: "农业银行", CardType: CardTypeDebit},
	{StartBIN: "621336", EndBIN: "621336", BankName: "农业银行", CardType: CardTypeDebit},
	{StartBIN: "621618", EndBIN: "621618", BankName: "农业银行", CardType: CardTypeDebit},
	{StartBIN: "622820", EndBIN: "622820", BankName: "农业银行", CardType: CardTypeDebit},
	{StartBIN: "622836", EndBIN: "622837", BankName: "农业银行", CardType: CardTypeCredit},
	{StartBIN: "628268", EndBIN: "628268", BankName: "农业银行", CardType: CardTypeCredit},
	{StartBIN: "403361", EndBIN: "403361", BankName: "农业银行", CardType: CardTypeCredit},
	{StartBIN: "404117", EndBIN: "404117", BankName: "农业银行", CardType: CardTypeCredit},
	{StartBIN: "461900", EndBIN: "461900", BankName: "农业银行", CardType: CardTypeCredit},
	{StartBIN: "519412", EndBIN: "519413", BankName: "农业银行", CardType: CardTypeCredit},

	// 中国银行
	{StartBIN: "621660", EndBIN: "621661", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "621725", EndBIN: "621726", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "621785", EndBIN: "621786", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "621790", EndBIN: "621790", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "623208", EndBIN: "623208", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "456351", EndBIN: "456351", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "601382", EndBIN: "601382", BankName: "中国银行", CardType: CardTypeDebit},
	{StartBIN: "512411", EndBIN: "512412", BankName: "中国银行", CardType: CardTypeCredit},
	{StartBIN: "514957", EndBIN: "514958", BankName: "中国银行", CardType: CardTypeCredit},
	{StartBIN: "409665", EndBIN: "409666", BankName: "中国银行", CardType: CardTypeCredit},
	{StartBIN: "409668", EndBIN: "409669", BankName: "中国银行", CardType: CardTypeCredit},
	{StartBIN: "377677", EndBIN: "377677", BankName: "中国银行", CardType: CardTypeCredit},

	// 招商银行
	{StartBIN: "622609", EndBIN: "622609", BankName: "招商银行", CardType: CardTypeDebit},
	{StartBIN: "621483", EndBIN: "621483", BankName: "招商银行", CardType: CardTypeDebit},
	{StartBIN: "621485", EndBIN: "621485", BankName: "招商银行", CardType: CardTypeDebit},
	{StartBIN: "621486", EndBIN: "621486", BankName: "招商银行", CardType: CardTypeDebit},
	{StartBIN: "622580", EndBIN: "622582", BankName: "招商银行", CardType: CardTypeDebit},
	{StartBIN: "439188", EndBIN: "439188", BankName: "招商银行", CardType: CardTypeCredit},
	{StartBIN: "356889", EndBIN: "356890", BankName: "招商银行", CardType: CardTypeCredit},
	{StartBIN: "356895", EndBIN: "356895", BankName: "招商银行", CardType: CardTypeCredit},
	{StartBIN: "512425", EndBIN: "512425", BankName: "招商银行", CardType: CardTypeCredit},
	{StartBIN: "518710", EndBIN: "518710", BankName: "招商银行", CardType: CardTypeCredit},
	{StartBIN: "622575", EndBIN: "622576", BankName: "招商银行", CardType: CardTypeCredit},
	{StartBIN: "622578", EndBIN: "622578", BankName: "招商银行", CardType: CardTypeCredit},

	// 交通银行
	{StartBIN: "622260", EndBIN: "622262", BankName: "交通银行", CardType: CardTypeDebit},
	{StartBIN: "622258", EndBIN: "622259", BankName: "交通银行", CardType: CardTypeDebit},
	{StartBIN: "621797", EndBIN: "621797", BankName: "交通银行", CardType: CardTypeDebit},
	{StartBIN: "622261", EndBIN: "622261", BankName: "交通银行", CardType: CardTypeDebit},
	{StartBIN: "458123", EndBIN: "458124", BankName: "交通银行", CardType: CardTypeCredit},
	{StartBIN: "520169", EndBIN: "520169", BankName: "交通银行", CardType: CardTypeCredit},
	{StartBIN: "522964", EndBIN: "522964", BankName: "交通银行", CardType: CardTypeCredit},
	{StartBIN: "622250", EndBIN: "622251", BankName: "交通银行", CardType: CardTypeCredit},
	{StartBIN: "622253", EndBIN: "622254", BankName: "交通银行", CardType: CardTypeCredit},

	// 邮储银行
	{StartBIN: "621798", EndBIN: "621799", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "621095", EndBIN: "621096", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "621098", EndBIN: "621098", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "622150", EndBIN: "622150", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "622151", EndBIN: "622151", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "622188", EndBIN: "622188", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "621622", EndBIN: "621622", BankName: "邮储银行", CardType: CardTypeDebit},
	{StartBIN: "623698", EndBIN: "623698", BankName: "邮储银行", CardType: CardTypeCredit},
	{StartBIN: "625919", EndBIN: "625919", BankName: "邮储银行", CardType: CardTypeCredit},
	{StartBIN: "622810", EndBIN: "622812", BankName: "邮储银行", CardType: CardTypeCredit},
}

func LookupBankAndType(cardNumber string) (string, CardType) {
	lengthsToTry := []int{6, 5, 4}

	for _, length := range lengthsToTry {
		if len(cardNumber) >= length {
			prefix := cardNumber[:length]
			for _, info := range binData {
				if binMatchesPrefix(prefix, info, length) {
					return info.BankName, info.CardType
				}
			}
		}
	}

	return "未知银行", CardTypeUnknown
}

func binMatchesPrefix(prefix string, info BINInfo, length int) bool {
	startPrefix := info.StartBIN[:length]
	endPrefix := info.EndBIN[:length]

	prefixNum, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return false
	}

	startNum, err := strconv.ParseInt(startPrefix, 10, 64)
	if err != nil {
		return false
	}

	endNum, err := strconv.ParseInt(endPrefix, 10, 64)
	if err != nil {
		return false
	}

	return prefixNum >= startNum && prefixNum <= endNum
}
