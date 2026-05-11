package common

const (
	CaseNoPrefix        = "CL"
	MaxHoursForFullPayout = 48
	MaxClaimsPerYear    = 3
	FreqDiscountRatio   = 0.80
	TimeDiscountRatio   = 0.70
)

func (at AccidentType) String() string {
	switch at {
	case AccidentTypeTraffic:
		return "交通事故"
	case AccidentTypeNatural:
		return "自然灾害"
	case AccidentTypeTheft:
		return "盗窃"
	case AccidentTypeGlass:
		return "玻璃单独破碎"
	case AccidentTypeScratch:
		return "划痕"
	default:
		return "未知"
	}
}

func (lt LiabilityType) String() string {
	switch lt {
	case LiabilityFull:
		return "全责"
	case LiabilityMajor:
		return "主责"
	case LiabilityEqual:
		return "同责"
	case LiabilityMinor:
		return "次责"
	case LiabilityNone:
		return "无责"
	default:
		return "未知"
	}
}

func (cs CaseStatus) String() string {
	switch cs {
	case CaseStatusSubmitted:
		return "已提交"
	case CaseStatusAssigned:
		return "已分配"
	case CaseStatusInvestigating:
		return "查勘中"
	case CaseStatusInvestigated:
		return "查勘完成"
	case CaseStatusAssessing:
		return "定损中"
	case CaseStatusAssessed:
		return "定损完成"
	case CaseStatusApproved:
		return "已审批"
	case CaseStatusPaid:
		return "已赔付"
	case CaseStatusReviewing:
		return "复核中"
	case CaseStatusRejected:
		return "已拒绝"
	default:
		return "未知"
	}
}

func (rt RepairType) String() string {
	switch rt {
	case RepairTypeRepair:
		return "维修"
	case RepairTypeReplace:
		return "更换"
	default:
		return "未知"
	}
}
