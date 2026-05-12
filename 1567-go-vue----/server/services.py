from datetime import datetime, timedelta, timezone
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import func

from . import models, schemas


STATUS_FLOW = [
    "CHECKED_IN",
    "SECURITY_CHECKED",
    "SORTED",
    "LOADED",
    "IN_TRANSIT",
    "ARRIVED",
    "ON_CONVEYOR",
    "PICKED_UP"
]

MAX_CLAIM_NORMAL = 1000.0
MAX_CLAIM_VALUABLE = 5000.0
SORTING_FAILURE_THRESHOLD = 5
ARRIVAL_TO_CONVEYOR_MINUTES = 30
CONVEYOR_STUCK_MINUTES = 60
LOST_CLAIM_WINDOW_HOURS = 72


def create_flight(db: Session, flight: schemas.FlightCreate) -> models.Flight:
    db_flight = models.Flight(
        id=flight.id,
        flight_number=flight.flight_number,
        departure=flight.departure,
        destination=flight.destination,
        scheduled_departure=flight.scheduled_departure,
        scheduled_arrival=flight.scheduled_arrival,
        max_load_weight=flight.max_load_weight
    )
    db.add(db_flight)
    db.commit()
    db.refresh(db_flight)
    return db_flight


def get_flight(db: Session, flight_id: str) -> Optional[models.Flight]:
    return db.query(models.Flight).filter(models.Flight.id == flight_id).first()


def get_all_flights(db: Session) -> List[models.Flight]:
    return db.query(models.Flight).all()


def update_flight(db: Session, flight_id: str, update: schemas.FlightUpdate) -> Optional[models.Flight]:
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    for key, value in update.model_dump(exclude_unset=True).items():
        setattr(flight, key, value)
    db.commit()
    db.refresh(flight)
    return flight


def create_baggage(db: Session, baggage: schemas.BaggageCreate) -> models.Baggage:
    db_baggage = models.Baggage(
        id=baggage.id,
        tag_number=baggage.tag_number,
        weight=baggage.weight,
        is_valuable=baggage.is_valuable,
        passenger_name=baggage.passenger_name,
        flight_id=baggage.flight_id,
        current_status="CHECKED_IN"
    )
    db.add(db_baggage)
    event = models.BaggageEvent(
        baggage_id=baggage.id,
        event_type="CHECKED_IN",
        notes="行李托运完成"
    )
    db.add(event)
    db.commit()
    db.refresh(db_baggage)
    return db_baggage


def get_baggage(db: Session, baggage_id: str) -> Optional[models.Baggage]:
    return db.query(models.Baggage).filter(models.Baggage.id == baggage_id).first()


def get_baggage_by_tag(db: Session, tag_number: str) -> Optional[models.Baggage]:
    return db.query(models.Baggage).filter(models.Baggage.tag_number == tag_number).first()


def update_baggage_flight(db: Session, baggage_id: str, flight_id: str) -> Optional[models.Baggage]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return None
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    baggage.flight_id = flight_id
    db.query(models.SortingRecord).filter(
        models.SortingRecord.baggage_id == baggage_id,
        models.SortingRecord.flight_id.is_(None)
    ).update({"flight_id": flight_id})
    db.commit()
    db.refresh(baggage)
    return baggage


def _check_status_transition(current: str, target: str) -> bool:
    if current == "EXCEPTION":
        return False
    try:
        current_idx = STATUS_FLOW.index(current)
        target_idx = STATUS_FLOW.index(target)
        return target_idx == current_idx + 1
    except ValueError:
        return False


def _create_event(db: Session, baggage_id: str, event_type: str, success: bool = True, notes: str = None) -> models.BaggageEvent:
    event = models.BaggageEvent(
        baggage_id=baggage_id,
        event_type=event_type,
        success=success,
        notes=notes
    )
    db.add(event)
    return event


def _create_alert(db: Session, alert_type: str, message: str, flight_id: str = None, baggage_id: str = None) -> models.Alert:
    alert = models.Alert(
        alert_type=alert_type,
        flight_id=flight_id,
        baggage_id=baggage_id,
        message=message
    )
    db.add(alert)
    return alert


