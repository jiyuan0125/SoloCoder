from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from fastapi import HTTPException, status

from server.models import (
    AgentService,
    StageHistory,
    Berth,
    BerthApplication,
    Supply,
    WasteRecovery,
    Todo,
    FeeSettlement,
)
from server.schemas import (
    AgentServiceCreate,
    AgentServiceUpdate,
    BerthCreate,
    BerthApplicationCreate,
    SupplyCreate,
    WasteRecoveryCreate,
    TodoCreate,
    FeeSettlementCreate,
)
from server.enums import (
    AgentServiceStage,
    StageStatus,
    BerthStatus,
    BerthApplicationStatus,
    SupplyType,
    SupplyStatus,
    WasteStatus,
    TodoStatus,
    SettlementStatus,
)


STAGE_ORDER = [
    AgentServiceStage.ORDER_RECEIVED.value,
    AgentServiceStage.DECLARATION.value,
    AgentServiceStage.BERTHING_ARRANGEMENT.value,
    AgentServiceStage.OPERATION_EXECUTION.value,
    AgentServiceStage.FEE_SETTLEMENT.value,
    AgentServiceStage.DEPARTURE_CONFIRMED.value,
]


def _get_stage_index(stage: str) -> int:
    try:
        return STAGE_ORDER.index(stage)
    except ValueError:
        raise HTTPException(status_code=400, detail=f"Invalid stage: {stage}")


def create_agent_service(db: Session, service_in: AgentServiceCreate) -> AgentService:
    db_service = AgentService(
        ship_name=service_in.ship_name,
        imo_number=service_in.imo_number,
        captain_name=service_in.captain_name,
        arrival_time=service_in.arrival_time,
        current_stage=AgentServiceStage.ORDER_RECEIVED.value,
    )
    db.add(db_service)
    db.commit()
    db.refresh(db_service)

    for stage in STAGE_ORDER:
        db_stage = StageHistory(
            agent_service_id=db_service.id,
            stage=stage,
            status=StageStatus.PENDING.value if stage != AgentServiceStage.ORDER_RECEIVED.value else StageStatus.COMPLETED.value,
            completed_at=datetime.now() if stage == AgentServiceStage.ORDER_RECEIVED.value else None,
        )
        db.add(db_stage)

    db.commit()
    db.refresh(db_service)
    return db_service


def get_agent_services(db: Session, skip: int = 0, limit: int = 100) -> List[AgentService]:
    return db.query(AgentService).offset(skip).limit(limit).all()


def get_agent_service(db: Session, service_id: int) -> Optional[AgentService]:
    return db.query(AgentService).filter(AgentService.id == service_id).first()


def update_agent_service(db: Session, service_id: int, service_in: AgentServiceUpdate) -> Optional[AgentService]:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        return None

    update_data = service_in.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_service, key, value)

    db.commit()
    db.refresh(db_service)
    return db_service


def advance_stage(db: Session, service_id: int, notes: Optional[str] = None) -> AgentService:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    current_index = _get_stage_index(db_service.current_stage)

    if current_index >= len(STAGE_ORDER) - 1:
        raise HTTPException(status_code=400, detail="Already in final stage")

    if db_service.current_stage == AgentServiceStage.ORDER_RECEIVED.value:
        _validate_order_received_stage(db, db_service)
    elif db_service.current_stage == AgentServiceStage.DECLARATION.value:
        _validate_declaration_stage(db, db_service)
    elif db_service.current_stage == AgentServiceStage.BERTHING_ARRANGEMENT.value:
        _validate_berthing_stage(db, db_service)
    elif db_service.current_stage == AgentServiceStage.OPERATION_EXECUTION.value:
        _validate_operation_stage(db, db_service)
    elif db_service.current_stage == AgentServiceStage.FEE_SETTLEMENT.value:
        _validate_fee_settlement_stage(db, db_service)

    next_stage = STAGE_ORDER[current_index + 1]
    db_service.current_stage = next_stage

    db_stage = db.query(StageHistory).filter(
        StageHistory.agent_service_id == service_id,
        StageHistory.stage == next_stage,
    ).first()

    if db_stage:
        db_stage.status = StageStatus.COMPLETED.value
        db_stage.completed_at = datetime.now()
        if notes:
            db_stage.notes = notes

    db.commit()
    db.refresh(db_service)
    return db_service


