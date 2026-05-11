from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from sqlalchemy.orm import Session
from .database import get_db
from .models import Berth, Ship, Operation, YardZone, YardItem
from . import scheduler

router = APIRouter()

class BerthCreate(BaseModel):
    name: str
    type: str
    capacity: float

class ShipCreate(BaseModel):
    name: str
    type: str
    draft: float
    eta: datetime

class ShipUpdate(BaseModel):
    name: Optional[str] = None
    type: Optional[str] = None
    draft: Optional[float] = None
    eta: Optional[datetime] = None

class OperationCreate(BaseModel):
    ship_id: int
    priority: int = 5
    yard_zone_id: Optional[int] = None
    quantity: float

class YardZoneCreate(BaseModel):
    name: str
    total_capacity: float

@router.get("/berths")
def list_berths(db: Session = Depends(get_db)):
    berths = db.query(Berth).all()
    return [
        {
            "id": b.id,
            "name": b.name,
            "type": b.type,
            "capacity": b.capacity,
            "is_under_maintenance": b.is_under_maintenance,
            "current_ship": b.ship.name if b.ship else None
        }
        for b in berths
    ]

@router.post("/berths")
def create_berth(data: BerthCreate, db: Session = Depends(get_db)):
    existing = db.query(Berth).filter(Berth.name == data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="泊位名称已存在")
    berth = Berth(name=data.name, type=data.type, capacity=data.capacity)
    db.add(berth)
    db.commit()
    db.refresh(berth)
    return {"id": berth.id, "message": "泊位创建成功"}

@router.post("/berths/{berth_id}/maintenance")
def toggle_maintenance(berth_id: int, db: Session = Depends(get_db)):
    berth = db.query(Berth).filter(Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="泊位不存在")
    if not berth.is_under_maintenance and berth.current_ship_id is not None:
        raise HTTPException(status_code=400, detail="泊位有船舶停靠，无法进入维护")
    berth.is_under_maintenance = not berth.is_under_maintenance
    db.commit()
    db.refresh(berth)
    return {"id": berth.id, "is_under_maintenance": berth.is_under_maintenance}

@router.get("/ships")
def list_ships(db: Session = Depends(get_db)):
    ships = db.query(Ship).all()
    return [
        {
            "id": s.id,
            "name": s.name,
            "type": s.type,
            "draft": s.draft,
            "status": s.status,
            "eta": s.eta.isoformat() if s.eta else None,
            "current_berth": s.current_berth.name if s.current_berth else None
        }
        for s in ships
    ]

@router.post("/ships")
def create_ship(data: ShipCreate, db: Session = Depends(get_db)):
    existing = db.query(Ship).filter(Ship.name == data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="船舶名称已存在")
    ship = Ship(name=data.name, type=data.type, draft=data.draft, eta=data.eta, status="arriving")
    db.add(ship)
    db.commit()
    db.refresh(ship)
    return {"id": ship.id, "message": "船舶创建成功"}