def process_security_check(db: Session, baggage_id: str, request: schemas.SecurityCheckRequest) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if not _check_status_transition(baggage.current_status, "SECURITY_CHECKED"):
        return False, f"当前状态 {baggage.current_status} 不能进行安检"
    if request.success:
        baggage.current_status = "SECURITY_CHECKED"
        _create_event(db, baggage_id, "SECURITY_CHECKED", True, request.notes or "安检通过")
    else:
        baggage.current_status = "EXCEPTION"
        _create_event(db, baggage_id, "SECURITY_FAILED", False, request.notes or "安检未通过，进入异常处理")
        _create_alert(db, "SECURITY_FAILED", f"行李 {baggage.tag_number} 安检未通过: {request.notes or ''}", baggage_id=baggage_id)
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def process_sorting(db: Session, baggage_id: str, request: schemas.SortingRequest) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if not _check_status_transition(baggage.current_status, "SORTED"):
        return False, f"当前状态 {baggage.current_status} 不能进行分拣"
    record = models.SortingRecord(
        baggage_id=baggage_id,
        flight_id=baggage.flight_id,
        success=request.success,
        lane=request.lane,
        failure_reason=request.failure_reason
    )
    db.add(record)
    if request.success:
        baggage.current_status = "SORTED"
        _create_event(db, baggage_id, "SORTED", True, request.lane or f"分拣完成，通道: {request.lane}")
    else:
        _create_event(db, baggage_id, "SORTING_FAILED", False, request.failure_reason or "分拣失败")
        flight_id = baggage.flight_id
        if flight_id:
            failed_count = db.query(models.SortingRecord).filter(
                models.SortingRecord.flight_id == flight_id,
                models.SortingRecord.success == False
            ).count()
            if failed_count >= SORTING_FAILURE_THRESHOLD:
                existing_alert = db.query(models.Alert).filter(
                    models.Alert.alert_type == "SORTING_FAILURE_THRESHOLD",
                    models.Alert.flight_id == flight_id,
                    models.Alert.resolved == False
                ).first()
                if not existing_alert:
                    _create_alert(db, "SORTING_FAILURE_THRESHOLD", 
                                  f"航班分拣失败已达 {SORTING_FAILURE_THRESHOLD} 件", 
                                  flight_id=flight_id)
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def process_loading(db: Session, baggage_id: str, request: schemas.LoadingRequest) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if not baggage.flight_id:
        return False, "行李未绑定航班"
    flight = get_flight(db, baggage.flight_id)
    if not flight:
        return False, "航班不存在"
    if not _check_status_transition(baggage.current_status, "LOADED"):
        return False, f"当前状态 {baggage.current_status} 不能进行装载"
    new_weight = flight.actual_load_weight + baggage.weight
    if new_weight > flight.max_load_weight:
        _create_alert(db, "OVERLOAD_WARNING", 
                      f"航班装载重量即将超限: 当前 {flight.actual_load_weight}kg, 新增 {baggage.weight}kg, 限额 {flight.max_load_weight}kg",
                      flight_id=flight.id)
        return False, f"装载重量超限: 航班限额 {flight.max_load_weight}kg, 剩余可用 {flight.max_load_weight - flight.actual_load_weight}kg"
    is_special = baggage.is_valuable
    if is_special and not request.position.startswith("SPECIAL-"):
        return False, "贵重物品必须装载到专用位置（SPECIAL-前缀）"
    if not is_special and request.position.startswith("SPECIAL-"):
        return False, "非贵重物品不能装载到专用位置"
    record = models.LoadingRecord(
        baggage_id=baggage_id,
        flight_id=flight.id,
        position=request.position,
        is_special_position=is_special
    )
    db.add(record)
    flight.actual_load_weight = new_weight
    baggage.current_status = "LOADED"
    _create_event(db, baggage_id, "LOADED", True, f"装载位置: {request.position}")
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def process_in_transit(db: Session, baggage_id: str) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if not _check_status_transition(baggage.current_status, "IN_TRANSIT"):
        return False, f"当前状态 {baggage.current_status} 不能进入运输"
    baggage.current_status = "IN_TRANSIT"
    _create_event(db, baggage_id, "IN_TRANSIT", True, "行李在运输途中")
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def process_arrival(db: Session, baggage_id: str) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if not _check_status_transition(baggage.current_status, "ARRIVED"):
        return False, f"当前状态 {baggage.current_status} 不能标记到达"
    baggage.current_status = "ARRIVED"
    _create_event(db, baggage_id, "ARRIVED", True, "行李已到达目的地")
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def process_conveyor_belt(db: Session, baggage_id: str, request: schemas.ConveyorBeltRequest) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if baggage.current_status not in ["ARRIVED", "ON_CONVEYOR"]:
        return False, f"当前状态 {baggage.current_status} 不能上传送带"
    now = datetime.now(timezone.utc)
    arrival_event = db.query(models.BaggageEvent).filter(
        models.BaggageEvent.baggage_id == baggage_id,
        models.BaggageEvent.event_type == "ARRIVED"
    ).order_by(models.BaggageEvent.event_time.desc()).first()
    is_delayed = False
    if arrival_event:
        arrival_time = arrival_event.event_time
        if arrival_time.tzinfo is None:
            arrival_time = arrival_time.replace(tzinfo=timezone.utc)
        if (now - arrival_time) > timedelta(minutes=ARRIVAL_TO_CONVEYOR_MINUTES):
            is_delayed = True
            _create_alert(db, "CONVEYOR_DELAY", 
                          f"行李 {baggage.tag_number} 到达后超过 {ARRIVAL_TO_CONVEYOR_MINUTES} 分钟未上传送带",
                          baggage_id=baggage_id)
    existing = db.query(models.ConveyorBeltRecord).filter(
        models.ConveyorBeltRecord.baggage_id == baggage_id,
        models.ConveyorBeltRecord.picked_at.is_(None)
    ).first()
    if existing:
        return False, "行李已在传送带上"
    record = models.ConveyorBeltRecord(
        baggage_id=baggage_id,
        conveyor_belt_number=request.conveyor_belt_number,
        is_delayed=is_delayed
    )
    db.add(record)
    baggage.current_status = "ON_CONVEYOR"
    _create_event(db, baggage_id, "ON_CONVEYOR", True, f"传送带: {request.conveyor_belt_number}")
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def check_conveyor_stuck(db: Session):
    now = datetime.now(timezone.utc)
    threshold_time = now - timedelta(minutes=CONVEYOR_STUCK_MINUTES)
    records = db.query(models.ConveyorBeltRecord).filter(
        models.ConveyorBeltRecord.picked_at.is_(None),
        models.ConveyorBeltRecord.is_stuck == False,
        models.ConveyorBeltRecord.placed_at <= threshold_time
    ).all()
    for record in records:
        record.is_stuck = True
        baggage = get_baggage(db, record.baggage_id)
        if baggage:
            _create_alert(db, "CONVEYOR_STUCK", 
                          f"行李 {baggage.tag_number} 在传送带上停留超过 {CONVEYOR_STUCK_MINUTES} 分钟",
                          baggage_id=baggage.id)
    db.commit()


