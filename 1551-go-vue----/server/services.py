from datetime import datetime
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from fastapi import HTTPException
from server.models import Berth, Ship, HandlingOperation, YardArea


class PortService:
    @staticmethod
    def get_available_berths(db: Session, ship_type: str, draught: float) -> List[Berth]:
        berths = db.query(Berth).filter(
            Berth.berth_type == ship_type,
            Berth.capacity >= draught,
            Berth.is_under_maintenance == False
        ).all()
        
        available = []
        for berth in berths:
            if berth.check_available():
                available.append(berth)
        
        return available

    @staticmethod
    def get_waiting_operations(db: Session) -> List[HandlingOperation]:
        return db.query(HandlingOperation).filter(
            HandlingOperation.status.in_(["waiting", "pending"])
        ).order_by(
            HandlingOperation.priority.desc(),
            HandlingOperation.created_at.asc()
        ).all()

    @staticmethod
    def get_scheduling_recommendation(db: Session, ship: Ship) -> Tuple[List[Berth], List[HandlingOperation]]:
        available_berths = PortService.get_available_berths(db, ship.ship_type, ship.draught)
        waiting_ops = PortService.get_waiting_operations(db)
        return available_berths, waiting_ops

    @staticmethod
    def create_handling_operation(db: Session, ship_id: int, priority: int, cargo_volume: float,
                                   description: Optional[str] = None, berth_id: Optional[int] = None,
                                   yard_area_id: Optional[int] = None) -> HandlingOperation:
        ship = db.query(Ship).filter(Ship.id == ship_id).first()
        if not ship:
            raise HTTPException(status_code=404, detail="Ship not found")
        
        if ship.has_active_operation():
            raise HTTPException(status_code=400, detail="Ship already has an active operation")
        
        if ship.status not in ["docked", "at_anchor"]:
            raise HTTPException(status_code=400, detail="Ship must be docked or at anchor to create operation")
        
        if yard_area_id:
            yard = db.query(YardArea).filter(YardArea.id == yard_area_id).first()
            if not yard:
                raise HTTPException(status_code=404, detail="Yard area not found")
            if not yard.is_available:
                raise HTTPException(status_code=400, detail="Yard area is not available")
            if not yard.has_capacity(cargo_volume, db):
                raise HTTPException(status_code=400, detail="Yard area has insufficient capacity")
        
        selected_berth_id = berth_id
        if selected_berth_id:
            berth = db.query(Berth).filter(Berth.id == selected_berth_id).first()
            if not berth:
                raise HTTPException(status_code=404, detail="Berth not found")
            if berth.berth_type != ship.ship_type:
                raise HTTPException(
                    status_code=400,
                    detail=f"Berth type mismatch: berth is {berth.berth_type}, ship is {ship.ship_type}"
                )
            if berth.capacity < ship.draught:
                raise HTTPException(
                    status_code=400,
                    detail=f"Berth capacity {berth.capacity}m is less than ship draught {ship.draught}m"
                )
            if berth.is_under_maintenance:
                raise HTTPException(status_code=400, detail="Berth is under maintenance")
            if not berth.check_available():
                raise HTTPException(status_code=400, detail="Berth is already occupied")
        else:
            available_berths = PortService.get_available_berths(db, ship.ship_type, ship.draught)
            if available_berths:
                selected_berth_id = available_berths[0].id
        
        operation = HandlingOperation(
            ship_id=ship_id,
            berth_id=selected_berth_id,
            yard_area_id=yard_area_id,
            priority=priority,
            cargo_volume=cargo_volume,
            status="in_progress" if selected_berth_id and yard_area_id else "waiting",
            description=description
        )
        
        if selected_berth_id:
            ship.status = "docked"
            ship.current_berth_id = selected_berth_id
            operation.started_at = datetime.utcnow()
        
        db.add(operation)
        db.commit()
        db.refresh(operation)
        db.refresh(ship)
        
        return operation

    @staticmethod
    def complete_handling_operation(db: Session, operation_id: int) -> HandlingOperation:
        operation = db.query(HandlingOperation).filter(HandlingOperation.id == operation_id).first()
        if not operation:
            raise HTTPException(status_code=404, detail="Operation not found")
        
        if operation.status == "completed":
            return operation
        
        operation.status = "completed"
        operation.completed_at = datetime.utcnow()
        
        ship = db.query(Ship).filter(Ship.id == operation.ship_id).first()
        if ship:
            ship.status = "at_anchor"
            ship.current_berth_id = None
        
        db.commit()
        db.refresh(operation)
        db.refresh(ship)
        
        PortService.process_waiting_queue(db)
        
        return operation

    @staticmethod
    def update_ship_draught(db: Session, ship_id: int, new_draught: float) -> Ship:
        ship = db.query(Ship).filter(Ship.id == ship_id).first()
        if not ship:
            raise HTTPException(status_code=404, detail="Ship not found")
        
        old_draught = ship.draught
        ship.draught = new_draught
        
        if ship.current_berth_id:
            berth = db.query(Berth).filter(Berth.id == ship.current_berth_id).first()
            if berth and berth.capacity < new_draught:
                for op in ship.handling_operations:
                    if op.status == "in_progress":
                        op.status = "waiting"
                        op.berth_id = None
                        ship.status = "at_anchor"
                        ship.current_berth_id = None
                        break
        
        db.commit()
        db.refresh(ship)
        
        PortService.process_waiting_queue(db)
        
        return ship

    @staticmethod
    def mark_yard_area_unavailable(db: Session, yard_area_id: int):
        yard = db.query(YardArea).filter(YardArea.id == yard_area_id).first()
        if not yard:
            raise HTTPException(status_code=404, detail="Yard area not found")
        
        yard.is_available = False
        
        for op in yard.handling_operations:
            if op.status in ["pending", "in_progress"]:
                op.status = "waiting"
                op.yard_area_id = None
        
        db.commit()

    @staticmethod
    def mark_yard_area_available(db: Session, yard_area_id: int):
        yard = db.query(YardArea).filter(YardArea.id == yard_area_id).first()
        if not yard:
            raise HTTPException(status_code=404, detail="Yard area not found")
        
        yard.is_available = True
        db.commit()
        
        PortService.process_waiting_queue(db)

    @staticmethod
    def process_waiting_queue(db: Session):
        waiting_ops = PortService.get_waiting_operations(db)
        
        for op in waiting_ops:
            ship = db.query(Ship).filter(Ship.id == op.ship_id).first()
            if not ship:
                continue
            
            if not op.berth_id:
                available_berths = PortService.get_available_berths(db, ship.ship_type, ship.draught)
                if available_berths:
                    op.berth_id = available_berths[0].id
                    ship.current_berth_id = available_berths[0].id
                    ship.status = "docked"
            
            if op.berth_id and not op.yard_area_id:
                yards = db.query(YardArea).filter(YardArea.is_available == True).all()
                for yard in yards:
                    if yard.has_capacity(op.cargo_volume, db):
                        op.yard_area_id = yard.id
                        break
            
            if op.berth_id and op.yard_area_id:
                op.status = "in_progress"
                if not op.started_at:
                    op.started_at = datetime.utcnow()
        
        db.commit()

    @staticmethod
    def update_ship_status(db: Session, ship_id: int, new_status: str) -> Ship:
        ship = db.query(Ship).filter(Ship.id == ship_id).first()
        if not ship:
            raise HTTPException(status_code=404, detail="Ship not found")
        
        valid_transitions = {
            "expected": ["arrived"],
            "arrived": ["at_anchor"],
            "at_anchor": ["docked", "departed"],
            "docked": ["at_anchor", "departed"],
            "departed": []
        }
        
        if new_status not in valid_transitions.get(ship.status, []):
            raise HTTPException(
                status_code=400,
                detail=f"Invalid status transition from {ship.status} to {new_status}"
            )
        
        if new_status == "docked":
            if not ship.has_active_operation():
                raise HTTPException(status_code=400, detail="Ship must have active operation to dock")
        
        if new_status == "departed":
            if ship.has_active_operation():
                raise HTTPException(status_code=400, detail="Ship has active operations")
        
        ship.status = new_status
        db.commit()
        db.refresh(ship)
        
        return ship