def _validate_order_received_stage(db: Session, service: AgentService):
    pass


def _validate_declaration_stage(db: Session, service: AgentService):
    pass


def _validate_berthing_stage(db: Session, service: AgentService):
    if not service.berth_application:
        raise HTTPException(
            status_code=400,
            detail="No berth application found. Please submit a berth application first."
        )
    if service.berth_application.status != BerthApplicationStatus.APPROVED.value:
        raise HTTPException(
            status_code=400,
            detail="Berth application must be approved before advancing from berthing arrangement stage."
        )


def _validate_operation_stage(db: Session, service: AgentService):
    pass


def _validate_fee_settlement_stage(db: Session, service: AgentService):
    pending_todos = db.query(Todo).filter(
        Todo.agent_service_id == service.id,
        Todo.status == TodoStatus.PENDING.value,
    ).count()
    if pending_todos > 0:
        raise HTTPException(
            status_code=400,
            detail=f"There are {pending_todos} pending todos. All todos must be completed before fee settlement."
        )


def validate_departure_preconditions(db: Session, service_id: int) -> dict:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    pending_supplies = db.query(Supply).filter(
        Supply.agent_service_id == service_id,
        Supply.status == SupplyStatus.SCHEDULED.value,
    ).count()

    if pending_supplies > 0:
        raise HTTPException(
            status_code=400,
            detail=f"There are {pending_supplies} supplies that have not been delivered. All supplies must be delivered before departure."
        )

    if db_service.waste_recovery:
        if db_service.waste_recovery.status != WasteStatus.COMPLETED.value:
            raise HTTPException(
                status_code=400,
                detail="Waste recovery must be completed before departure."
            )

    return {"status": "ready", "message": "All preconditions met for departure confirmation"}


def create_berth(db: Session, berth_in: BerthCreate) -> Berth:
    existing = db.query(Berth).filter(Berth.berth_number == berth_in.berth_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="Berth number already exists")

    db_berth = Berth(
        berth_number=berth_in.berth_number,
        capacity=berth_in.capacity,
        status=BerthStatus.AVAILABLE.value,
    )
    db.add(db_berth)
    db.commit()
    db.refresh(db_berth)
    return db_berth


def get_berths(db: Session, skip: int = 0, limit: int = 100) -> List[Berth]:
    return db.query(Berth).offset(skip).limit(limit).all()


def create_berth_application(
    db: Session,
    service_id: int,
    app_in: BerthApplicationCreate,
) -> BerthApplication:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    if db_service.current_stage != AgentServiceStage.DECLARATION.value:
        raise HTTPException(
            status_code=400,
            detail="Berth application can only be submitted during the declaration stage."
        )

    if app_in.requested_berthing_time <= datetime.now():
        raise HTTPException(
            status_code=400,
            detail="Requested berthing time must be in the future."
        )

    if app_in.expected_duration_hours < 2 or app_in.expected_duration_hours > 72:
        raise HTTPException(
            status_code=400,
            detail="Expected berthing duration must be between 2 and 72 hours."
        )

    if db_service.berth_application:
        raise HTTPException(
            status_code=400,
            detail="A berth application already exists for this service."
        )

    time_diff = app_in.requested_berthing_time - db_service.arrival_time
    is_waiting_timeout = time_diff.total_seconds() > 24 * 3600

    db_app = BerthApplication(
        agent_service_id=service_id,
        requested_berthing_time=app_in.requested_berthing_time,
        expected_duration_hours=app_in.expected_duration_hours,
        berth_preference=app_in.berth_preference,
        status=BerthApplicationStatus.PENDING.value,
        is_waiting_timeout=is_waiting_timeout,
    )
    db.add(db_app)
    db.commit()
    db.refresh(db_app)
    return db_app


