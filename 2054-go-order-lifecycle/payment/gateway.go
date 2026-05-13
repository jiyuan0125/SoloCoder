package payment

import (
	"fmt"
	"time"

	"order-lifecycle/models"
)

type PaymentGateway interface {
	ProcessPayment(req models.PaymentRequest) models.PaymentResult
	RefundPayment(orderID string, amount float64) error
}

type BalanceGateway struct{}

func (g *BalanceGateway) ProcessPayment(req models.PaymentRequest) models.PaymentResult {
	time.Sleep(10 * time.Millisecond)
	return models.PaymentResult{Success: true, Message: "余额支付成功"}
}

func (g *BalanceGateway) RefundPayment(orderID string, amount float64) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

type BankCardGateway struct{}

func (g *BankCardGateway) ProcessPayment(req models.PaymentRequest) models.PaymentResult {
	time.Sleep(10 * time.Millisecond)
	return models.PaymentResult{Success: true, Message: "银行卡支付成功"}
}

func (g *BankCardGateway) RefundPayment(orderID string, amount float64) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

type ThirdPartyGateway struct{}

func (g *ThirdPartyGateway) ProcessPayment(req models.PaymentRequest) models.PaymentResult {
	time.Sleep(10 * time.Millisecond)
	return models.PaymentResult{Success: true, Message: "第三方支付成功"}
}

func (g *ThirdPartyGateway) RefundPayment(orderID string, amount float64) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

type GatewayFactory struct{}

func (f *GatewayFactory) GetGateway(method models.PaymentMethod) (PaymentGateway, error) {
	switch method {
	case models.PaymentMethodBalance:
		return &BalanceGateway{}, nil
	case models.PaymentMethodBankCard:
		return &BankCardGateway{}, nil
	case models.PaymentMethodThirdParty:
		return &ThirdPartyGateway{}, nil
	default:
		return nil, fmt.Errorf("不支持的支付方式: %s", method)
	}
}
