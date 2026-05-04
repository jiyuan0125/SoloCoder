package main

import (
	"flag"
	"fmt"
	"hospital-bed/internal/server/handler"
	"hospital-bed/internal/server/service"
	"hospital-bed/internal/server/store"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := flag.Int("port", 8080, "HTTP 服务端口")
	dataPath := flag.String("data", "./data/hospital-bed.json", "数据持久化文件路径")
	flag.Parse()

	st := store.NewStore(*dataPath)
	svc := service.NewBedService(st)
	h := handler.NewAPIHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		fmt.Printf("医院床位管理服务端启动，监听端口: %d\n", *port)
		fmt.Printf("数据文件: %s\n", *dataPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	fmt.Println("\n正在关闭服务...")
	if err := st.SaveAll(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	} else {
		fmt.Println("数据已保存")
	}
}