def get_berth_application(db: Session, service_id: int) -> Optional[BerthApplication]:
    return db.query(BerthApplication).filter(BerthApplication.agent_service_id == service_id).first()


def approve_berth_application(
    db: Session,
    service_id: int,
    approved: bool,
    notes: Optional[str] = None,
) -> BerthApplication:
    db_app = get_berth_application(db, service_id)
    if not db_app:
        raise HTTPException(status_code=404, detail="Berth application not found")

    if db_app.status != BerthApplicationStatus.PENDING.value:
        raise HTTPException(
            status_code=400,
            detail="Application has already been processed."
        )

    if approved:
        available_berth = None
        if db_app.berth_preference:
            available_berth = db.query(Berth).filter(
                Berth.berth_number == db_app.berth_preference,
                Berth.status == BerthStatus.AVAILABLE.value,
            ).first()

        if not available_berth:
            available_berth = db.query(Berth).filter(
                Berth.status == BerthStatus.AVAILABLE.value,
            ).first()

        if not available_berth:
            raise HTTPException(
                status_code=400,
                detail="No available berths to assign. Please add more berths first."
            )

        db_app.assigned_berth_id = available_berth.id
        available_berth.status = BerthStatus.OCCUPIED.value
        db_app.assigned_berth_time = datetime.now()
        db_app.approved_at = datetime.now()
        db_app.status = BerthApplicationStatus.APPROVED.value
    else:
        db_app.status = BerthApplicationStatus.REJECTED.value

    db.commit()
    db.refresh(db_app)
    return db_app


def create_supply(db: Session, service_id: int, supply_in: SupplyCreate) -> Supply:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    existing_supplies = db.query(Supply).filter(
        Supply.agent_service_id == service_id,
        Supply.status != SupplyStatus.CANCELLED.value,
    ).all()

    if supply_in.supply_type == SupplyType.FUEL.value:
        has_fresh_water = any(s.supply_type == SupplyType.FRESH_WATER.value for s in existing_supplies)
        if has_fresh_water:
            raise HTTPException(
                status_code=400,
                detail="Fuel and fresh water cannot be delivered together. Please cancel fresh water supply first."
            )
    elif supply_in.supply_type == SupplyType.FRESH_WATER.value:
        has_fuel = any(s.supply_type == SupplyType.FUEL.value for s in existing_supplies)
        if has_fuel:
            raise HTTPException(
                status_code=400,
                detail="Fuel and fresh water cannot be delivered together. Please cancel fuel supply first."
            )

    total_price = supply_in.quantity * supply_in.unit_price

    db_supply = Supply(
        agent_service_id=service_id,
        supply_type=supply_in.supply_type,
        quantity=supply_in.quantity,
        unit_price=supply_in.unit_price,
        total_price=total_price,
        scheduled_time=supply_in.scheduled_time,
        status=SupplyStatus.SCHEDULED.value,
        notes=supply_in.notes,
    )
    db.add(db_supply)
    db.commit()
    db.refresh(db_supply)
    return db_supply


def get_supplies(db: Session, service_id: int) -> List[Supply]:
    return db.query(Supply).filter(Supply.agent_service_id == service_id).all()


def deliver_supply(db: Session, supply_id: int) -> Supply:
    db_supply = db.query(Supply).filter(Supply.id == supply_id).first()
    if not db_supply:
        raise HTTPException(status_code=404, detail="Supply not found")

    if db_supply.status != SupplyStatus.SCHEDULED.value:
        raise HTTPException(status_code=400, detail="Only scheduled supplies can be delivered")

    db_supply.status = SupplyStatus.DELIVERED.value
    db_supply.delivered_at = datetime.now()
    db.commit()
    db.refresh(db_supply)
    return db_supply