def process_pickup(db: Session, baggage_id: str) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if not _check_status_transition(baggage.current_status, "PICKED_UP"):
        return False, f"当前状态 {baggage.current_status} 不能提取"
    now = datetime.now(timezone.utc)
    record = db.query(models.ConveyorBeltRecord).filter(
        models.ConveyorBeltRecord.baggage_id == baggage_id,
        models.ConveyorBeltRecord.picked_at.is_(None)
    ).first()
    if record:
        record.picked_at = now
    baggage.current_status = "PICKED_UP"
    _create_event(db, baggage_id, "PICKED_UP", True, "行李已提取")
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def report_lost(db: Session, baggage_id: str) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if baggage.current_status == "PICKED_UP":
        return False, "已提取的行李不能报失"
    baggage.current_status = "LOST"
    _create_event(db, baggage_id, "LOST", False, "行李报失")
    _create_alert(db, "BAGGAGE_LOST", f"行李 {baggage.tag_number} 丢失", baggage_id=baggage_id)
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def report_found(db: Session, baggage_id: str) -> Tuple[bool, str]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在"
    if baggage.current_status != "LOST":
        return False, "行李当前不是丢失状态"
    now = datetime.now(timezone.utc)
    lost_event = db.query(models.BaggageEvent).filter(
        models.BaggageEvent.baggage_id == baggage_id,
        models.BaggageEvent.event_type == "LOST"
    ).order_by(models.BaggageEvent.event_time.desc()).first()
    within_window = False
    if lost_event:
        lost_time = lost_event.event_time
        if lost_time.tzinfo is None:
            lost_time = lost_time.replace(tzinfo=timezone.utc)
        within_window = (now - lost_time) <= timedelta(hours=LOST_CLAIM_WINDOW_HOURS)
    if within_window:
        pending_claims = db.query(models.Claim).filter(
            models.Claim.baggage_id == baggage_id,
            models.Claim.status.in_(["PENDING", "APPROVED"])
        ).all()
        for claim in pending_claims:
            claim.status = "CANCELLED"
            claim.resolved_at = now
            claim.notes = (claim.notes or "") + " | 72小时内找回，自动取消赔偿"
        alerts = db.query(models.Alert).filter(
            models.Alert.baggage_id == baggage_id,
            models.Alert.resolved == False
        ).all()
        for alert in alerts:
            alert.resolved = True
            alert.resolved_at = now
    baggage.current_status = "FOUND"
    _create_event(db, baggage_id, "FOUND", True, "行李找回" + ("（72小时内，已取消赔偿）" if within_window else ""))
    db.commit()
    db.refresh(baggage)
    return True, "处理成功"


