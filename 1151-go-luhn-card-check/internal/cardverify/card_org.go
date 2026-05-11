package cardverify

import (
	"strconv"
	"strings"
)

func DetectCardOrganization(number string) CardOrganization {
	if len(number) < 2 {
		return OrgUnknown
	}

	if strings.HasPrefix(number, "4") {
		return OrgVisa
	}

	if len(number) >= 2 {
		prefix2, _ := strconv.Atoi(number[:2])
		if prefix2 >= 51 && prefix2 <= 55 {
			return OrgMasterCard
		}
	}

	if len(number) >= 4 {
		prefix4, _ := strconv.Atoi(number[:4])
		if prefix4 >= 2221 && prefix4 <= 2720 {
			return OrgMasterCard
		}
	}

	if strings.HasPrefix(number, "62") {
		return OrgUnionPay
	}

	if strings.HasPrefix(number, "35") {
		return OrgJCB
	}

	if strings.HasPrefix(number, "34") || strings.HasPrefix(number, "37") {
		return OrgAmex
	}

	return OrgUnknown
}
