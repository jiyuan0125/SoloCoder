package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"order-lifecycle/handler"
	"order-lifecycle/inventory"
	"order-lifecycle/payment"
	"order-lifecycle/repository"
	"order-lifecycle/service"
	"order-lifecycle/statemachine"
)

func main() {
	store := repository.NewStore()
	repo := repository.NewOrderRepository(store)
	inventorySvc := inventory.NewInventoryService(store)
	stateMachine := statemachine.NewStateMachine()
	paymentFactory := &payment.GatewayFactory{}
	orderService := service.NewOrderService(repo, inventorySvc, stateMachine, paymentFactory)
	h := handler.NewHandler(orderService)

	if err := orderService.AddInventory("P001", 100); err != nil {
		log.Fatalf("初始化库存失败: %v", err)
	}
	fmt.Println("[初始化] 添加测试库存: 商品 P001, 数量 100")

	go startOrderExpiryJob(orderService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", h.CreateOrder)
	mux.HandleFunc("/api/orders/", h.GetOrder)
	mux.HandleFunc("/api/pay", h.Pay)
	mux.HandleFunc("/api/ship", h.Ship)
	mux.HandleFunc("/api/deliver", h.Deliver)
	mux.HandleFunc("/api/complete", h.Complete)
	mux.HandleFunc("/api/refund", h.Refund)
	mux.HandleFunc("/api/inventory/add", h.AddInventory)

	fmt.Println("订单生命周期系统启动中，端口: 9104")
	fmt.Println("API接口:")
	fmt.Println("  POST /api/orders          - 创建订单")
	fmt.Println("  GET  /api/orders/:id      - 查询订单")
	fmt.Println("  POST /api/pay             - 支付订单")
	fmt.Println("  POST /api/ship            - 发货")
	fmt.Println("  POST /api/deliver         - 签收")
	fmt.Println("  POST /api/complete        - 完成订单")
	fmt.Println("  POST /api/refund          - 退货退款")
	fmt.Println("  POST /api/inventory/add   - 添加库存")

	server := &http.Server{
		Addr:         ":9104",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func startOrderExpiryJob(orderService *service.OrderService) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	fmt.Println("[定时任务] 订单超时检查已启动（每1分钟执行一次）")
	for range ticker.C {
		if err := orderService.CancelExpiredOrders(); err != nil {
			fmt.Printf("[定时任务] 订单超时检查失败: %v\n", err)
		}
	}
}
