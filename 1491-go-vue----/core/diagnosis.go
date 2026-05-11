package core

import (
	"strings"

	"repair-platform/common"
)

var faultRules = map[common.ApplianceCategory][]faultRuleEntry{
	common.AirConditioner: {
		{keywords: []string{"不制冷", "制冷差"}, causes: []common.PossibleCause{
			{Cause: "缺氟", MinEstimateFee: 150, MaxEstimateFee: 300},
			{Cause: "压缩机故障", MinEstimateFee: 800, MaxEstimateFee: 1500},
			{Cause: "滤网堵塞", MinEstimateFee: 50, MaxEstimateFee: 100},
		}},
		{keywords: []string{"不启动", "不开机"}, causes: []common.PossibleCause{
			{Cause: "电源故障", MinEstimateFee: 50, MaxEstimateFee: 200},
			{Cause: "遥控器故障", MinEstimateFee: 30, MaxEstimateFee: 80},
		}},
		{keywords: []string{"漏水"}, causes: []common.PossibleCause{
			{Cause: "排水管堵塞", MinEstimateFee: 50, MaxEstimateFee: 150},
			{Cause: "安装问题", MinEstimateFee: 100, MaxEstimateFee: 300},
		}},
	},
	common.Refrigerator: {
		{keywords: []string{"不制冷", "制冷差"}, causes: []common.PossibleCause{
			{Cause: "制冷剂泄漏", MinEstimateFee: 200, MaxEstimateFee: 500},
			{Cause: "压缩机故障", MinEstimateFee: 800, MaxEstimateFee: 2000},
		}},
		{keywords: []string{"异响"}, causes: []common.PossibleCause{
			{Cause: "风扇故障", MinEstimateFee: 100, MaxEstimateFee: 300},
		}},
	},
	common.WashingMachine: {
		{keywords: []string{"不转", "不洗衣"}, causes: []common.PossibleCause{
			{Cause: "电机故障", MinEstimateFee: 300, MaxEstimateFee: 800},
			{Cause: "皮带断裂", MinEstimateFee: 80, MaxEstimateFee: 200},
		}},
		{keywords: []string{"漏水"}, causes: []common.PossibleCause{
			{Cause: "排水管老化", MinEstimateFee: 50, MaxEstimateFee: 150},
			{Cause: "密封圈损坏", MinEstimateFee: 100, MaxEstimateFee: 300},
		}},
	},
	common.TV: {
		{keywords: []string{"不开机", "黑屏"}, causes: []common.PossibleCause{
			{Cause: "电源板故障", MinEstimateFee: 200, MaxEstimateFee: 500},
			{Cause: "主板故障", MinEstimateFee: 500, MaxEstimateFee: 1500},
		}},
		{keywords: []string{"无声音"}, causes: []common.PossibleCause{
			{Cause: "扬声器故障", MinEstimateFee: 100, MaxEstimateFee: 300},
		}},
	},
	common.WaterHeater: {
		{keywords: []string{"不加热"}, causes: []common.PossibleCause{
			{Cause: "加热管损坏", MinEstimateFee: 200, MaxEstimateFee: 500},
			{Cause: "温控器故障", MinEstimateFee: 100, MaxEstimateFee: 300},
		}},
		{keywords: []string{"漏水"}, causes: []common.PossibleCause{
			{Cause: "内胆泄漏", MinEstimateFee: 300, MaxEstimateFee: 800},
			{Cause: "接口松动", MinEstimateFee: 50, MaxEstimateFee: 150},
		}},
	},
	common.RangeHood: {
		{keywords: []string{"不吸烟", "吸力小"}, causes: []common.PossibleCause{
			{Cause: "电机故障", MinEstimateFee: 200, MaxEstimateFee: 500},
			{Cause: "滤网堵塞", MinEstimateFee: 50, MaxEstimateFee: 100},
		}},
	},
	common.GasStove: {
		{keywords: []string{"打不着火"}, causes: []common.PossibleCause{
			{Cause: "点火器故障", MinEstimateFee: 80, MaxEstimateFee: 200},
			{Cause: "电池没电", MinEstimateFee: 20, MaxEstimateFee: 50},
		}},
	},
}

type faultRuleEntry struct {
	keywords []string
	causes   []common.PossibleCause
}

func (s *Service) DiagnoseFault(cat common.ApplianceCategory, desc string) *common.FaultDiagnosis {
	descLower := strings.ToLower(desc)

	rules, ok := faultRules[cat]
	if !ok {
		return &common.FaultDiagnosis{PossibleCauses: []common.PossibleCause{
			{Cause: "通用故障检测", MinEstimateFee: 100, MaxEstimateFee: 300},
		}}
	}

	resultSet := make(map[string]common.PossibleCause)
	matchedAny := false

	for _, rule := range rules {
		for _, kw := range rule.keywords {
			if strings.Contains(descLower, kw) {
				matchedAny = true
				for _, c := range rule.causes {
					resultSet[c.Cause] = c
				}
				break
			}
		}
	}

	causes := make([]common.PossibleCause, 0, len(resultSet))
	for _, c := range resultSet {
		causes = append(causes, c)
	}

	if !matchedAny || len(causes) == 0 {
		causes = []common.PossibleCause{
			{Cause: "初步检测", MinEstimateFee: 100, MaxEstimateFee: 300},
		}
	}

	return &common.FaultDiagnosis{PossibleCauses: causes}
}
