package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"return-exchange/database"
	"return-exchange/handler"
)

func main() {
	fmt.Println("Starting return-exchange service...")
	
	// 初始化数据库
	fmt.Println("正在初始化数据库...")
	if err := database.InitDB(); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer database.CloseDB()
	fmt.Println("数据库初始化成功")

	// 创建处理器
	h := handler.NewHandler()

	// 设置路由
	mux := http.NewServeMux()

	// 用户端接口
	mux.HandleFunc("/api/return", h.CreateReturnRequestHandler)         // POST 创建退货申请
	mux.HandleFunc("/api/exchange", h.CreateExchangeRequestHandler)     // POST 创建换货申请
	mux.HandleFunc("/api/request", h.GetRequestByIDHandler)             // GET 获取申请详情 ?id=1
	mux.HandleFunc("/api/requests", h.GetRequestsByUserIDHandler)       // GET 获取用户申请列表 ?user_id=1
	mux.HandleFunc("/api/refund", h.GetRefundByRequestIDHandler)        // GET 获取退款记录 ?request_id=1
	mux.HandleFunc("/api/shipment", h.GetShipmentByRequestIDHandler)    // GET 获取发货单 ?request_id=1

	// 管理员端接口
	mux.HandleFunc("/api/admin/approve", h.ApproveRequestHandler)       // POST 审核通过 ?id=1
	mux.HandleFunc("/api/admin/reject", h.RejectRequestHandler)         // POST 审核拒绝 ?id=1

	// 测试接口
	mux.HandleFunc("/api/order", h.GetOrderByNoHandler)                 // GET 获取订单信息 ?order_no=xxx

	// 启动服务器
	port := ":8080"
	fmt.Printf("服务器启动在端口 %s...\n", port)
	fmt.Println("可用接口:")
	fmt.Println("  POST /api/return          - 创建退货申请")
	fmt.Println("  POST /api/exchange        - 创建换货申请")
	fmt.Println("  GET  /api/request         - 获取申请详情 (?id=1)")
	fmt.Println("  GET  /api/requests        - 获取用户申请列表 (?user_id=1)")
	fmt.Println("  GET  /api/refund          - 获取退款记录 (?request_id=1)")
	fmt.Println("  GET  /api/shipment        - 获取发货单 (?request_id=1)")
	fmt.Println("  POST /api/admin/approve   - 审核通过 (?id=1)")
	fmt.Println("  POST /api/admin/reject    - 审核拒绝 (?id=1)")
	fmt.Println("  GET  /api/order           - 获取订单信息 (?order_no=ORD20260501001)")

	// 优雅关闭
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// 处理信号
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("正在关闭服务器...")
		database.CloseDB()
		os.Exit(0)
	}()

	fmt.Println("服务器准备就绪，正在监听端口 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