def cancel_supply(db: Session, supply_id: int) -> Supply:
    db_supply = db.query(Supply).filter(Supply.id == supply_id).first()
    if not db_supply:
        raise HTTPException(status_code=404, detail="Supply not found")

    if db_supply.status != SupplyStatus.SCHEDULED.value:
        raise HTTPException(status_code=400, detail="Only scheduled supplies can be cancelled")

    db_supply.status = SupplyStatus.CANCELLED.value
    db.commit()
    db.refresh(db_supply)
    return db_supply


def create_waste_recovery(db: Session, service_id: int, waste_in: WasteRecoveryCreate) -> WasteRecovery:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    if db_service.waste_recovery:
        raise HTTPException(
            status_code=400,
            detail="A waste recovery record already exists for this service."
        )

    base_fee = 0
    discounted_weight = 0
    discounted_unit_price = 0

    if waste_in.total_weight_kg <= 500:
        base_fee = waste_in.total_weight_kg * waste_in.unit_price_per_kg
    else:
        normal_weight = 500
        normal_fee = normal_weight * waste_in.unit_price_per_kg
        discounted_weight = waste_in.total_weight_kg - 500
        discounted_unit_price = waste_in.unit_price_per_kg * 0.8
        base_fee = normal_fee + (discounted_weight * discounted_unit_price)

    db_waste = WasteRecovery(
        agent_service_id=service_id,
        total_weight_kg=waste_in.total_weight_kg,
        unit_price_per_kg=waste_in.unit_price_per_kg,
        discounted_weight_kg=discounted_weight if discounted_weight > 0 else None,
        discounted_unit_price=discounted_unit_price if discounted_unit_price > 0 else None,
        total_fee=base_fee,
        scheduled_time=waste_in.scheduled_time,
        status=WasteStatus.SCHEDULED.value,
        notes=waste_in.notes,
    )
    db.add(db_waste)
    db.commit()
    db.refresh(db_waste)
    return db_waste


def get_waste_recovery(db: Session, service_id: int) -> Optional[WasteRecovery]:
    return db.query(WasteRecovery).filter(WasteRecovery.agent_service_id == service_id).first()


def complete_waste_recovery(db: Session, service_id: int) -> WasteRecovery:
    db_waste = get_waste_recovery(db, service_id)
    if not db_waste:
        raise HTTPException(status_code=404, detail="Waste recovery record not found")

    if db_waste.status != WasteStatus.SCHEDULED.value:
        raise HTTPException(status_code=400, detail="Only scheduled waste recovery can be completed")

    db_waste.status = WasteStatus.COMPLETED.value
    db_waste.completed_at = datetime.now()
    db.commit()
    db.refresh(db_waste)
    return db_waste


def create_todo(db: Session, service_id: int, todo_in: TodoCreate) -> Todo:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    existing_count = db.query(Todo).filter(Todo.agent_service_id == service_id).count()

    db_todo = Todo(
        agent_service_id=service_id,
        title=todo_in.title,
        description=todo_in.description,
        due_time=todo_in.due_time,
        status=TodoStatus.PENDING.value,
        sequence=existing_count + 1,
    )
    db.add(db_todo)
    db.commit()
    db.refresh(db_todo)
    return db_todo


def get_todos(db: Session, service_id: int) -> List[Todo]:
    return db.query(Todo).filter(
        Todo.agent_service_id == service_id
    ).order_by(Todo.due_time, Todo.id).all()


def complete_todo(db: Session, todo_id: int) -> Todo:
    db_todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not db_todo:
        raise HTTPException(status_code=404, detail="Todo not found")

    if db_todo.status == TodoStatus.COMPLETED.value:
        return db_todo

    pending_earlier = db.query(Todo).filter(
        Todo.agent_service_id == db_todo.agent_service_id,
        Todo.status == TodoStatus.PENDING.value,
        Todo.due_time < db_todo.due_time,
    ).count()

    if pending_earlier > 0:
        raise HTTPException(
            status_code=400,
            detail="Earlier todos must be completed first. Please complete pending todos with earlier due times."
        )

    db_todo.status = TodoStatus.COMPLETED.value
    db_todo.completed_at = datetime.now()
    db.commit()
    db.refresh(db_todo)
    return db_todo


