package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"housekeeping/internal/housekeeping"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9008"
	}

	flagPort := flag.String("port", "", "服务端监听端口")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	return port
}

func main() {
	platform := housekeeping.NewPlatform()
	port := getPort()

	http.HandleFunc("/api/aunt/register", handleRegisterAunt(platform))
	http.HandleFunc("/api/aunt", handleGetAunt(platform))
	http.HandleFunc("/api/customer/register", handleRegisterCustomer(platform))
	http.HandleFunc("/api/customer", handleGetCustomer(platform))
	http.HandleFunc("/api/booking/create", handleCreateBooking(platform))
	http.HandleFunc("/api/booking", handleGetBooking(platform))
	http.HandleFunc("/api/service/start", handleStartService(platform))
	http.HandleFunc("/api/service/complete", handleCompleteService(platform))
	http.HandleFunc("/api/review/add", handleAddReview(platform))
	http.HandleFunc("/api/stats", handleGetStats(platform))

	fmt.Printf("家政服务预约平台服务端启动，监听端口: %s\n", port)
	fmt.Printf("可用接口:\n")
	fmt.Printf("  POST /api/aunt/register   - 阿姨入驻\n")
	fmt.Printf("  GET  /api/aunt            - 查询阿姨列表/详情\n")
	fmt.Printf("  POST /api/customer/register - 客户注册\n")
	fmt.Printf("  GET  /api/customer        - 查询客户列表/详情\n")
	fmt.Printf("  POST /api/booking/create  - 创建预约订单\n")
	fmt.Printf("  GET  /api/booking         - 查询订单列表/详情\n")
	fmt.Printf("  POST /api/service/start   - 开始服务\n")
	fmt.Printf("  POST /api/service/complete- 完成服务\n")
	fmt.Printf("  POST /api/review/add      - 添加评价\n")
	fmt.Printf("  GET  /api/stats           - 获取月度统计\n")

	addr := ":" + port
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
