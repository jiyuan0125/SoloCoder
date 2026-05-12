import logging
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from typing import List, Tuple

from .config import STORAGE_FREE_HOURS, STORAGE_RATE_PER_KG_PER_DAY
from .models import Cargo, Compartment, LoadAssignment, Dispatcher, AlertLog, Flight

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def calculate_chargeable_weight(actual_weight: float, volume_weight: float = 0) -> float:
    return max(actual_weight, volume_weight)


def calculate_freight_charge(chargeable_weight: float, price_per_kg: float) -> float:
    return round(chargeable_weight * price_per_kg, 2)


def calculate_storage_charge(arrival_time: datetime, chargeable_weight: float,
                              current_time: datetime = None) -> float:
    if current_time is None:
        current_time = datetime.utcnow()
    
    time_since_arrival = current_time - arrival_time
    free_period = timedelta(hours=STORAGE_FREE_HOURS)
    
    if time_since_arrival <= free_period:
        return 0.0
    
    excess_time = time_since_arrival - free_period
    days = (excess_time.total_seconds() // 86400) + 1
    
    return round(days * chargeable_weight * STORAGE_RATE_PER_KG_PER_DAY, 2)


def check_cargo_compartment_compatibility(cargo: Cargo, compartment: Compartment) -> Tuple[bool, str]:
    cargo_type = cargo.cargo_type.lower()
    
    if cargo_type == "cold_chain" or cargo_type == "fresh" or cargo_type == "生鲜" or cargo_type == "冷链":
        if not compartment.is_temperature_controlled:
            return False, "生鲜冷链货物只能装温控舱位"
    
    return True, ""


def check_compartment_mixed_cargo(
    db: Session,
    compartment: Compartment,
    new_cargo_type: str,
    exclude_assignment_id: int = None
) -> Tuple[bool, str]:
    assignments = db.query(LoadAssignment).filter(
        LoadAssignment.compartment_id == compartment.id
    ).all()
    
    all_cargo_types = []
    
    for assignment in assignments:
        if exclude_assignment_id and assignment.id == exclude_assignment_id:
            continue
        existing_cargo = db.query(Cargo).filter(Cargo.id == assignment.cargo_id).first()
        if existing_cargo:
            all_cargo_types.append(existing_cargo.cargo_type.lower())
    
    all_cargo_types.append(new_cargo_type.lower())
    
    has_dangerous = False
    has_live_animal = False
    
    for cargo_type in all_cargo_types:
        is_dangerous = cargo_type in ["dangerous", "hazardous", "危险品"]
        is_live = cargo_type in ["live_animal", "活体动物", "活物"]
        
        if is_dangerous:
            has_dangerous = True
        if is_live:
            has_live_animal = True
        
        if has_dangerous and has_live_animal:
            return False, "活体动物不能与危险品同舱"
    
    return True, ""


def check_pending_assignments_mixed_cargo(
    db: Session,
    compartment: Compartment,
    pending_assignments: List[LoadAssignment]
) -> Tuple[bool, str]:
    all_cargo_types = []
    
    for assignment in pending_assignments:
        cargo = db.query(Cargo).filter(Cargo.id == assignment.cargo_id).first()
        if cargo:
            all_cargo_types.append(cargo.cargo_type.lower())
    
    confirmed_assignments = db.query(LoadAssignment).filter(
        LoadAssignment.compartment_id == compartment.id,
        LoadAssignment.is_confirmed == True
    ).all()
    
    for assignment in confirmed_assignments:
        cargo = db.query(Cargo).filter(Cargo.id == assignment.cargo_id).first()
        if cargo:
            all_cargo_types.append(cargo.cargo_type.lower())
    
    has_dangerous = False
    has_live_animal = False
    
    for cargo_type in all_cargo_types:
        is_dangerous = cargo_type in ["dangerous", "hazardous", "危险品"]
        is_live = cargo_type in ["live_animal", "活体动物", "活物"]
        
        if is_dangerous:
            has_dangerous = True
        if is_live:
            has_live_animal = True
        
        if has_dangerous and has_live_animal:
            return False, "活体动物不能与危险品同舱"
    
    return True, ""


def check_front_rear_weight_balance(db: Session, flight_id: int) -> Tuple[bool, str, float]:
    compartments = db.query(Compartment).filter(Compartment.flight_id == flight_id).all()
    
    front_weight = 0.0
    rear_weight = 0.0
    
    for compartment in compartments:
        used_weight = compartment.total_capacity_weight - compartment.remaining_capacity_weight
        if compartment.position.lower() in ["front", "前舱", "前"]:
            front_weight += used_weight
        elif compartment.position.lower() in ["rear", "后舱", "后"]:
            rear_weight += used_weight
    
    total_weight = front_weight + rear_weight
    if total_weight == 0:
        return True, "", 0.0
    
    diff = abs(front_weight - rear_weight)
    diff_percent = diff / total_weight
    
    if diff_percent > 0.10:
        return False, f"前后舱重量差超过总载重的10%（当前：{diff_percent*100:.1f}%）", diff_percent
    
    return True, "", diff_percent


def check_min_load_rate(db: Session, flight: Flight) -> Tuple[bool, float, float]:
    compartments = db.query(Compartment).filter(Compartment.flight_id == flight.id).all()
    
    total_capacity = sum(c.total_capacity_weight for c in compartments)
    total_used = sum((c.total_capacity_weight - c.remaining_capacity_weight) for c in compartments)
    
    if total_capacity == 0:
        return True, 0.0, 0.0
    
    current_rate = total_used / total_capacity
    min_rate = flight.min_load_rate
    
    if current_rate < min_rate:
        return False, current_rate, min_rate
    
    return True, current_rate, min_rate


def has_active_dispatcher(db: Session) -> bool:
    return db.query(Dispatcher).filter(Dispatcher.is_active == True).first() is not None


def create_low_load_alert(db: Session, flight: Flight, current_rate: float, min_rate: float) -> AlertLog:
    has_dispatcher = has_active_dispatcher(db)
    message = f"航班 {flight.flight_number} ({flight.route}) 当前配载率 {current_rate*100:.1f}%，低于最低配载率 {min_rate*100:.1f}%"
    
    if has_dispatcher:
        logger.info(f"[调度提醒] {message}")
    else:
        logger.warning(f"[日志记录] 无调度员信息，仅记录日志: {message}")
    
    alert = AlertLog(
        flight_id=flight.id,
        message=message,
        alert_type="low_load_rate",
        sent_to_dispatcher=has_dispatcher
    )
    db.add(alert)
    db.commit()
    db.refresh(alert)
    
    return alert
