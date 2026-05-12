from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
import logging

from database import SessionLocal
from routers.lost_found import run_daily_matching, mark_stale_items
from routers.complaints import check_overdue_complaints, auto_close_pending_confirmation
from config import settings

logging.basicConfig()
logging.getLogger("apscheduler").setLevel(logging.WARNING)

scheduler = BackgroundScheduler()


def daily_task():
    db = SessionLocal()
    try:
        match_count = run_daily_matching(db)
        stale_count = mark_stale_items(db)
        print(f"[定时任务] 凌晨自动匹配完成: 创建 {match_count} 条匹配, 标记 {stale_count} 件过期物品")
    except Exception as e:
        print(f"[定时任务] 凌晨任务执行失败: {e}")
    finally:
        db.close()


def check_overdue_task():
    db = SessionLocal()
    try:
        result = check_overdue_complaints(db)
        if result["overdue_count"] > 0:
            print(f"[定时任务] 投诉超时检查: 标记 {result['overdue_count']} 条逾期, 通知主管 {result['supervisor_notifications']} 次")
    except Exception as e:
        print(f"[定时任务] 超时检查失败: {e}")
    finally:
        db.close()


def auto_close_task():
    db = SessionLocal()
    try:
        closed_count = auto_close_pending_confirmation(db)
        if closed_count > 0:
            print(f"[定时任务] 自动关闭 {closed_count} 条7天未确认的投诉")
    except Exception as e:
        print(f"[定时任务] 自动关闭失败: {e}")
    finally:
        db.close()


def start_scheduler():
    if not settings.SCHEDULER_START:
        return
    
    scheduler.add_job(
        daily_task,
        trigger=CronTrigger(hour=0, minute=0),
        id="daily_matching",
        replace_existing=True
    )
    
    scheduler.add_job(
        check_overdue_task,
        trigger=CronTrigger(minute="*/30"),
        id="check_overdue",
        replace_existing=True
    )
    
    scheduler.add_job(
        auto_close_task,
        trigger=CronTrigger(hour="*/6"),
        id="auto_close",
        replace_existing=True
    )
    
    scheduler.start()
    print("[定时任务] 调度器已启动: 凌晨匹配(00:00), 超时检查(每30分钟), 自动关闭(每6小时)")


def stop_scheduler():
    if scheduler.running:
        scheduler.shutdown(wait=False)
        print("[定时任务] 调度器已停止")
