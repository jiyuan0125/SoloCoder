package address

import (
	"errors"
	"shipping-cost/pkg/config"
)

var ErrAddressNotFound = errors.New("无法识别该地址")

type Zone int

const (
	ZoneLocal Zone = iota
	ZoneIntra
	ZoneNeighbor
	ZoneInter
	ZoneRemote
)

func (z Zone) String() string {
	switch z {
	case ZoneLocal:
		return "同城"
	case ZoneIntra:
		return "省内"
	case ZoneNeighbor:
		return "邻省"
	case ZoneInter:
		return "跨区"
	case ZoneRemote:
		return "偏远地区"
	default:
		return "未知"
	}
}

func Resolve(address string) (*config.Address, error) {
	cfg := config.Get()
	addr, exists := cfg.AddressDB[address]
	if !exists {
		return nil, ErrAddressNotFound
	}
	return &addr, nil
}

func DetermineZone(sender, receiver *config.Address) Zone {
	cfg := config.Get()

	for _, area := range cfg.RemoteAreas {
		if receiver.Province == area || receiver.City == area {
			return ZoneRemote
		}
	}

	if sender.Province == receiver.Province {
		if sender.City == receiver.City {
			return ZoneLocal
		}
		return ZoneIntra
	}

	if isNeighbor(sender.Province, receiver.Province) {
		return ZoneNeighbor
	}

	return ZoneInter
}

var neighborMap = map[string][]string{
	"北京市": {"天津市", "河北省"},
	"天津市": {"北京市", "河北省"},
	"上海市": {"江苏省", "浙江省"},
	"重庆市": {"四川省", "贵州省", "湖南省", "湖北省", "陕西省"},
	"广东省": {"广西壮族自治区", "湖南省", "江西省", "福建省", "海南省"},
	"江苏省": {"上海市", "浙江省", "安徽省", "山东省", "河南省"},
	"浙江省": {"上海市", "江苏省", "安徽省", "江西省", "福建省"},
	"山东省": {"河北省", "河南省", "安徽省", "江苏省"},
	"河南省": {"山东省", "河北省", "山西省", "陕西省", "湖北省", "安徽省", "江苏省"},
	"湖北省": {"河南省", "湖南省", "江西省", "安徽省", "重庆市", "陕西省"},
	"湖南省": {"湖北省", "江西省", "广东省", "广西壮族自治区", "贵州省", "重庆市"},
	"四川省": {"重庆市", "贵州省", "云南省", "西藏自治区", "青海省", "甘肃省", "陕西省"},
	"河北省": {"北京市", "天津市", "山西省", "内蒙古自治区", "山东省", "河南省"},
	"山西省": {"河北省", "内蒙古自治区", "陕西省", "河南省"},
	"辽宁省": {"吉林省", "内蒙古自治区", "河北省"},
	"吉林省": {"黑龙江省", "内蒙古自治区", "辽宁省"},
	"黑龙江省": {"内蒙古自治区", "吉林省"},
	"安徽省": {"江苏省", "浙江省", "江西省", "湖北省", "河南省", "山东省"},
	"福建省": {"浙江省", "江西省", "广东省", "台湾省"},
	"江西省": {"浙江省", "安徽省", "湖北省", "湖南省", "广东省", "福建省"},
	"广西壮族自治区": {"广东省", "湖南省", "贵州省", "云南省"},
	"贵州省": {"湖南省", "广西壮族自治区", "云南省", "四川省", "重庆市"},
	"云南省": {"贵州省", "广西壮族自治区", "西藏自治区", "四川省"},
	"西藏自治区": {"四川省", "云南省", "青海省", "新疆维吾尔自治区"},
	"陕西省": {"内蒙古自治区", "山西省", "河南省", "湖北省", "重庆市", "四川省", "甘肃省", "宁夏回族自治区"},
	"甘肃省": {"内蒙古自治区", "宁夏回族自治区", "陕西省", "青海省", "新疆维吾尔自治区", "四川省"},
	"青海省": {"甘肃省", "西藏自治区", "新疆维吾尔自治区", "四川省"},
	"宁夏回族自治区": {"内蒙古自治区", "甘肃省", "陕西省"},
	"新疆维吾尔自治区": {"甘肃省", "青海省", "西藏自治区"},
	"内蒙古自治区": {"黑龙江省", "吉林省", "辽宁省", "河北省", "山西省", "陕西省", "宁夏回族自治区", "甘肃省"},
	"海南省": {"广东省"},
	"香港特别行政区": {"广东省"},
	"澳门特别行政区": {"广东省"},
	"台湾省": {"福建省"},
}

func isNeighbor(provinceA, provinceB string) bool {
	neighbors, exists := neighborMap[provinceA]
	if !exists {
		return false
	}
	for _, n := range neighbors {
		if n == provinceB {
			return true
		}
	}
	return false
}
