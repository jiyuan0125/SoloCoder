from typing import List, Optional
from sqlalchemy.orm import Session
from .models import Berth, Ship, Operation, YardZone, YardItem
from datetime import datetime

def find_suitable_berth(db: Session, ship: Ship) -> Optional[Berth]:
    suitable = db.query(Berth).filter(
        Berth.type == ship.type,
        Berth.capacity >= ship.draft,
        Berth.is_under_maintenance == False,
        Berth.current_ship_id == None
    ).order_by(Berth.capacity).all()
    
    if suitable:
        return suitable[0]
    return None

def get_waiting_operations(db: Session) -> List[Operation]:
    return db.query(Operation).filter(
        Operation.status.in_(["queued", "waiting"])
    ).order_by(Operation.priority.desc(), Operation.created_at).all()

def has_active_operation(db: Session, ship_id: int) -> bool:
    return db.query(Operation).filter(
        Operation.ship_id == ship_id,
        Operation.status.in_(["queued", "waiting", "in_progress"])
    ).first() is not None

def validate_berth_for_ship(db: Session, berth: Berth, ship: Ship) -> tuple[bool, str]:
    if berth.type != ship.type:
        return False, f"泊位类型不匹配：泊位{berth.type} vs 船舶{ship.type}"
    if berth.capacity < ship.draft:
        return False, f"靠泊能力不足：泊位{berth.capacity} < 船舶吃水{ship.draft}"
    if berth.is_under_maintenance:
        return False, "泊位正在维护"
    if berth.current_ship_id is not None:
        return False, "泊位已被占用"
    return True, ""

def check_yard_capacity(db: Session, zone: YardZone, quantity: float) -> tuple[bool, str]:
    if zone.used_capacity + quantity > zone.total_capacity:
        return False, f"堆场容量不足：剩余{zone.total_capacity - zone.used_capacity}，需要{quantity}"
    if not zone.is_available:
        return False, "该堆场区域当前不可用"
    return True, ""

def start_operation(db: Session, operation: Operation) -> Operation:
    if operation.yard_zone_id:
        zone = db.query(YardZone).filter(YardZone.id == operation.yard_zone_id).first()
        if zone:
            ok, msg = check_yard_capacity(db, zone, operation.quantity)
            if not ok:
                operation.status = "waiting"
                db.commit()
                db.refresh(operation)
                return operation
            
            zone.used_capacity += operation.quantity
            item = YardItem(
                zone_id=zone.id,
                operation_id=operation.id,
                quantity=operation.quantity,
                status="in_storage"
            )
            db.add(item)
    
    berth = db.query(Berth).filter(Berth.id == operation.berth_id).first()
    ship = db.query(Ship).filter(Ship.id == operation.ship_id).first()
    
    if berth:
        berth.current_ship_id = ship.id
    
    if ship:
        ship.status = "loading"
    
    operation.status = "in_progress"
    operation.started_at = datetime.utcnow()
    
    db.commit()
    db.refresh(operation)
    return operation

def complete_operation(db: Session, operation: Operation) -> Operation:
    berth = db.query(Berth).filter(Berth.id == operation.berth_id).first()
    ship = db.query(Ship).filter(Ship.id == operation.ship_id).first()
    
    if operation.yard_zone_id and operation.quantity > 0:
        zone = db.query(YardZone).filter(YardZone.id == operation.yard_zone_id).first()
        if zone:
            zone.used_capacity = max(0, zone.used_capacity - operation.quantity)
            items = db.query(YardItem).filter(YardItem.operation_id == operation.id).all()
            for item in items:
                db.delete(item)
    
    if berth:
        berth.current_ship_id = None
    
    if ship:
        ship.status = "anchored"
    
    operation.status = "completed"
    operation.completed_at = datetime.utcnow()
    
    db.commit()
    db.refresh(operation)
    return operation

def handle_zone_unavailable(db: Session, zone_id: int):
    zone = db.query(YardZone).filter(YardZone.id == zone_id).first()
    if not zone:
        return
    
    zone.is_available = False
    
    items = db.query(YardItem).filter(
        YardItem.zone_id == zone_id,
        YardItem.status == "in_queue"
    ).all()
    
    for item in items:
        operation = db.query(Operation).filter(Operation.id == item.operation_id).first()
        if operation and operation.status == "queued":
            operation.status = "waiting"
        
        zone.used_capacity = max(0, zone.used_capacity - item.quantity)
        item.status = "waiting"
    
    db.commit()

def revalidate_ship_berth(db: Session, ship: Ship):
    if not ship.current_berth:
        return None, None
    
    berth = ship.current_berth
    ok, msg = validate_berth_for_ship(db, berth, ship)
    
    if ok:
        return True, "靠泊能力仍然满足"
    
    active_ops = db.query(Operation).filter(
        Operation.ship_id == ship.id,
        Operation.status.in_(["queued", "waiting", "in_progress"])
    ).all()
    
    for op in active_ops:
        if op.status == "in_progress":
            op.status = "waiting"
        elif op.status == "queued":
            op.status = "waiting"
    
    berth.current_ship_id = None
    ship.status = "anchored"
    
    db.commit()
    db.refresh(ship)
    
    return False, msg

def advance_ship_status(db: Session, ship: Ship, target_status: str) -> tuple[bool, str]:
    transitions = {
        "arriving": ["docked"],
        "docked": ["loading"],
        "loading": ["departing"],
        "departing": ["anchored", "arriving"],
        "anchored": ["arriving"]
    }
    
    if ship.status not in transitions:
        return False, f"未知状态: {ship.status}"
    
    if target_status not in transitions[ship.status]:
        return False, f"无法从 {ship.status} 状态转换到 {target_status}"
    
    if target_status == "docked" and ship.status == "arriving":
        suitable = find_suitable_berth(db, ship)
        if not suitable:
            return False, "没有可用的合适泊位"
        suitable.current_ship_id = ship.id
        ship.status = "docked"
        db.commit()
        db.refresh(ship)
        return True, f"船舶已靠泊到泊位 {suitable.name}"
    
    if target_status == "departing" and ship.status == "loading":
        active = has_active_operation(db, ship.id)
        if active:
            return False, "存在进行中的装卸作业，无法申请离港"
        ship.status = "departing"
        db.commit()
        db.refresh(ship)
        return True, "船舶已申请离港"
    
    if target_status == "anchored" and ship.status == "departing":
        if ship.current_berth:
            ship.current_berth.current_ship_id = None
        ship.status = "anchored"
        db.commit()
        db.refresh(ship)
        return True, "船舶已离港并进入锚泊状态"
    
    ship.status = target_status
    db.commit()
    db.refresh(ship)
    return True, f"状态已更新为 {target_status}"
