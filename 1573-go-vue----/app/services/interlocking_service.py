from sqlalchemy.orm import Session
from app.models import Interlocking, Switch, Semaphore, BlockSection, Alert
from app.services import switch_service, semaphore_service
from datetime import datetime
from typing import List, Dict


def get_interlocking(db: Session, device_id: str):
    return db.query(Interlocking).filter(Interlocking.device_id == device_id).first()


def get_all_interlockings(db: Session) -> List[Interlocking]:
    return db.query(Interlocking).all()


def create_interlocking(db: Session, device_id: str, name: str) -> Interlocking:
    db_interlocking = Interlocking(device_id=device_id, name=name)
    db.add(db_interlocking)
    db.commit()
    db.refresh(db_interlocking)
    return db_interlocking


def set_route(db: Session, device_id: str, route_request: Dict) -> dict:
    interlocking = get_interlocking(db, device_id)
    if not interlocking:
        return {"success": False, "message": "联锁设备不存在"}
    
    if interlocking.current_route:
        return {"success": False, "message": "已有进路已建立，请先取消"}
    
    interlocking.route_status = "planning"
    interlocking.last_updated = datetime.utcnow()
    db.commit()
    
    conflicting = check_conflicting_routes(db, route_request.get("conflicting_routes", []))
    if conflicting:
        interlocking.route_status = "failed"
        interlocking.last_updated = datetime.utcnow()
        db.commit()
        return {"success": False, "message": f"存在敌对进路冲突: {', '.join(conflicting)}"}
    
    switches = route_request.get("switches", [])
    for sw_id in switches:
        switch = switch_service.get_switch(db, sw_id)
        if switch and not switch.has_indication:
            interlocking.route_status = "failed"
            interlocking.last_updated = datetime.utcnow()
            db.commit()
            return {"success": False, "message": f"道岔{sw_id}失去表示，无法排路"}
    
    blocks = route_request.get("blocks", [])
    for block_id in blocks:
        block = db.query(BlockSection).filter(BlockSection.device_id == block_id).first()
        if block and block.occupied:
            interlocking.route_status = "failed"
            interlocking.last_updated = datetime.utcnow()
            db.commit()
            return {"success": False, "message": f"闭塞分区{block_id}已占用"}
    
    interlocking.route_status = "switching"
    interlocking.last_updated = datetime.utcnow()
    db.commit()
    
    for sw_id in switches:
        switch = switch_service.get_switch(db, sw_id)
        if switch:
            switch.locked = True
            switch.last_updated = datetime.utcnow()
    
    interlocking.route_status = "locked"
    interlocking.last_updated = datetime.utcnow()
    db.commit()
    
    semaphores = route_request.get("semaphores", [])
    for sem_id in semaphores:
        result = semaphore_service.open_semaphore(db, sem_id)
        if not result["success"]:
            interlocking.route_status = "failed"
            interlocking.last_updated = datetime.utcnow()
            db.commit()
            return {"success": False, "message": result["message"]}
    
    interlocking.current_route = route_request["route_id"]
    interlocking.route_status = "open"
    interlocking.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(interlocking)
    
    return {"success": True, "message": "进路已建立并开放", "interlocking": interlocking}


def cancel_route(db: Session, device_id: str) -> dict:
    interlocking = get_interlocking(db, device_id)
    if not interlocking:
        return {"success": False, "message": "联锁设备不存在"}
    
    if not interlocking.current_route:
        return {"success": False, "message": "无进路可取消"}
    
    semaphores = db.query(Semaphore).all()
    for sem in semaphores:
        if sem.status == "green":
            semaphore_service.close_semaphore(db, sem.device_id)
    
    switches = db.query(Switch).filter(Switch.locked == True).all()
    for sw in switches:
        sw.locked = False
        sw.last_updated = datetime.utcnow()
    
    interlocking.current_route = None
    interlocking.route_status = "idle"
    interlocking.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(interlocking)
    
    return {"success": True, "message": "进路已取消", "interlocking": interlocking}


def check_conflicting_routes(db: Session, conflicting_routes: List[str]) -> List[str]:
    conflicts = []
    for route_id in conflicting_routes:
        existing = db.query(Interlocking).filter(
            Interlocking.current_route == route_id
        ).first()
        if existing:
            conflicts.append(route_id)
    return conflicts
