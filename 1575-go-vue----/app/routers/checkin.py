from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime, timedelta
from app.database import get_db, Station, Train, Alert
from app.schemas import TrainCreate, TrainResponse

router = APIRouter()

CHECKIN_START_MINUTES_BEFORE = 15
CHECKIN_END_MINUTES_BEFORE = 3
MIN_PLATFORM_INTERVAL_MINUTES = 10

@router.post("/trains", response_model=TrainResponse, status_code=status.HTTP_201_CREATED)
def create_train(train: TrainCreate, db: Session = Depends(get_db)):
    existing = db.query(Train).filter(Train.train_number == train.train_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="车次已存在")
    
    checkin_start = train.departure_time - timedelta(minutes=CHECKIN_START_MINUTES_BEFORE)
    checkin_end = train.departure_time - timedelta(minutes=CHECKIN_END_MINUTES_BEFORE)
    
    db_train = Train(
        station_id=1,
        train_number=train.train_number,
        platform=train.platform,
        departure_time=train.departure_time,
        checkin_start_time=checkin_start,
        checkin_end_time=checkin_end
    )
    db.add(db_train)
    db.commit()
    
    check_platform_conflicts(db, train.platform, train.departure_time, db_train.id)
    
    db.refresh(db_train)
    return db_train

@router.post("/trains/{train_number}/checkin/start", response_model=TrainResponse)
def start_checkin(train_number: str, db: Session = Depends(get_db)):
    train = db.query(Train).filter(Train.train_number == train_number).first()
    if not train:
        raise HTTPException(status_code=404, detail="车次不存在")
    
    now = datetime.utcnow()
    if now < train.checkin_start_time:
        raise HTTPException(
            status_code=400, 
            detail=f"检票时间未到，检票将在 {train.checkin_start_time} 开始"
        )
    if now > train.checkin_end_time:
        raise HTTPException(
            status_code=400, 
            detail=f"检票已结束，检票结束时间为 {train.checkin_end_time}"
        )
    
    train.is_checkin_active = True
    db.commit()
    db.refresh(train)
    return train

@router.post("/trains/{train_number}/checkin/stop", response_model=TrainResponse)
def stop_checkin(train_number: str, db: Session = Depends(get_db)):
    train = db.query(Train).filter(Train.train_number == train_number).first()
    if not train:
        raise HTTPException(status_code=404, detail="车次不存在")
    
    train.is_checkin_active = False
    db.commit()
    db.refresh(train)
    return train

@router.get("/trains/{train_number}", response_model=TrainResponse)
def get_train(train_number: str, db: Session = Depends(get_db)):
    train = db.query(Train).filter(Train.train_number == train_number).first()
    if not train:
        raise HTTPException(status_code=404, detail="车次不存在")
    return train

@router.get("/trains", response_model=List[TrainResponse])
def list_trains(platform: str = None, db: Session = Depends(get_db)):
    query = db.query(Train)
    if platform:
        query = query.filter(Train.platform == platform)
    return query.order_by(Train.departure_time).all()

@router.delete("/trains/{train_number}", status_code=status.HTTP_204_NO_CONTENT)
def delete_train(train_number: str, db: Session = Depends(get_db)):
    train = db.query(Train).filter(Train.train_number == train_number).first()
    if not train:
        raise HTTPException(status_code=404, detail="车次不存在")
    
    db.delete(train)
    db.commit()

def check_platform_conflicts(db: Session, platform: str, departure_time: datetime, current_train_id: int):
    min_interval = timedelta(minutes=MIN_PLATFORM_INTERVAL_MINUTES)
    
    trains_on_platform = db.query(Train).filter(
        Train.platform == platform,
        Train.id != current_train_id
    ).all()
    
    for train in trains_on_platform:
        time_diff = abs((train.departure_time - departure_time).total_seconds() / 60)
        if time_diff < MIN_PLATFORM_INTERVAL_MINUTES:
            alert = Alert(
                station_id=train.station_id,
                alert_type="platform_conflict",
                severity="warning",
                message=f"站台 {platform} 存在客流冲突：车次 {train.train_number} 与新车次发车时间间隔仅 {int(time_diff)} 分钟，小于 {MIN_PLATFORM_INTERVAL_MINUTES} 分钟",
                is_active=True
            )
            db.add(alert)
    
    db.commit()
