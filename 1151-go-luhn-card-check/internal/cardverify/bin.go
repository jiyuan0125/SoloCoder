package cardverify

var binData = map[string]BINInfo{
	"621226": {BIN: "621226", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622202": {BIN: "622202", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622203": {BIN: "622203", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621225": {BIN: "621225", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621558": {BIN: "621558", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621559": {BIN: "621559", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622230": {BIN: "622230", BankName: "中国工商银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622235": {BIN: "622235", BankName: "中国工商银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"628288": {BIN: "628288", BankName: "中国工商银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622208": {BIN: "622208", BankName: "中国工商银行", CardType: CardTypeDebit, AccountType: AccountDebit},

	"621700": {BIN: "621700", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622700": {BIN: "622700", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621081": {BIN: "621081", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621082": {BIN: "621082", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622280": {BIN: "622280", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621466": {BIN: "621466", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621467": {BIN: "621467", BankName: "中国建设银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628366": {BIN: "628366", BankName: "中国建设银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625964": {BIN: "625964", BankName: "中国建设银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625965": {BIN: "625965", BankName: "中国建设银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"621661": {BIN: "621661", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621660": {BIN: "621660", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621662": {BIN: "621662", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621663": {BIN: "621663", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621785": {BIN: "621785", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621786": {BIN: "621786", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621788": {BIN: "621788", BankName: "中国银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628388": {BIN: "628388", BankName: "中国银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625905": {BIN: "625905", BankName: "中国银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625906": {BIN: "625906", BankName: "中国银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622848": {BIN: "622848", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622843": {BIN: "622843", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622845": {BIN: "622845", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622846": {BIN: "622846", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"623052": {BIN: "623052", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621282": {BIN: "621282", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621283": {BIN: "621283", BankName: "中国农业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628268": {BIN: "628268", BankName: "中国农业银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625996": {BIN: "625996", BankName: "中国农业银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625998": {BIN: "625998", BankName: "中国农业银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622262": {BIN: "622262", BankName: "交通银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622260": {BIN: "622260", BankName: "交通银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621436": {BIN: "621436", BankName: "交通银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621437": {BIN: "621437", BankName: "交通银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622252": {BIN: "622252", BankName: "交通银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628216": {BIN: "628216", BankName: "交通银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622258": {BIN: "622258", BankName: "交通银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"625969": {BIN: "625969", BankName: "交通银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625970": {BIN: "625970", BankName: "交通银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"625971": {BIN: "625971", BankName: "交通银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"621483": {BIN: "621483", BankName: "招商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621485": {BIN: "621485", BankName: "招商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622588": {BIN: "622588", BankName: "招商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622598": {BIN: "622598", BankName: "招商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621486": {BIN: "621486", BankName: "招商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621489": {BIN: "621489", BankName: "招商银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628362": {BIN: "628362", BankName: "招商银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622575": {BIN: "622575", BankName: "招商银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622576": {BIN: "622576", BankName: "招商银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622582": {BIN: "622582", BankName: "招商银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"621799": {BIN: "621799", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622150": {BIN: "622150", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622151": {BIN: "622151", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621798": {BIN: "621798", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621797": {BIN: "621797", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628310": {BIN: "628310", BankName: "中国邮政储蓄银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622188": {BIN: "622188", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"620529": {BIN: "620529", BankName: "中国邮政储蓄银行", CardType: CardTypeDebit, AccountType: AccountDebit},

	"622690": {BIN: "622690", BankName: "中信银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622691": {BIN: "622691", BankName: "中信银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621771": {BIN: "621771", BankName: "中信银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621772": {BIN: "621772", BankName: "中信银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628206": {BIN: "628206", BankName: "中信银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622696": {BIN: "622696", BankName: "中信银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622698": {BIN: "622698", BankName: "中信银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622616": {BIN: "622616", BankName: "中国民生银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622615": {BIN: "622615", BankName: "中国民生银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621691": {BIN: "621691", BankName: "中国民生银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621692": {BIN: "621692", BankName: "中国民生银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628258": {BIN: "628258", BankName: "中国民生银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622621": {BIN: "622621", BankName: "中国民生银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622622": {BIN: "622622", BankName: "中国民生银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622556": {BIN: "622556", BankName: "华夏银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622632": {BIN: "622632", BankName: "华夏银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621222": {BIN: "621222", BankName: "华夏银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628212": {BIN: "628212", BankName: "华夏银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622636": {BIN: "622636", BankName: "华夏银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622637": {BIN: "622637", BankName: "华夏银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622270": {BIN: "622270", BankName: "上海浦东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621795": {BIN: "621795", BankName: "上海浦东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621796": {BIN: "621796", BankName: "上海浦东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621793": {BIN: "621793", BankName: "上海浦东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628222": {BIN: "628222", BankName: "上海浦东发展银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622177": {BIN: "622177", BankName: "上海浦东发展银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622176": {BIN: "622176", BankName: "上海浦东发展银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622660": {BIN: "622660", BankName: "中国光大银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622661": {BIN: "622661", BankName: "中国光大银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621492": {BIN: "621492", BankName: "中国光大银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621488": {BIN: "621488", BankName: "中国光大银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628201": {BIN: "628201", BankName: "中国光大银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622665": {BIN: "622665", BankName: "中国光大银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622666": {BIN: "622666", BankName: "中国光大银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622820": {BIN: "622820", BankName: "广东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622568": {BIN: "622568", BankName: "广东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621462": {BIN: "621462", BankName: "广东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621481": {BIN: "621481", BankName: "广东发展银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628259": {BIN: "628259", BankName: "广东发展银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622557": {BIN: "622557", BankName: "广东发展银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622558": {BIN: "622558", BankName: "广东发展银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622761": {BIN: "622761", BankName: "平安银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622298": {BIN: "622298", BankName: "平安银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621626": {BIN: "621626", BankName: "平安银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"623058": {BIN: "623058", BankName: "平安银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621602": {BIN: "621602", BankName: "平安银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628296": {BIN: "628296", BankName: "平安银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622156": {BIN: "622156", BankName: "平安银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622155": {BIN: "622155", BankName: "平安银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"621418": {BIN: "621418", BankName: "兴业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622908": {BIN: "622908", BankName: "兴业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622909": {BIN: "622909", BankName: "兴业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621784": {BIN: "621784", BankName: "兴业银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628491": {BIN: "628491", BankName: "兴业银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622901": {BIN: "622901", BankName: "兴业银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622902": {BIN: "622902", BankName: "兴业银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622962": {BIN: "622962", BankName: "北京银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621245": {BIN: "621245", BankName: "北京银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621414": {BIN: "621414", BankName: "北京银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628210": {BIN: "628210", BankName: "北京银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622163": {BIN: "622163", BankName: "北京银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"622888": {BIN: "622888", BankName: "上海银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"622178": {BIN: "622178", BankName: "上海银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"621709": {BIN: "621709", BankName: "上海银行", CardType: CardTypeDebit, AccountType: AccountDebit},
	"628320": {BIN: "628320", BankName: "上海银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622286": {BIN: "622286", BankName: "上海银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622278": {BIN: "622278", BankName: "上海银行", CardType: CardTypeCredit, AccountType: AccountCredit},
	"622014": {BIN: "622014", BankName: "上海银行", CardType: CardTypeCredit, AccountType: AccountCredit},

	"428318": {BIN: "428318", BankName: "上海银行", CardType: CardTypeCredit, AccountType: AccountCredit},
}

func LookupBIN(number string) (*BINInfo, bool) {
	if len(number) < 6 {
		return nil, false
	}

	for length := 8; length >= 6; length-- {
		if len(number) < length {
			continue
		}
		prefix := number[:length]
		if info, exists := binData[prefix]; exists {
			copyInfo := info
			return &copyInfo, true
		}
	}

	if len(number) >= 6 {
		prefix6 := number[:6]
		if info, exists := binData[prefix6]; exists {
			copyInfo := info
			return &copyInfo, true
		}
	}

	return nil, false
}
