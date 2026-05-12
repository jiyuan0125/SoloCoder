package scheduler

import (
	"time"

	"ambulance-scheduler/internal/logger"
	"ambulance-scheduler/internal/service"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron            *cron.Cron
	dispatchService *service.DispatchService
	callService     *service.CallService
	logger          *logger.Logger
}

func New(dispatchService *service.DispatchService, callService *service.CallService, log *logger.Logger) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		dispatchService: dispatchService,
		callService:     callService,
		logger:          log,
	}
}

func (s *Scheduler) Start() {
	_, _ = s.cron.AddFunc("0 * * * * *", s.checkTimeoutVehicles)
	
	_, _ = s.cron.AddFunc("0 */10 * * * *", s.upgradeWaitingCalls)
	
	_, _ = s.cron.AddFunc("0 0 * * * *", s.closeExpiredCalls)

	s.cron.Start()
	s.logger.Info("定时任务调度器已启动")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("定时任务调度器已停止")
}

func (s *Scheduler) checkTimeoutVehicles() {
	timeoutVehicles := s.dispatchService.CheckTimeoutVehicles()
	if len(timeoutVehicles) > 0 {
		s.logger.Error("检测到 %d 辆超时车辆: %v", len(timeoutVehicles), timeoutVehicles)
	}
}

func (s *Scheduler) upgradeWaitingCalls() {
	s.dispatchService.UpgradeWaitingCalls()
	s.logger.Info("已执行等待队列升级检查: %s", time.Now().Format("2006-01-02 15:04:05"))
}

func (s *Scheduler) closeExpiredCalls() {
	s.callService.CloseExpiredCalls()
	s.logger.Info("已执行过期求救关闭检查: %s", time.Now().Format("2006-01-02 15:04:05"))
}
