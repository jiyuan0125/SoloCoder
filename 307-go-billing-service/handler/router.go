package handler

import (
	"billing-service/service"
	"net/http"
)

type Router struct {
	billHandler     *BillHandler
	customerHandler *CustomerHandler
	usageHandler    *UsageHandler
	configHandler   *ConfigHandler
}

func NewRouter(
	billService *service.BillService,
	customerService *service.CustomerService,
	usageService *service.UsageService,
	configService *service.ConfigService,
) *Router {
	return &Router{
		billHandler:     NewBillHandler(billService),
		customerHandler: NewCustomerHandler(customerService),
		usageHandler:    NewUsageHandler(usageService),
		configHandler:   NewConfigHandler(configService),
	}
}

func (r *Router) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/customers", r.customerHandler.HandleCustomers)
	mux.HandleFunc("/api/customers/", r.customerHandler.HandleCustomerByID)

	mux.HandleFunc("/api/bills", r.billHandler.HandleBills)
	mux.HandleFunc("/api/bills/", r.billHandler.HandleBillByID)

	mux.HandleFunc("/api/usages", r.usageHandler.HandleUsages)

	mux.HandleFunc("/api/configs", r.configHandler.HandleConfigs)
	mux.HandleFunc("/api/configs/", r.configHandler.HandleConfigByKey)
}
