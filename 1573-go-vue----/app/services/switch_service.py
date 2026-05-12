from sqlalchemy.orm import Session
from app.models import Switch, SwitchLog, Alert
from datetime import datetime
from typing import List, Optional


def get_switch(db: Session, device_id: str) -> Optional[Switch]:
    return db.query(Switch).filter(Switch.device_id == device_id).first()


def get_all_switches(db: Session) -> List[Switch]:
    return db.query(Switch).all()


def create_switch(db: Session, device_id: str, name: str) -> Switch:
    db_switch = Switch(device_id=device_id, name=name)
    db.add(db_switch)
    db.commit()
    db.refresh(db_switch)
    return db_switch


def switch_position(db: Session, device_id: str, position: str) -> dict:
    switch = get_switch(db, device_id)
    if not switch:
        return {"success": False, "message": "道岔不存在"}
    
    if switch.occupied:
        return {"success": False, "message": "道岔被占用，无法转换"}
    if switch.locked:
        return {"success": False, "message": "道岔已锁闭，无法转换"}
    
    switch.position = position
    switch.last_updated = datetime.utcnow()
    
    log = SwitchLog(
        switch_id=device_id,
        action=f"position_changed_to_{position}",
        is_fault=False
    )
    db.add(log)
    db.commit()
    db.refresh(switch)
    
    return {"success": True, "message": f"道岔已转换到{position}位置", "switch": switch}


def lock_switch(db: Session, device_id: str, locked: bool) -> dict:
    switch = get_switch(db, device_id)
    if not switch:
        return {"success": False, "message": "道岔不存在"}
    
    switch.locked = locked
    switch.last_updated = datetime.utcnow()
    
    action = "locked" if locked else "unlocked"
    log = SwitchLog(
        switch_id=device_id,
        action=action,
        is_fault=False
    )
    db.add(log)
    db.commit()
    db.refresh(switch)
    
    return {"success": True, "message": f"道岔已{'锁闭' if locked else '解锁'}", "switch": switch}


def set_occupied(db: Session, device_id: str, occupied: bool) -> dict:
    switch = get_switch(db, device_id)
    if not switch:
        return {"success": False, "message": "道岔不存在"}
    
    switch.occupied = occupied
    switch.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(switch)
    
    return {"success": True, "message": f"道岔占用状态已更新", "switch": switch}


def set_indication(db: Session, device_id: str, has_indication: bool) -> dict:
    switch = get_switch(db, device_id)
    if not switch:
        return {"success": False, "message": "道岔不存在"}
    
    switch.has_indication = has_indication
    if not has_indication:
        alert = Alert(
            device_type="switch",
            device_id=device_id,
            message=f"道岔失去表示"
        )
        db.add(alert)
        log = SwitchLog(
            switch_id=device_id,
            action="indication_lost",
            is_fault=True
        )
        db.add(log)
    switch.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(switch)
    
    return {"success": True, "message": "道岔表示状态已更新", "switch": switch}
