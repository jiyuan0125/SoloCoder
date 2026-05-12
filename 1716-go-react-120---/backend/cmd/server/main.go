package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ambulance-scheduler/internal/config"
	"ambulance-scheduler/internal/handler"
	"ambulance-scheduler/internal/logger"
	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/router"
	"ambulance-scheduler/internal/service"
	"ambulance-scheduler/internal/store"
	"ambulance-scheduler/pkg/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	cfg := config.Load()
	log := logger.New()

	dataStore := store.New()

	callService := service.NewCallService(dataStore)
	vehicleService := service.NewVehicleService(dataStore)
	dispatchService := service.NewDispatchService(dataStore, vehicleService)
	triageService := service.NewTriageService(dataStore)
	statsService := service.NewStatsService(dataStore)

	h := handler.New(callService, vehicleService, dispatchService, triageService, statsService)
	r := router.SetupRouter(h)

	sched := scheduler.New(dispatchService, callService, log)
	sched.Start()
	defer sched.Stop()

	seedData(dataStore, log)

	log.Info("服务器启动中，端口: %s", cfg.Port)

	serverAddr := ":" + cfg.Port
	go func() {
		if err := r.Run(serverAddr); err != nil {
			log.Error("服务器启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	log.Info("服务器已关闭")
}

func seedData(store *store.Store, log *logger.Logger) {
	vehicles := []struct {
		Number  string
		Type    string
		Loc     string
		Doctors int
		Nurses  int
	}{
		{"AMB-001", "普通转运车", "市中心急救中心", 1, 2},
		{"AMB-002", "抢救监护车", "东城区急救站", 2, 2},
		{"AMB-003", "新生儿转运车", "南城区急救站", 1, 3},
		{"AMB-004", "普通转运车", "西城区急救站", 1, 2},
		{"AMB-005", "抢救监护车", "北城区急救站", 2, 2},
		{"AMB-006", "普通转运车", "东郊急救站", 1, 2},
	}

	vehicleService := service.NewVehicleService(store)
	for _, v := range vehicles {
		_, _ = vehicleService.CreateVehicle(&service.CreateVehicleRequest{
			VehicleNumber:   v.Number,
			VehicleType:     models.VehicleType(v.Type),
			CurrentLocation: v.Loc,
			DoctorCount:     v.Doctors,
			NurseCount:      v.Nurses,
		})
	}

	log.Info("已初始化6辆救护车数据")
}
