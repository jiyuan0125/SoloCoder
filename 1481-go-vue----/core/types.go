package core

import (
	"fmt"
	"usedcar/api"
)

type VehicleInternal struct {
	*api.Vehicle
	LastEvaluation *api.Evaluation
}

type DepositInternal struct {
	*api.Deposit
	ReportID string
}

type ReferencePriceDB struct {
	prices map[string]int64
}

func NewReferencePriceDB() *ReferencePriceDB {
	return &ReferencePriceDB{
		prices: defaultReferencePrices(),
	}
}

func defaultReferencePrices() map[string]int64 {
	return map[string]int64{
		"大众-帕萨特-2020": 15000000,
		"大众-迈腾-2020":   16000000,
		"丰田-凯美瑞-2020": 18000000,
		"丰田-卡罗拉-2020": 10000000,
		"本田-雅阁-2020":   17000000,
		"本田-思域-2020":   12000000,
		"宝马-3系-2020":   28000000,
		"奥迪-A4L-2020":  25000000,
		"奔驰-C级-2020":  30000000,
		"大众-帕萨特-2019": 12000000,
		"大众-迈腾-2019":   13000000,
		"丰田-凯美瑞-2019": 15000000,
		"丰田-卡罗拉-2019": 8000000,
		"本田-雅阁-2019":   14000000,
		"本田-思域-2019":   10000000,
		"宝马-3系-2019":   24000000,
		"奥迪-A4L-2019":  22000000,
		"奔驰-C级-2019":  26000000,
		"大众-帕萨特-2021": 18000000,
		"大众-迈腾-2021":   19000000,
		"丰田-凯美瑞-2021": 21000000,
		"丰田-卡罗拉-2021": 12000000,
		"本田-雅阁-2021":   20000000,
		"本田-思域-2021":   14000000,
	}
}

func (r *ReferencePriceDB) GetReferencePrice(brand, model string, year int) (int64, bool) {
	key := fmt.Sprintf("%s-%s-%d", brand, model, year)
	price, ok := r.prices[key]
	if ok {
		return price, true
	}
	return 10000000, false
}

func (r *ReferencePriceDB) SetReferencePrice(brand, model string, year int, priceFen int64) {
	key := fmt.Sprintf("%s-%s-%d", brand, model, year)
	r.prices[key] = priceFen
}
