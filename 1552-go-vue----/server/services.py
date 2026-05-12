from datetime import datetime, timedelta
from typing import Optional
from sqlalchemy.orm import Session
from fastapi import HTTPException, status
from . import models, schemas
from .config import (
    FUEL_PRICE_PER_TON,
    WATER_PRICE_PER_TON,
    WASTE_PRICE_PER_KG,
    WASTE_DISCOUNT_THRESHOLD_KG,
    WASTE_DISCOUNT_RATE,
    WAIT_TIMEOUT_HOURS
)


STATUS_TRANSITIONS = {
    models.AgentServiceStatus.ACCEPTED: [models.AgentServiceStatus.DECLARED],
    models.AgentServiceStatus.DECLARED: [models.AgentServiceStatus.BERTHING_ARRANGED],
    models.AgentServiceStatus.BERTHING_ARRANGED: [models.AgentServiceStatus.OPERATION_EXECUTED],
    models.AgentServiceStatus.OPERATION_EXECUTED: [models.AgentServiceStatus.SETTLED],
    models.AgentServiceStatus.SETTLED: [models.AgentServiceStatus.DEPARTURE_CONFIRMED],
    models.AgentServiceStatus.DEPARTURE_CONFIRMED: [],
}


def can_transition(current: models.AgentServiceStatus, target: models.AgentServiceStatus) -> bool:
    return target in STATUS_TRANSITIONS.get(current, [])


def create_agent_service(db: Session, service: schemas.AgentServiceCreate) -> models.AgentService:
    db_service = models.AgentService(
        ship_name=service.ship_name,
        imo_number=service.imo_number,
        port_of_call=service.port_of_call,
        arrival_time=service.arrival_time,
        estimated_departure_time=service.estimated_departure_time,
        agency_fee=service.agency_fee,
        notes=service.notes,
        status=models.AgentServiceStatus.ACCEPTED,
    )
    db.add(db_service)
    db.commit()
    db.refresh(db_service)
    return db_service


