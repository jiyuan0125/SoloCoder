package license

type PlateType string

const (
	PlateTypeRegularSmall    PlateType = "普通小型车"
	PlateTypeRegularLarge    PlateType = "普通大型车"
	PlateTypeNewEnergyPure   PlateType = "新能源纯电"
	PlateTypeNewEnergyHybrid PlateType = "新能源混动"
	PlateTypeMilitary        PlateType = "军车"
	PlateTypeEmbassy         PlateType = "使领馆"
	PlateTypeArmedPolice     PlateType = "武警"
	PlateTypeInvalid         PlateType = "无效"
)

type ValidationResult struct {
	Valid   bool
	Type    PlateType
	Reason  string
	Details string
}

var provinces = map[rune]bool{
	'京': true, '津': true, '冀': true, '晋': true, '蒙': true,
	'辽': true, '吉': true, '黑': true, '沪': true, '苏': true,
	'浙': true, '皖': true, '闽': true, '赣': true, '鲁': true,
	'豫': true, '鄂': true, '湘': true, '粤': true, '桂': true,
	'琼': true, '川': true, '贵': true, '云': true, '渝': true,
	'藏': true, '陕': true, '甘': true, '青': true, '宁': true,
	'新': true,
}
