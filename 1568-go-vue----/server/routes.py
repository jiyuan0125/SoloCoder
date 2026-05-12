from datetime import datetime, date, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import and_

from .database import get_db
from .models import (
    Officer, Channel, Schedule, CheckIn, Contraband, TrafficLog,
    ChannelType, ChannelStatus
)
from .schemas import (
    OfficerCreate, OfficerResponse,
    ChannelCreate, ChannelResponse, ChannelDetail, ChannelStatusUpdate,
    ScheduleCreate, ScheduleResponse,
    CheckInCreate, CheckInResponse,
    ContrabandCreate, ContrabandResponse,
    TrafficLogCreate, TrafficLogResponse, TrafficSummary,
    ChannelStatusInfo, ChannelCapacityAlert
)
from .services import (
    ScheduleService, ChannelService, ContrabandService,
    CheckInService, TrafficService, get_current_shift
)

router = APIRouter()


@router.post("/officers", response_model=OfficerResponse, status_code=201)
def create_officer(officer: OfficerCreate, db: Session = Depends(get_db)):
    existing = db.query(Officer).filter(Officer.badge_number == officer.badge_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="警号已存在")
    
    db_officer = Officer(**officer.model_dump())
    db.add(db_officer)
    db.commit()
    db.refresh(db_officer)
    return db_officer


@router.get("/officers", response_model=List[OfficerResponse])
def list_officers(db: Session = Depends(get_db)):
    return db.query(Officer).all()


@router.get("/officers/{officer_id}", response_model=OfficerResponse)
def get_officer(officer_id: int, db: Session = Depends(get_db)):
    officer = db.query(Officer).filter(Officer.id == officer_id).first()
    if not officer:
        raise HTTPException(status_code=404, detail="安检员不存在")
    return officer


@router.post("/channels", response_model=ChannelResponse, status_code=201)
def create_channel(channel: ChannelCreate, db: Session = Depends(get_db)):
    db_channel = Channel(**channel.model_dump())
    db.add(db_channel)
    db.commit()
    db.refresh(db_channel)
    return db_channel


@router.get("/channels", response_model=List[ChannelResponse])
def list_channels(db: Session = Depends(get_db)):
    return db.query(Channel).all()


@router.get("/channels/{channel_id}", response_model=ChannelDetail)
def get_channel(channel_id: int, db: Session = Depends(get_db)):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    base_url = f"/channels/{channel_id}"
    return ChannelDetail(
        id=channel.id,
        name=channel.name,
        channel_type=channel.channel_type,
        capacity_per_hour=channel.capacity_per_hour,
        status=channel.status,
        created_at=channel.created_at,
        schedules_href=f"{base_url}/schedules",
        contrabands_href=f"{base_url}/contrabands",
        traffic_href=f"{base_url}/traffic"
    )


@router.get("/channels/{channel_id}/status-info", response_model=ChannelStatusInfo)
def get_channel_status_info(
    channel_id: int,
    schedule_date: Optional[date] = None,
    shift: Optional[str] = None,
    db: Session = Depends(get_db)
):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    if schedule_date is None:
        schedule_date = datetime.now().date()
    
    if shift is None:
        from .models import ShiftType
        shift_enum = get_current_shift(datetime.now())
    else:
        from .models import ShiftType
        shift_enum = ShiftType(shift)
    
    return ChannelService.check_channel_staffing(db, channel_id, schedule_date, shift_enum)


@router.get("/channels/{channel_id}/capacity-alert", response_model=Optional[ChannelCapacityAlert])
def get_capacity_alert(channel_id: int, db: Session = Depends(get_db)):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    return ChannelService.check_capacity_alert(db, channel_id)