def create_fee_settlement(db: Session, service_id: int, fee_in: FeeSettlementCreate) -> FeeSettlement:
    db_service = get_agent_service(db, service_id)
    if not db_service:
        raise HTTPException(status_code=404, detail="Agent service not found")

    if db_service.fee_settlement:
        raise HTTPException(
            status_code=400,
            detail="A fee settlement already exists for this service."
        )

    supplies_fee = db.query(Supply).filter(
        Supply.agent_service_id == service_id,
        Supply.status == SupplyStatus.DELIVERED.value,
    ).all()

    total_supplies_fee = sum(s.total_price for s in supplies_fee)

    waste_fee = 0
    if db_service.waste_recovery:
        waste_fee = db_service.waste_recovery.total_fee

    total_amount = fee_in.agent_fee + total_supplies_fee + waste_fee + fee_in.other_fees
    final_amount = total_amount

    db_fee = FeeSettlement(
        agent_service_id=service_id,
        agent_fee=fee_in.agent_fee,
        supplies_fee=total_supplies_fee,
        waste_recovery_fee=waste_fee,
        other_fees=fee_in.other_fees,
        total_amount=total_amount,
        final_amount=final_amount,
        status=SettlementStatus.PENDING.value,
        notes=fee_in.notes,
    )
    db.add(db_fee)
    db.commit()
    db.refresh(db_fee)
    return db_fee


def get_fee_settlement(db: Session, service_id: int) -> Optional[FeeSettlement]:
    return db.query(FeeSettlement).filter(FeeSettlement.agent_service_id == service_id).first()


def settle_fee(db: Session, service_id: int) -> FeeSettlement:
    db_fee = get_fee_settlement(db, service_id)
    if not db_fee:
        raise HTTPException(status_code=404, detail="Fee settlement not found")

    if db_fee.status == SettlementStatus.SETTLED.value:
        return db_fee

    supplies_fee = db.query(Supply).filter(
        Supply.agent_service_id == service_id,
        Supply.status == SupplyStatus.DELIVERED.value,
    ).all()

    total_supplies_fee = sum(s.total_price for s in supplies_fee)

    waste_fee = 0
    db_service = db_fee.agent_service
    if db_service and db_service.waste_recovery:
        waste_fee = db_service.waste_recovery.total_fee

    db_fee.supplies_fee = total_supplies_fee
    db_fee.waste_recovery_fee = waste_fee
    db_fee.total_amount = db_fee.agent_fee + total_supplies_fee + waste_fee + db_fee.other_fees
    db_fee.final_amount = db_fee.total_amount - total_supplies_fee - waste_fee
    db_fee.status = SettlementStatus.SETTLED.value
    db_fee.settled_at = datetime.now()

    db.commit()
    db.refresh(db_fee)
    return db_fee


def confirm_departure(db: Session, service_id: int, notes: Optional[str] = None) -> AgentService:
    validate_departure_preconditions(db, service_id)

    db_service = get_agent_service(db, service_id)

    if db_service.current_stage != AgentServiceStage.FEE_SETTLEMENT.value:
        raise HTTPException(
            status_code=400,
            detail="Departure can only be confirmed from fee settlement stage."
        )

    settle_fee(db, service_id)

    db_service.current_stage = AgentServiceStage.DEPARTURE_CONFIRMED.value
    db_service.departure_time = datetime.now()

    if db_service.berth_application and db_service.berth_application.assigned_berth:
        db_service.berth_application.assigned_berth.status = BerthStatus.AVAILABLE.value

    db_stage = db.query(StageHistory).filter(
        StageHistory.agent_service_id == service_id,
        StageHistory.stage == AgentServiceStage.DEPARTURE_CONFIRMED.value,
    ).first()

    if db_stage:
        db_stage.status = StageStatus.COMPLETED.value
        db_stage.completed_at = datetime.now()
        if notes:
            db_stage.notes = notes

    db.commit()
    db.refresh(db_service)
    return db_service
