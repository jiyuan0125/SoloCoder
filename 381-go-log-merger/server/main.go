package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	port      = flag.Int("port", 9876, "服务监听端口")
	debugMode = flag.Bool("debug", false, "调试模式")
)

func main() {
	flag.Parse()

	fmt.Println("=== 日志合并服务启动 ===")
	fmt.Printf("监听端口: %d\n", *port)
	fmt.Printf("调试模式: %v\n", *debugMode)

	taskManager := NewTaskManager()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	defer listener.Close()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-stopCh:
					return
				default:
					log.Printf("接受连接失败: %v", err)
					continue
				}
			}
			go handleConnection(conn, taskManager)
		}
	}()

	fmt.Println("服务已就绪，等待客户端连接...")
	<-stopCh
	fmt.Println("\n收到停止信号，正在关闭服务...")
	taskManager.StopAll()
	fmt.Println("服务已关闭")
}