@router.get("/channels/{channel_id}/schedules", response_model=List[ScheduleResponse])
def get_channel_schedules(
    channel_id: int,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    query = db.query(Schedule).filter(Schedule.channel_id == channel_id)
    
    if start_date:
        query = query.filter(Schedule.schedule_date >= start_date)
    if end_date:
        query = query.filter(Schedule.schedule_date <= end_date)
    
    return query.order_by(Schedule.schedule_date.asc(), Schedule.shift.asc()).all()


@router.get("/channels/{channel_id}/contrabands", response_model=List[ContrabandResponse])
def get_channel_contrabands(
    channel_id: int,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    query = db.query(Contraband).filter(Contraband.channel_id == channel_id)
    
    if start_date:
        query = query.filter(Contraband.recorded_at >= datetime.combine(start_date, datetime.min.time()))
    if end_date:
        query = query.filter(Contraband.recorded_at <= datetime.combine(end_date, datetime.max.time()))
    
    return query.order_by(Contraband.recorded_at.desc()).all()


@router.get("/channels/{channel_id}/traffic", response_model=List[TrafficLogResponse])
def get_channel_traffic(
    channel_id: int,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    query = db.query(TrafficLog).filter(TrafficLog.channel_id == channel_id)
    
    if start_date:
        query = query.filter(TrafficLog.log_date >= start_date)
    if end_date:
        query = query.filter(TrafficLog.log_date <= end_date)
    
    return query.order_by(TrafficLog.log_date.desc(), TrafficLog.hour.desc()).all()


@router.get("/channels/{channel_id}/traffic/summary", response_model=TrafficSummary)
def get_channel_traffic_summary(
    channel_id: int,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    channel = db.query(Channel).filter(Channel.id == channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    if end_date is None:
        end_date = datetime.now().date()
    if start_date is None:
        start_date = end_date - timedelta(days=7)
    
    return TrafficService.get_channel_traffic_summary(db, channel_id, start_date, end_date)


@router.post("/schedules", response_model=ScheduleResponse, status_code=201)
def create_schedule(schedule: ScheduleCreate, db: Session = Depends(get_db)):
    officer = db.query(Officer).filter(Officer.id == schedule.officer_id).first()
    if not officer:
        raise HTTPException(status_code=404, detail="安检员不存在")
    
    channel = db.query(Channel).filter(Channel.id == schedule.channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    errors = ScheduleService.validate_schedule(
        db, schedule.officer_id, schedule.channel_id,
        schedule.schedule_date, schedule.shift
    )
    if errors:
        raise HTTPException(status_code=400, detail="; ".join(errors))
    
    db_schedule = Schedule(**schedule.model_dump())
    db.add(db_schedule)
    db.commit()
    db.refresh(db_schedule)
    
    ChannelService.update_channel_status(db, schedule.channel_id)
    
    return db_schedule


@router.get("/schedules", response_model=List[ScheduleResponse])
def list_schedules(
    officer_id: Optional[int] = None,
    channel_id: Optional[int] = None,
    schedule_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Schedule)
    
    if officer_id:
        query = query.filter(Schedule.officer_id == officer_id)
    if channel_id:
        query = query.filter(Schedule.channel_id == channel_id)
    if schedule_date:
        query = query.filter(Schedule.schedule_date == schedule_date)
    
    return query.order_by(Schedule.schedule_date.asc(), Schedule.shift.asc()).all()


@router.get("/schedules/{schedule_id}", response_model=ScheduleResponse)
def get_schedule(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(Schedule).filter(Schedule.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    return schedule


@router.post("/check-ins", response_model=CheckInResponse, status_code=201)
def create_check_in(check_in: CheckInCreate, db: Session = Depends(get_db)):
    officer = db.query(Officer).filter(Officer.id == check_in.officer_id).first()
    if not officer:
        raise HTTPException(status_code=404, detail="安检员不存在")
    
    return CheckInService.create_check_in(db, check_in.officer_id, check_in.check_in_time)


@router.get("/check-ins", response_model=List[CheckInResponse])
def list_check_ins(
    officer_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(CheckIn)
    
    if officer_id:
        query = query.filter(CheckIn.officer_id == officer_id)
    if start_date:
        query = query.filter(CheckIn.check_in_time >= datetime.combine(start_date, datetime.min.time()))
    if end_date:
        query = query.filter(CheckIn.check_in_time <= datetime.combine(end_date, datetime.max.time()))
    
    return query.order_by(CheckIn.check_in_time.desc()).all()


@router.get("/check-ins/{check_in_id}", response_model=CheckInResponse)
def get_check_in(check_in_id: int, db: Session = Depends(get_db)):
    check_in = db.query(CheckIn).filter(CheckIn.id == check_in_id).first()
    if not check_in:
        raise HTTPException(status_code=404, detail="打卡记录不存在")
    return check_in


@router.post("/contrabands", response_model=ContrabandResponse, status_code=201)
def create_contraband(contraband: ContrabandCreate, db: Session = Depends(get_db)):
    channel = db.query(Channel).filter(Channel.id == contraband.channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    errors = ContrabandService.validate_contraband(
        contraband.item_type, contraband.disposal_type, contraband.police_badge
    )
    if errors:
        raise HTTPException(status_code=400, detail="; ".join(errors))
    
    if contraband.recorded_at is None:
        contraband.recorded_at = datetime.now()
    
    db_contraband = Contraband(**contraband.model_dump())
    db.add(db_contraband)
    db.commit()
    db.refresh(db_contraband)
    return db_contraband


@router.get("/contrabands", response_model=List[ContrabandResponse])
def list_contrabands(
    channel_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Contraband)
    
    if channel_id:
        query = query.filter(Contraband.channel_id == channel_id)
    if start_date:
        query = query.filter(Contraband.recorded_at >= datetime.combine(start_date, datetime.min.time()))
    if end_date:
        query = query.filter(Contraband.recorded_at <= datetime.combine(end_date, datetime.max.time()))
    
    return query.order_by(Contraband.recorded_at.desc()).all()


@router.post("/traffic-logs", response_model=TrafficLogResponse, status_code=201)
def create_traffic_log(log: TrafficLogCreate, db: Session = Depends(get_db)):
    channel = db.query(Channel).filter(Channel.id == log.channel_id).first()
    if not channel:
        raise HTTPException(status_code=404, detail="通道不存在")
    
    db_log = TrafficLog(**log.model_dump())
    db.add(db_log)
    db.commit()
    db.refresh(db_log)
    return db_log


@router.get("/traffic-logs", response_model=List[TrafficLogResponse])
def list_traffic_logs(
    channel_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(TrafficLog)
    
    if channel_id:
        query = query.filter(TrafficLog.channel_id == channel_id)
    if start_date:
        query = query.filter(TrafficLog.log_date >= start_date)
    if end_date:
        query = query.filter(TrafficLog.log_date <= end_date)
    
    return query.order_by(TrafficLog.log_date.desc(), TrafficLog.hour.desc()).all()
