from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from datetime import datetime
from app.database import get_db
from app.schemas import QueueRealtimeResponse, QueueEstimateResponse
from app.services.queue_manager import QueueManager
from app.services.scheduler import Scheduler

router = APIRouter(prefix="/queue", tags=["排队管理"])

scheduler = Scheduler()


@router.get("/realtime", response_model=QueueRealtimeResponse)
def get_realtime_queue(db: Session = Depends(get_db)):
    queue_manager = QueueManager(db)
    stats = queue_manager.get_queue_stats()
    
    return QueueRealtimeResponse(
        total_queue=stats["total_queue"],
        upward_passengers=stats["upward_passengers"],
        effective_queue=stats["effective_queue"],
        queue_groups=stats["queue_groups"],
        average_wait_seconds=stats["average_wait_seconds"],
        first_join_time=stats["first_join_time"]
    )


@router.get("/estimate", response_model=QueueEstimateResponse)
def get_queue_estimate(db: Session = Depends(get_db)):
    queue_manager = QueueManager(db)
    stats = queue_manager.get_queue_stats()
    
    effective_queue = stats["effective_queue"]
    interval_info = scheduler.get_interval_info(effective_queue)
    current_interval = interval_info["smoothed_interval_minutes"]
    
    if effective_queue == 0:
        estimated_wait = 0
    else:
        average_cabin_capacity = 12
        cabins_needed = (effective_queue + average_cabin_capacity - 1) // average_cabin_capacity
        estimated_wait = cabins_needed * current_interval
    
    return QueueEstimateResponse(
        estimated_wait_minutes=round(estimated_wait, 1),
        queue_size=effective_queue,
        current_interval_minutes=current_interval,
        next_dispatch_minutes=round(current_interval / 2, 1)
    )
