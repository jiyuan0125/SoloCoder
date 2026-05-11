package main

import (
	"bus-station/internal/core"
	"bus-station/pkg/server"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := flag.String("port", "", "服务端口 (也可通过环境变量 PORT 设置)")
	flag.Parse()

	listenPort := "8900"
	if *port != "" {
		listenPort = *port
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		listenPort = envPort
	}

	store := core.NewStore()

	initSampleData(store)

	scheduler := core.NewScheduler(store)
	scheduler.Start()

	srv := server.NewServer(store, scheduler)

	addr := ":" + listenPort

	go func() {
		if err := srv.Start(addr); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("正在停止服务...")
	scheduler.Stop()
	log.Println("服务已停止")
}

func initSampleData(store *core.Store) {
	templates := []*core.ScheduleTemplate{
		{
			ScheduleNo:      "K001",
			DepartureStation: "北京",
			ArrivalStation:   "天津",
			DepartureTime:    "07:30",
			ArrivalTime:      "09:30",
			BusType:          "豪华大巴",
			Price:            80,
			TotalSeats:       45,
		},
		{
			ScheduleNo:      "K002",
			DepartureStation: "北京",
			ArrivalStation:   "天津",
			DepartureTime:    "10:00",
			ArrivalTime:      "12:00",
			BusType:          "普通大巴",
			Price:            60,
			TotalSeats:       45,
		},
		{
			ScheduleNo:      "K003",
			DepartureStation: "北京",
			ArrivalStation:   "上海",
			DepartureTime:    "08:00",
			ArrivalTime:      "18:00",
			BusType:          "卧铺大巴",
			Price:            280,
			TotalSeats:       36,
		},
		{
			ScheduleNo:      "K004",
			DepartureStation: "北京",
			ArrivalStation:   "石家庄",
			DepartureTime:    "09:00",
			ArrivalTime:      "12:00",
			BusType:          "豪华大巴",
			Price:            120,
			TotalSeats:       45,
		},
	}

	for _, tpl := range templates {
		if err := store.AddScheduleTemplate(tpl); err != nil {
			log.Printf("添加班次模板失败: %v", err)
		}
	}

	today := time.Now().Format("2006-01-02")
	for _, tpl := range templates {
		if err := store.CreateScheduleFromTemplate(tpl, today); err != nil && err != core.ErrScheduleAlreadyExist {
			log.Printf("创建今日班次失败: %v", err)
		}
	}

	log.Println("示例数据初始化完成")
}
