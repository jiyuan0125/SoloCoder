package main

import (
	"coupon-service/dao"
	"coupon-service/handler"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	dbPath := flag.String("db", "./coupons.db", "SQLite database file path")
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	fmt.Println("=== 优惠券服务启动中 ===")
	fmt.Printf("数据库路径: %s\n", *dbPath)
	fmt.Printf("服务端口: %d\n", *port)

	fmt.Println("正在初始化数据库...")
	if err := dao.InitDB(*dbPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	fmt.Println("数据库初始化成功")

	fmt.Println("正在设置路由...")
	r := handler.SetupRouter()

	go func() {
		addr := fmt.Sprintf(":%d", *port)
		fmt.Printf("服务已启动，监听地址: http://localhost%s\n", addr)
		fmt.Println("=== 服务运行中 ===")
		if err := r.Run(addr); err != nil {
			log.Fatalf("HTTP服务启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务...")
	if err := dao.CloseDB(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}
	fmt.Println("服务已关闭")
}
