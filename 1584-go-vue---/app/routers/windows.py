from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import date
from app.database import get_db
from app.models import Line, MaintenanceWindow
from app.schemas import MaintenanceWindowCreate, MaintenanceWindowUpdate, MaintenanceWindowResponse
from app.services import get_maintenance_window

router = APIRouter()


@router.post("/windows", response_model=MaintenanceWindowResponse)
def create_window(window: MaintenanceWindowCreate, db: Session = Depends(get_db)):
    line = db.query(Line).filter(Line.id == window.line_id).first()
    if not line:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    if not window.is_general and not window.date:
        raise HTTPException(status_code=400, detail="非通用配置必须指定日期")
    
    if window.is_general:
        existing_general = db.query(MaintenanceWindow).filter(
            MaintenanceWindow.line_id == window.line_id,
            MaintenanceWindow.is_general == True
        ).first()
        if existing_general:
            raise HTTPException(status_code=400, detail="该线路已存在通用配置")
    else:
        existing_date = db.query(MaintenanceWindow).filter(
            MaintenanceWindow.line_id == window.line_id,
            MaintenanceWindow.date == window.date
        ).first()
        if existing_date:
            raise HTTPException(status_code=400, detail=f"该线路在日期 {window.date} 已存在配置")
    
    db_window = MaintenanceWindow(
        line_id=window.line_id,
        date=window.date,
        start_time=window.start_time,
        end_time=window.end_time,
        is_general=window.is_general
    )
    db.add(db_window)
    db.commit()
    db.refresh(db_window)
    return db_window


@router.get("/lines/{line_id}/windows/{date_str}", response_model=MaintenanceWindowResponse)
def get_window_for_date(line_id: int, date_str: str, db: Session = Depends(get_db)):
    try:
        target_date = date.fromisoformat(date_str)
    except ValueError:
        raise HTTPException(status_code=400, detail="日期格式错误，请使用 YYYY-MM-DD")
    
    line = db.query(Line).filter(Line.id == line_id).first()
    if not line:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    window = get_maintenance_window(db, line_id, target_date)
    if not window:
        raise HTTPException(status_code=404, detail="未找到天窗时间配置")
    
    return window


@router.get("/lines/{line_id}/windows", response_model=List[MaintenanceWindowResponse])
def list_windows(line_id: int, db: Session = Depends(get_db)):
    line = db.query(Line).filter(Line.id == line_id).first()
    if not line:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    return db.query(MaintenanceWindow).filter(
        MaintenanceWindow.line_id == line_id
    ).all()


@router.put("/windows/{window_id}", response_model=MaintenanceWindowResponse)
def update_window(
    window_id: int,
    window_update: MaintenanceWindowUpdate,
    db: Session = Depends(get_db)
):
    window = db.query(MaintenanceWindow).filter(
        MaintenanceWindow.id == window_id
    ).first()
    if not window:
        raise HTTPException(status_code=404, detail="天窗配置不存在")
    
    if window_update.start_time is not None:
        window.start_time = window_update.start_time
    if window_update.end_time is not None:
        window.end_time = window_update.end_time
    if window_update.is_general is not None:
        if window_update.is_general:
            existing_general = db.query(MaintenanceWindow).filter(
                MaintenanceWindow.line_id == window.line_id,
                MaintenanceWindow.is_general == True,
                MaintenanceWindow.id != window_id
            ).first()
            if existing_general:
                raise HTTPException(status_code=400, detail="该线路已存在通用配置")
            window.date = None
        elif not window.date:
            raise HTTPException(status_code=400, detail="改为非通用配置需要指定日期")
        window.is_general = window_update.is_general
    
    db.commit()
    db.refresh(window)
    return window


@router.delete("/windows/{window_id}")
def delete_window(window_id: int, db: Session = Depends(get_db)):
    window = db.query(MaintenanceWindow).filter(
        MaintenanceWindow.id == window_id
    ).first()
    if not window:
        raise HTTPException(status_code=404, detail="天窗配置不存在")
    
    db.delete(window)
    db.commit()
    return {"message": "天窗配置已删除"}
