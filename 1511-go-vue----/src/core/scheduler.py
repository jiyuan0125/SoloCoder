from datetime import datetime, timedelta
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.interval import IntervalTrigger
from typing import Callable, List

from .storage import storage
from .models import AlertType, HarvestRecord


class TeaScheduler:
    def __init__(self):
        self.scheduler = BackgroundScheduler()
        self._is_running = False
    
    def start(self):
        if not self._is_running:
            self.scheduler.add_job(
                self._check_todo_reminders,
                trigger=IntervalTrigger(hours=1),
                id='todo_reminder_check',
                replace_existing=True
            )
            
            self.scheduler.add_job(
                self._check_urgency_alerts,
                trigger=IntervalTrigger(minutes=30),
                id='urgency_alert_check',
                replace_existing=True
            )
            
            self.scheduler.start()
            self._is_running = True
            print("[Scheduler] 茶叶管理调度器已启动")
    
    def stop(self):
        if self._is_running:
            self.scheduler.shutdown()
            self._is_running = False
            print("[Scheduler] 茶叶管理调度器已停止")
    
    def _check_todo_reminders(self):
        tomorrow = datetime.now().date() + timedelta(days=1)
        plans = storage.get_harvest_plans_for_date(tomorrow)
        
        existing_alerts = storage.list_alerts(alert_type=AlertType.TODO)
        existing_plan_ids = {
            alert.related_entity_id 
            for alert in existing_alerts 
            if alert.related_entity_type == "HarvestPlan"
        }
        
        for plan in plans:
            if plan.id not in existing_plan_ids:
                message = f"明天（{tomorrow}）有采摘计划：地块ID={plan.plot_id}，预计采摘量={plan.expected_quantity}公斤"
                storage.create_alert(
                    alert_type=AlertType.TODO,
                    message=message,
                    related_entity_id=plan.id,
                    related_entity_type="HarvestPlan"
                )
                print(f"[TODO] 生成待办提醒：{message}")
    
    def _check_urgency_alerts(self):
        now = datetime.now()
        threshold_time = now - timedelta(hours=24)
        
        overdue_records: List[HarvestRecord] = storage.get_harvest_records_without_processing(
            since=threshold_time
        )
        
        existing_alerts = storage.list_alerts(alert_type=AlertType.URGENCY)
        existing_record_ids = {
            alert.related_entity_id 
            for alert in existing_alerts 
            if alert.related_entity_type == "HarvestRecord"
        }
        
        for record in overdue_records:
            if record.id not in existing_record_ids:
                time_overdue = (now - record.harvest_time).total_seconds() / 3600
                message = f"紧急告警：采摘记录ID={record.id} 已超过24小时未开始加工（超时{time_overdue:.1f}小时）"
                storage.create_alert(
                    alert_type=AlertType.URGENCY,
                    message=message,
                    related_entity_id=record.id,
                    related_entity_type="HarvestRecord"
                )
                print(f"[URGENCY] 生成紧急告警：{message}")


scheduler = TeaScheduler()