def transition_status(db: Session, service_id: int, target: models.AgentServiceStatus) -> models.AgentService:
    service = db.query(models.AgentService).filter(models.AgentService.id == service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    if not can_transition(service.status, target):
        raise HTTPException(
            status_code=400,
            detail=f"无法从状态 {service.status.value} 转换到 {target.value}"
        )
    
    if target == models.AgentServiceStatus.BERTHING_ARRANGED:
        if not service.berthing_request or service.berthing_request.status != models.BerthingRequestStatus.ASSIGNED:
            raise HTTPException(status_code=400, detail="靠泊安排未完成，无法进入下一环节")
    
    if target == models.AgentServiceStatus.SETTLED:
        pass
    
    if target == models.AgentServiceStatus.DEPARTURE_CONFIRMED:
        _validate_departure_conditions(db, service)
    
    service.status = target
    db.commit()
    db.refresh(service)
    return service


def _validate_departure_conditions(db: Session, service: models.AgentService):
    pending_deliveries = db.query(models.MaterialDelivery).filter(
        models.MaterialDelivery.agent_service_id == service.id,
        models.MaterialDelivery.status == models.DeliveryStatus.PENDING
    ).count()
    
    pending_waste = db.query(models.WasteCollection).filter(
        models.WasteCollection.agent_service_id == service.id,
        models.WasteCollection.status == models.WasteCollectionStatus.PENDING
    ).count()
    
    if pending_deliveries > 0:
        raise HTTPException(status_code=400, detail="还有物料补给未交付，无法确认离港")
    if pending_waste > 0:
        raise HTTPException(status_code=400, detail="还有垃圾回收未完成，无法确认离港")


def create_berth(db: Session, berth: schemas.BerthCreate) -> models.Berth:
    existing = db.query(models.Berth).filter(models.Berth.name == berth.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="泊位名称已存在")
    db_berth = models.Berth(**berth.model_dump())
    db.add(db_berth)
    db.commit()
    db.refresh(db_berth)
    return db_berth


def create_berthing_request(db: Session, req: schemas.BerthingRequestCreate) -> models.BerthingRequest:
    service = db.query(models.AgentService).filter(models.AgentService.id == req.agent_service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    if service.berthing_request:
        raise HTTPException(status_code=400, detail="该代理服务已有靠泊申请")
    
    if req.requested_berthing_time <= datetime.utcnow():
        raise HTTPException(status_code=400, detail="申请靠泊时间必须是未来时间")
    
    if not (2 <= req.estimated_duration_hours <= 72):
        raise HTTPException(status_code=400, detail="预计靠泊时长必须在2到72小时之间")
    
    db_req = models.BerthingRequest(
        agent_service_id=req.agent_service_id,
        requested_berthing_time=req.requested_berthing_time,
        estimated_duration_hours=req.estimated_duration_hours,
        berth_preference=req.berth_preference,
        status=models.BerthingRequestStatus.PENDING,
    )
    db.add(db_req)
    db.commit()
    db.refresh(db_req)
    return db_req


def approve_berthing_request(db: Session, request_id: int, data: schemas.BerthingRequestApprove) -> models.BerthingRequest:
    req = db.query(models.BerthingRequest).filter(models.BerthingRequest.id == request_id).first()
    if not req:
        raise HTTPException(status_code=404, detail="靠泊申请不存在")
    
    if req.status != models.BerthingRequestStatus.PENDING:
        raise HTTPException(status_code=400, detail="靠泊申请状态不允许审批")
    
    req.status = models.BerthingRequestStatus.APPROVED
    req.approved_by = data.approved_by
    req.approval_time = datetime.utcnow()
    
    time_diff = req.requested_berthing_time - req.agent_service.arrival_time
    if time_diff.total_seconds() > WAIT_TIMEOUT_HOURS * 3600:
        req.is_timeout = True
    
    if data.auto_assign_berth:
        assigned_berth = _auto_assign_berth(db, req)
        if assigned_berth:
            req.assigned_berth_id = assigned_berth.id
            req.actual_berthing_time = req.requested_berthing_time
            req.status = models.BerthingRequestStatus.ASSIGNED
            assigned_berth.is_available = False
    
    db.commit()
    db.refresh(req)
    return req


def _auto_assign_berth(db: Session, req: models.BerthingRequest) -> Optional[models.Berth]:
    preferred = None
    if req.berth_preference:
        preferred = db.query(models.Berth).filter(
            models.Berth.name == req.berth_preference,
            models.Berth.is_available == True
        ).first()
    
    if preferred:
        return preferred
    
    return db.query(models.Berth).filter(models.Berth.is_available == True).first()


def create_material_delivery(db: Session, delivery: schemas.MaterialDeliveryCreate) -> models.MaterialDelivery:
    service = db.query(models.AgentService).filter(models.AgentService.id == delivery.agent_service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    existing_deliveries = db.query(models.MaterialDelivery).filter(
        models.MaterialDelivery.agent_service_id == delivery.agent_service_id
    ).all()
    
    existing_types = {d.material_type for d in existing_deliveries}
    if delivery.material_type == models.MaterialType.FUEL and models.MaterialType.FRESH_WATER in existing_types:
        raise HTTPException(status_code=400, detail="已安排淡水配送，不能同时安排燃油配送")
    if delivery.material_type == models.MaterialType.FRESH_WATER and models.MaterialType.FUEL in existing_types:
        raise HTTPException(status_code=400, detail="已安排燃油配送，不能同时安排淡水配送")
    
    unit_price = FUEL_PRICE_PER_TON if delivery.material_type == models.MaterialType.FUEL else WATER_PRICE_PER_TON
    total_price = delivery.quantity_tons * unit_price
    
    db_delivery = models.MaterialDelivery(
        agent_service_id=delivery.agent_service_id,
        material_type=delivery.material_type,
        quantity_tons=delivery.quantity_tons,
        unit_price=unit_price,
        total_price=total_price,
        status=models.DeliveryStatus.PENDING,
        notes=delivery.notes,
    )
    db.add(db_delivery)
    db.commit()
    db.refresh(db_delivery)
    return db_delivery


def complete_material_delivery(db: Session, delivery_id: int) -> models.MaterialDelivery:
    delivery = db.query(models.MaterialDelivery).filter(models.MaterialDelivery.id == delivery_id).first()
    if not delivery:
        raise HTTPException(status_code=404, detail="物料配送不存在")
    
    if delivery.status == models.DeliveryStatus.DELIVERED:
        raise HTTPException(status_code=400, detail="物料配送已完成")
    
    delivery.status = models.DeliveryStatus.DELIVERED
    delivery.delivery_time = datetime.utcnow()
    db.commit()
    db.refresh(delivery)
    return delivery


def calculate_waste_fee(weight_kg: float) -> tuple:
    if weight_kg <= WASTE_DISCOUNT_THRESHOLD_KG:
        total = weight_kg * WASTE_PRICE_PER_KG
        return WASTE_PRICE_PER_KG, total
    else:
        normal_amount = WASTE_DISCOUNT_THRESHOLD_KG * WASTE_PRICE_PER_KG
        excess = weight_kg - WASTE_DISCOUNT_THRESHOLD_KG
        excess_amount = excess * WASTE_PRICE_PER_KG * WASTE_DISCOUNT_RATE
        total = normal_amount + excess_amount
        effective_unit = total / weight_kg
        return effective_unit, total


def create_waste_collection(db: Session, collection: schemas.WasteCollectionCreate) -> models.WasteCollection:
    service = db.query(models.AgentService).filter(models.AgentService.id == collection.agent_service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    unit_price, total_price = calculate_waste_fee(collection.weight_kg)
    
    db_collection = models.WasteCollection(
        agent_service_id=collection.agent_service_id,
        waste_type=collection.waste_type,
        weight_kg=collection.weight_kg,
        unit_price=unit_price,
        total_price=total_price,
        status=models.WasteCollectionStatus.PENDING,
        notes=collection.notes,
    )
    db.add(db_collection)
    db.commit()
    db.refresh(db_collection)
    return db_collection


def complete_waste_collection(db: Session, collection_id: int) -> models.WasteCollection:
    collection = db.query(models.WasteCollection).filter(models.WasteCollection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="垃圾回收不存在")
    
    if collection.status == models.WasteCollectionStatus.COMPLETED:
        raise HTTPException(status_code=400, detail="垃圾回收已完成")
    
    collection.status = models.WasteCollectionStatus.COMPLETED
    collection.collection_time = datetime.utcnow()
    db.commit()
    db.refresh(collection)
    return collection


def create_todo(db: Session, todo: schemas.TodoCreate) -> models.Todo:
    service = db.query(models.AgentService).filter(models.AgentService.id == todo.agent_service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    db_todo = models.Todo(
        agent_service_id=todo.agent_service_id,
        title=todo.title,
        description=todo.description,
        due_time=todo.due_time,
        status=models.TodoStatus.PENDING,
    )
    db.add(db_todo)
    db.commit()
    db.refresh(db_todo)
    return db_todo


def complete_todo(db: Session, todo_id: int) -> models.Todo:
    todo = db.query(models.Todo).filter(models.Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    
    if todo.status == models.TodoStatus.COMPLETED:
        raise HTTPException(status_code=400, detail="待办事项已完成")
    
    service = todo.agent_service
    pending_todos = [t for t in service.todos if t.status == models.TodoStatus.PENDING]
    
    if pending_todos and pending_todos[0].id != todo.id:
        raise HTTPException(status_code=400, detail="请先完成更早的待办事项")
    
    todo.status = models.TodoStatus.COMPLETED
    todo.completed_time = datetime.utcnow()
    db.commit()
    db.refresh(todo)
    return todo


def create_settlement(db: Session, settlement: schemas.SettlementCreate) -> models.Settlement:
    service = db.query(models.AgentService).filter(models.AgentService.id == settlement.agent_service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    if service.status != models.AgentServiceStatus.OPERATION_EXECUTED:
        raise HTTPException(status_code=400, detail="只有作业执行完成后才能进行费用结算")
    
    material_fee = sum(
        d.total_price for d in service.material_deliveries
        if d.status == models.DeliveryStatus.DELIVERED
    ) if service.material_deliveries else 0
    
    waste_fee = sum(
        w.total_price for w in service.waste_collections
        if w.status == models.WasteCollectionStatus.COMPLETED
    ) if service.waste_collections else 0
    
    agency_fee = service.agency_fee
    total_amount = agency_fee + material_fee + waste_fee
    
    db_settlement = models.Settlement(
        agent_service_id=settlement.agent_service_id,
        agency_fee=agency_fee,
        material_fee=material_fee,
        waste_fee=waste_fee,
        total_amount=total_amount,
        settlement_time=datetime.utcnow(),
        notes=settlement.notes,
    )
    db.add(db_settlement)
    
    service.material_fee = material_fee
    service.waste_fee = waste_fee
    service.total_fee = total_amount
    service.status = models.AgentServiceStatus.SETTLED
    
    db.commit()
    db.refresh(db_settlement)
    db.refresh(service)
    return db_settlement


def confirm_departure(db: Session, service_id: int) -> models.AgentService:
    service = db.query(models.AgentService).filter(models.AgentService.id == service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    
    if service.status != models.AgentServiceStatus.SETTLED:
        raise HTTPException(status_code=400, detail="只有费用结算完成后才能确认离港")
    
    _validate_departure_conditions(db, service)
    
    service.status = models.AgentServiceStatus.DEPARTURE_CONFIRMED
    if service.berthing_request and service.berthing_request.assigned_berth:
        service.berthing_request.assigned_berth.is_available = True
    
    db.commit()
    db.refresh(service)
    return service
