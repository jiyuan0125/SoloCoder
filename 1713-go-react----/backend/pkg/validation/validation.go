package validation

import (
	"regexp"
	"strconv"
	"strings"
)

var usccRegex = regexp.MustCompile(`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`)

var hazardLevels = map[string]bool{
	"一般": true,
	"较重": true,
	"严重": true,
}

var validCategories = map[string]bool{
	"粉尘类":    true,
	"化学毒物类":  true,
	"物理因素类":  true,
	"生物因素类":  true,
}

var factorToItems = map[string][]struct{ Code, Name string }{
	"矽尘":   {{"CHEST_XRAY", "胸部X光检查"}, {"PFT", "肺功能检查"}},
	"煤尘":   {{"CHEST_XRAY", "胸部X光检查"}, {"PFT", "肺功能检查"}},
	"石棉尘":  {{"CHEST_XRAY", "胸部X光检查"}, {"PFT", "肺功能检查"}},
	"苯":    {{"BLOOD", "血常规"}, {"LIVER", "肝功能"}, {"URINE", "尿常规"}},
	"铅":    {{"BLOOD_PB", "血铅检测"}, {"URINE_PB", "尿铅检测"}},
	"汞":    {{"BLOOD_HG", "血汞检测"}, {"URINE_HG", "尿汞检测"}},
	"甲醛":   {{"BLOOD", "血常规"}, {"LIVER", "肝功能"}},
	"噪声":   {{"AUDIO", "纯音听力测试"}},
	"高温":   {{"BP", "血压"}, {"ECG", "心电图"}},
	"辐射":   {{"BLOOD", "血常规"}, {"CHROMOSOME", "染色体检查"}},
	"布鲁氏菌": {{"BLOOD", "血常规"}, {"SERUM", "血清学检测"}},
}

var categoryToCycle = map[string]int{
	"粉尘类":   12,
	"化学毒物类": 12,
	"物理因素类": 12,
	"生物因素类": 12,
}

var specialFactorCycle = map[string]int{
	"辐射": 6,
}

var exposureLimits = map[string]float64{
	"矽尘":   1.0,
	"煤尘":   4.0,
	"石棉尘":  0.8,
	"苯":    6.0,
	"铅":    0.05,
	"汞":    0.02,
	"甲醛":   0.5,
	"噪声":   85.0,
	"高温":   28.0,
	"辐射":   20.0,
	"布鲁氏菌": 0.0,
}

var factorUnits = map[string]string{
	"矽尘":   "mg/m³",
	"煤尘":   "mg/m³",
	"石棉尘":  "f/mL",
	"苯":    "mg/m³",
	"铅":    "mg/m³",
	"汞":    "mg/m³",
	"甲醛":   "mg/m³",
	"噪声":   "dB(A)",
	"高温":   "°C WBGT",
	"辐射":   "mSv/年",
	"布鲁氏菌": "-",
}

func IsValidUSCC(code string) bool {
	if len(code) != 18 {
		return false
	}
	if !usccRegex.MatchString(code) {
		return false
	}
	return verifyUSCCChecksum(code)
}

func GenerateTestUSCC() string {
	chars := "0123456789ABCDEFGHJKLMNPQRTUWXY"
	weights := []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
	charValues := make(map[rune]int)
	for i, c := range chars {
		charValues[c] = i
	}

	code := "91110108MA0084XN1"
	sum := 0
	for i := 0; i < 17; i++ {
		c := rune(code[i])
		val := charValues[c]
		sum += val * weights[i]
	}

	mod := 31
	checkVal := mod - (sum % mod)
	if checkVal == mod {
		checkVal = 0
	}
	return code + string(chars[checkVal])
}

func verifyUSCCChecksum(code string) bool {
	weights := []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
	chars := "0123456789ABCDEFGHJKLMNPQRTUWXY"
	charValues := make(map[rune]int)
	for i, c := range chars {
		charValues[c] = i
	}

	sum := 0
	for i := 0; i < 17; i++ {
		c := rune(code[i])
		val, ok := charValues[c]
		if !ok {
			return false
		}
		sum += val * weights[i]
	}

	mod := 31
	checkVal := mod - (sum % mod)
	if checkVal == mod {
		checkVal = 0
	}

	expectedCheck := chars[checkVal]
	return code[17] == expectedCheck
}

func IsValidHazardLevel(level string) bool {
	return hazardLevels[level]
}

func IsValidCategory(cat string) bool {
	return validCategories[cat]
}

func GetExposureLimit(factor string) (float64, string) {
	limit, ok := exposureLimits[factor]
	if !ok {
		return 0, ""
	}
	unit := factorUnits[factor]
	return limit, unit
}

func IsExceedingLimit(factor string, value float64) bool {
	limit, _ := GetExposureLimit(factor)
	if limit == 0 {
		return false
	}
	return value > limit
}

func GetExamItemsForFactors(factors []string) []struct{ Code, Name string } {
	itemMap := make(map[string]struct{ Code, Name string })
	for _, f := range factors {
		if items, ok := factorToItems[f]; ok {
			for _, item := range items {
				itemMap[item.Code] = item
			}
		}
	}

	result := make([]struct{ Code, Name string }, 0, len(itemMap))
	for _, item := range itemMap {
		result = append(result, item)
	}
	return result
}

func IsValidExamItemCode(code string, factorNames []string) bool {
	validCodes := make(map[string]bool)
	for _, f := range factorNames {
		if items, ok := factorToItems[f]; ok {
			for _, item := range items {
				validCodes[item.Code] = true
			}
		}
	}
	return validCodes[code]
}

func CalculateExamCycleMonths(categories []string, factorNames []string) int {
	minCycle := 12

	for _, name := range factorNames {
		if cycle, ok := specialFactorCycle[name]; ok {
			if cycle < minCycle {
				minCycle = cycle
			}
		}
	}

	for _, cat := range categories {
		if cycle, ok := categoryToCycle[cat]; ok {
			if cycle < minCycle {
				minCycle = cycle
			}
		}
	}

	return minCycle
}

func ParseMonitorValue(valueStr string) (float64, error) {
	valueStr = strings.TrimSpace(valueStr)
	return strconv.ParseFloat(valueStr, 64)
}

func GetAllExamItemCodes() map[string]string {
	result := make(map[string]string)
	for _, items := range factorToItems {
		for _, item := range items {
			result[item.Code] = item.Name
		}
	}
	return result
}

func GetFactorUnit(factor string) string {
	return factorUnits[factor]
}