@router.patch("/ships/{ship_id}")
def update_ship(ship_id: int, data: ShipUpdate, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    old_draft = ship.draft
    
    if data.name is not None:
        ship.name = data.name
    if data.type is not None:
        ship.type = data.type
    if data.draft is not None:
        ship.draft = data.draft
    if data.eta is not None:
        ship.eta = data.eta
    
    db.commit()
    db.refresh(ship)
    
    result = {"message": "船舶信息已更新"}
    
    if data.draft is not None and data.draft != old_draft:
        ok, msg = scheduler.revalidate_ship_berth(db, ship)
        result["revalidation"] = {"passed": ok, "message": msg}
    
    return result

@router.post("/ships/{ship_id}/status/{target}")
def change_ship_status(ship_id: int, target: str, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    ok, msg = scheduler.advance_ship_status(db, ship, target)
    if not ok:
        raise HTTPException(status_code=400, detail=msg)
    return {"message": msg}

@router.get("/operations")
def list_operations(db: Session = Depends(get_db)):
    ops = db.query(Operation).all()
    return [
        {
            "id": o.id,
            "ship": o.ship.name if o.ship else None,
            "berth": o.berth.name if o.berth else None,
            "priority": o.priority,
            "yard_zone": o.yard_zone.name if o.yard_zone else None,
            "quantity": o.quantity,
            "status": o.status,
            "created_at": o.created_at.isoformat() if o.created_at else None,
            "started_at": o.started_at.isoformat() if o.started_at else None,
            "completed_at": o.completed_at.isoformat() if o.completed_at else None
        }
        for o in ops
    ]

@router.post("/operations")
def create_operation(data: OperationCreate, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == data.ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    if scheduler.has_active_operation(db, ship.id):
        raise HTTPException(status_code=400, detail="该船舶已有活跃的装卸作业")
    
    if ship.status not in ["docked", "loading"]:
        raise HTTPException(status_code=400, detail="船舶未靠泊，无法安排装卸作业")
    
    if not ship.current_berth:
        raise HTTPException(status_code=400, detail="船舶未分配到泊位")
    
    berth = ship.current_berth
    
    if data.yard_zone_id:
        zone = db.query(YardZone).filter(YardZone.id == data.yard_zone_id).first()
        if not zone:
            raise HTTPException(status_code=404, detail="堆场区域不存在")
        
        ok, msg = scheduler.check_yard_capacity(db, zone, data.quantity)
        if not ok:
            raise HTTPException(status_code=400, detail=msg)
    
    op = Operation(
        ship_id=ship.id,
        berth_id=berth.id,
        priority=data.priority,
        yard_zone_id=data.yard_zone_id,
        quantity=data.quantity,
        status="queued"
    )
    db.add(op)
    db.commit()
    db.refresh(op)
    
    return {"id": op.id, "message": "装卸作业已创建"}

@router.post("/operations/{op_id}/start")
def start_operation(op_id: int, db: Session = Depends(get_db)):
    op = db.query(Operation).filter(Operation.id == op_id).first()
    if not op:
        raise HTTPException(status_code=404, detail="作业不存在")
    if op.status not in ["queued", "waiting"]:
        raise HTTPException(status_code=400, detail=f"作业状态为 {op.status}，无法开始")
    
    result = scheduler.start_operation(db, op)
    return {"id": result.id, "status": result.status, "message": "作业已开始" if result.status == "in_progress" else "作业进入等待状态"}

@router.post("/operations/{op_id}/complete")
def complete_operation(op_id: int, db: Session = Depends(get_db)):
    op = db.query(Operation).filter(Operation.id == op_id).first()
    if not op:
        raise HTTPException(status_code=404, detail="作业不存在")
    if op.status != "in_progress":
        raise HTTPException(status_code=400, detail=f"作业状态为 {op.status}，无法完成")
    
    result = scheduler.complete_operation(db, op)
    return {"id": result.id, "status": result.status, "message": "作业已完成"}

@router.get("/schedule")
def get_schedule(db: Session = Depends(get_db)):
    waiting = scheduler.get_waiting_operations(db)
    recommended = []
    
    for op in waiting:
        ship = db.query(Ship).filter(Ship.id == op.ship_id).first()
        if ship:
            berth = scheduler.find_suitable_berth(db, ship)
            recommended.append({
                "operation_id": op.id,
                "ship": ship.name,
                "priority": op.priority,
                "recommended_berth": berth.name if berth else None
            })
    
    return {
        "waiting_queue": [
            {
                "operation_id": w.id,
                "ship": w.ship.name if w.ship else None,
                "priority": w.priority,
                "status": w.status,
                "created_at": w.created_at.isoformat()
            }
            for w in waiting
        ],
        "recommendations": recommended
    }

@router.get("/yard")
def list_yard_zones(db: Session = Depends(get_db)):
    zones = db.query(YardZone).all()
    return [
        {
            "id": z.id,
            "name": z.name,
            "total_capacity": z.total_capacity,
            "used_capacity": z.used_capacity,
            "available_capacity": z.total_capacity - z.used_capacity,
            "is_available": z.is_available
        }
        for z in zones
    ]

@router.post("/yard")
def create_yard_zone(data: YardZoneCreate, db: Session = Depends(get_db)):
    existing = db.query(YardZone).filter(YardZone.name == data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="堆场区域名称已存在")
    zone = YardZone(name=data.name, total_capacity=data.total_capacity, used_capacity=0)
    db.add(zone)
    db.commit()
    db.refresh(zone)
    return {"id": zone.id, "message": "堆场区域创建成功"}

@router.post("/yard/{zone_id}/toggle")
def toggle_yard_zone(zone_id: int, db: Session = Depends(get_db)):
    zone = db.query(YardZone).filter(YardZone.id == zone_id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="堆场区域不存在")
    
    if zone.is_available:
        scheduler.handle_zone_unavailable(db, zone.id)
    else:
        zone.is_available = True
        db.commit()
        db.refresh(zone)
    
    return {"id": zone.id, "is_available": zone.is_available}