def file_claim(db: Session, baggage_id: str, claim: schemas.ClaimCreate) -> Tuple[bool, str, Optional[models.Claim]]:
    baggage = get_baggage(db, baggage_id)
    if not baggage:
        return False, "行李不存在", None
    max_allowed = MAX_CLAIM_VALUABLE if baggage.is_valuable else MAX_CLAIM_NORMAL
    if claim.claim_amount > max_allowed:
        return False, f"赔偿金额超出上限: 最高 {max_allowed} 元", None
    if baggage.current_status not in ["LOST", "EXCEPTION"]:
        return False, "只有丢失或异常行李才能申请赔偿", None
    db_claim = models.Claim(
        baggage_id=baggage_id,
        claim_amount=claim.claim_amount,
        max_allowed=max_allowed,
        status="PENDING",
        notes=claim.notes
    )
    db.add(db_claim)
    _create_event(db, baggage_id, "CLAIM_FILED", True, f"申请赔偿 {claim.claim_amount} 元")
    db.commit()
    db.refresh(db_claim)
    return True, "申请成功", db_claim


def get_flight_baggage_list(db: Session, flight_id: str) -> Optional[schemas.FlightBaggageListResponse]:
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    baggage_list = db.query(models.Baggage).filter(models.Baggage.flight_id == flight_id).all()
    return schemas.FlightBaggageListResponse(
        flight_id=flight_id,
        total_count=len(baggage_list),
        baggage=[schemas.BaggageResponse.model_validate(b) for b in baggage_list]
    )


def get_flight_sorting(db: Session, flight_id: str) -> Optional[schemas.FlightSortingResponse]:
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    records = db.query(models.SortingRecord).filter(models.SortingRecord.flight_id == flight_id).all()
    success_count = sum(1 for r in records if r.success)
    failure_count = len(records) - success_count
    return schemas.FlightSortingResponse(
        flight_id=flight_id,
        total_count=len(records),
        success_count=success_count,
        failure_count=failure_count,
        records=[schemas.SortingRecordResponse.model_validate(r) for r in records]
    )


def get_flight_loading(db: Session, flight_id: str) -> Optional[schemas.FlightLoadingResponse]:
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    records = db.query(models.LoadingRecord).filter(models.LoadingRecord.flight_id == flight_id).all()
    special_count = sum(1 for r in records if r.is_special_position)
    return schemas.FlightLoadingResponse(
        flight_id=flight_id,
        total_count=len(records),
        total_weight=flight.actual_load_weight,
        max_load_weight=flight.max_load_weight,
        special_position_count=special_count,
        records=[schemas.LoadingRecordResponse.model_validate(r) for r in records]
    )
