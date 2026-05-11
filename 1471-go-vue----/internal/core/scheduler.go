package core

import (
	"log"
	"time"
)

type Scheduler struct {
	store         *Store
	stopChan      chan struct{}
	ticker        *time.Ticker
}

func NewScheduler(store *Store) *Scheduler {
	return &Scheduler{
		store:    store,
		stopChan: make(chan struct{}),
	}
}

func (sch *Scheduler) Start() {
	sch.ticker = time.NewTicker(1 * time.Minute)
	
	go func() {
		for {
			select {
			case <-sch.ticker.C:
				sch.checkAndRunTasks()
			case <-sch.stopChan:
				sch.ticker.Stop()
				return
			}
		}
	}()
	
	log.Println("定时任务调度器已启动")
}

func (sch *Scheduler) Stop() {
	close(sch.stopChan)
	log.Println("定时任务调度器已停止")
}

func (sch *Scheduler) checkAndRunTasks() {
	now := time.Now()
	
	if now.Hour() == 0 && now.Minute() == 0 {
		sch.runDailyTasks()
	}
}

func (sch *Scheduler) runDailyTasks() {
	log.Println("执行每日定时任务...")
	sch.archiveYesterdaySchedules()
	sch.generateTodaySchedules()
	sch.checkNotBoarded()
}

func (sch *Scheduler) archiveYesterdaySchedules() {
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")
	
	sch.store.mu.Lock()
	defer sch.store.mu.Unlock()
	
	count := 0
	for _, schedule := range sch.store.schedules {
		if schedule.Date == yesterdayStr && schedule.Status == ScheduleStatusArrived {
			schedule.Archived = true
			schedule.UpdatedAt = time.Now()
			count++
		}
	}
	
	log.Printf("已归档 %d 个昨天的已到达班次", count)
}

func (sch *Scheduler) generateTodaySchedules() {
	today := time.Now().Format("2006-01-02")
	
	sch.store.mu.Lock()
	defer sch.store.mu.Unlock()
	
	count := 0
	for _, tpl := range sch.store.scheduleTemplates {
		key := sch.store.GetScheduleKey(tpl.ScheduleNo, today)
		if _, exists := sch.store.schedules[key]; !exists {
			schedule := &Schedule{
				ScheduleNo:       tpl.ScheduleNo,
				Date:             today,
				DepartureStation: tpl.DepartureStation,
				ArrivalStation:   tpl.ArrivalStation,
				DepartureTime:    tpl.DepartureTime,
				ArrivalTime:      tpl.ArrivalTime,
				BusType:          tpl.BusType,
				Price:            tpl.Price,
				TotalSeats:       tpl.TotalSeats,
				SoldSeats:        0,
				Status:           ScheduleStatusNotDeparted,
				IsAlmostSoldOut:  false,
				Archived:         false,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			
			sch.store.schedules[key] = schedule
			
			seats := make([]*Seat, tpl.TotalSeats)
			for i := range seats {
				seats[i] = &Seat{
					SeatNo: i + 1,
					IsSold: false,
				}
			}
			sch.store.scheduleSeats[key] = seats
			
			count++
		}
	}
	
	log.Printf("已生成 %d 个今天的班次", count)
}

func (sch *Scheduler) checkNotBoarded() {
	now := time.Now()
	today := now.Format("2006-01-02")
	
	sch.store.mu.Lock()
	defer sch.store.mu.Unlock()
	
	markedCount := 0
	for _, schedule := range sch.store.schedules {
		if schedule.Date != today || schedule.Status != ScheduleStatusNotDeparted {
			continue
		}
		
		departureDateTime, err := getDepartureDateTime(schedule.Date, schedule.DepartureTime)
		if err != nil {
			continue
		}
		
		if now.After(departureDateTime.Add(15 * time.Minute)) {
			for _, ticket := range sch.store.tickets {
				if ticket.ScheduleNo == schedule.ScheduleNo &&
					ticket.Date == schedule.Date &&
					ticket.Status == TicketStatusSold {
					ticket.Status = TicketStatusNotBoarded
					
					refundReq := &RefundRequest{
						ID:           sch.store.generateRefundID(),
						TicketNo:     ticket.TicketNo,
						Status:       RefundStatusPending,
						RefundAmount: 0,
						RequestTime:  time.Now(),
					}
					sch.store.refundRequests[refundReq.ID] = refundReq
					markedCount++
				}
			}
			
			schedule.Status = ScheduleStatusDeparted
			schedule.UpdatedAt = now
		}
	}
	
	if markedCount > 0 {
		log.Printf("已标记 %d 张未乘车的车票", markedCount)
	}
}
