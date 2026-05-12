from sqlalchemy.orm import Session
from app.models import Semaphore, Alert
from datetime import datetime
from typing import List, Optional


def get_semaphore(db: Session, device_id: str) -> Optional[Semaphore]:
    return db.query(Semaphore).filter(Semaphore.device_id == device_id).first()


def get_all_semaphores(db: Session) -> List[Semaphore]:
    return db.query(Semaphore).all()


def create_semaphore(db: Session, device_id: str, name: str) -> Semaphore:
    db_semaphore = Semaphore(device_id=device_id, name=name)
    db.add(db_semaphore)
    db.commit()
    db.refresh(db_semaphore)
    return db_semaphore


def open_semaphore(db: Session, device_id: str) -> dict:
    semaphore = get_semaphore(db, device_id)
    if not semaphore:
        return {"success": False, "message": "信号机不存在"}
    
    if semaphore.fault:
        return {"success": False, "message": "信号机故障，无法开放"}
    
    semaphore.status = "green"
    semaphore.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(semaphore)
    
    return {"success": True, "message": "信号机已开放", "semaphore": semaphore}


def close_semaphore(db: Session, device_id: str) -> dict:
    semaphore = get_semaphore(db, device_id)
    if not semaphore:
        return {"success": False, "message": "信号机不存在"}
    
    semaphore.status = "red"
    semaphore.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(semaphore)
    
    return {"success": True, "message": "信号机已关闭", "semaphore": semaphore}


def set_fault_status(db: Session, device_id: str, fault: bool) -> dict:
    semaphore = get_semaphore(db, device_id)
    if not semaphore:
        return {"success": False, "message": "信号机不存在"}
    
    semaphore.fault = fault
    if fault:
        semaphore.status = "red"
        alert = Alert(
            device_type="semaphore",
            device_id=device_id,
            message="信号机故障，已自动显示红灯"
        )
        db.add(alert)
    semaphore.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(semaphore)
    
    return {"success": True, "message": "信号机故障状态已更新", "semaphore": semaphore}
