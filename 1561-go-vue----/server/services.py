from datetime import datetime, timedelta
from typing import Optional, List
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_
from .models import (
    Flight, FlightStatus, AircraftType,
    Stand, Gate, GateAdjacency,
    Vehicle, VehicleStatus, VehicleType, VehicleDispatch, VehicleMaintenance
)
from .schemas import FlightCreate, FlightUpdate, StandCreate, StandUpdate, GateCreate, GateAdjacencyCreate, VehicleCreate, VehicleUpdate, VehicleDispatchCreate, VehicleMaintenanceCreate, VehicleMaintenanceUpdate


def get_ground_time_minutes(aircraft_type: AircraftType) -> int:
    return 60 if aircraft_type == AircraftType.WIDE_BODY else 40


def update_flight_status_based_on_time(db: Session, flight: Flight, now: datetime):
    if flight.status == FlightStatus.COMPLETED:
        return
    
    if flight.status == FlightStatus.SCHEDULED and flight.check_in_start and now >= flight.check_in_start:
        flight.status = FlightStatus.CHECK_IN
    
    if flight.status == FlightStatus.CHECK_IN and flight.boarding_start and now >= flight.boarding_start:
        can_board, reason = can_flight_board_at_gate(db, flight)
        if can_board:
            flight.status = FlightStatus.BOARDING
    
    if flight.status == FlightStatus.BOARDING:
        if flight.actual_departure and now >= flight.actual_departure:
            flight.status = FlightStatus.DEPARTED
            release_stand_for_flight(flight)
        elif not flight.actual_departure and now >= flight.scheduled_departure + timedelta(minutes=10):
            flight.status = FlightStatus.DEPARTED
            release_stand_for_flight(flight)
    
    if flight.status == FlightStatus.DEPARTED:
        dep_time = flight.actual_departure or flight.scheduled_departure
        if now >= dep_time + timedelta(minutes=30):
            flight.status = FlightStatus.COMPLETED


def release_stand_for_flight(flight: Flight):
    if flight.stand:
        flight.stand.is_occupied = False
        flight.stand_id = None


def auto_update_all_flight_statuses(db: Session):
    now = datetime.utcnow()
    flights = db.query(Flight).filter(Flight.status != FlightStatus.COMPLETED).all()
    for flight in flights:
        update_flight_status_based_on_time(db, flight, now)
    db.commit()


def get_flight(db: Session, flight_id: int):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if flight:
        update_flight_status_based_on_time(db, flight, datetime.utcnow())
        db.commit()
    return flight


def get_flight_by_number(db: Session, flight_number: str):
    flight = db.query(Flight).filter(Flight.flight_number == flight_number).first()
    if flight:
        update_flight_status_based_on_time(db, flight, datetime.utcnow())
        db.commit()
    return flight


def get_flights(db: Session, skip: int = 0, limit: int = 100):
    auto_update_all_flight_statuses(db)
    return db.query(Flight).offset(skip).limit(limit).all()


def create_flight(db: Session, flight: FlightCreate):
    db_flight = Flight(**flight.model_dump())
    db.add(db_flight)
    db.commit()
    db.refresh(db_flight)
    return db_flight


def update_flight(db: Session, flight_id: int, flight_update: FlightUpdate):
    db_flight = get_flight(db, flight_id)
    if not db_flight:
        return None
    
    update_data = flight_update.model_dump(exclude_unset=True)
    
    if "status" in update_data and update_data["status"] == FlightStatus.BOARDING:
        can_board, reason = can_flight_board_at_gate(db, db_flight)
        if not can_board:
            raise ValueError(reason)
    
    for key, value in update_data.items():
        setattr(db_flight, key, value)
    
    update_flight_status_based_on_time(db, db_flight, datetime.utcnow())
    db.commit()
    db.refresh(db_flight)
    return db_flight


def delete_flight(db: Session, flight_id: int):
    db_flight = get_flight(db, flight_id)
    if not db_flight:
        return False
    db.delete(db_flight)
    db.commit()
    return True


def get_stand(db: Session, stand_id: int):
    return db.query(Stand).filter(Stand.id == stand_id).first()


def get_stands(db: Session, skip: int = 0, limit: int = 100):
    return db.query(Stand).offset(skip).limit(limit).all()


def create_stand(db: Session, stand: StandCreate):
    db_stand = Stand(**stand.model_dump())
    db.add(db_stand)
    db.commit()
    db.refresh(db_stand)
    return db_stand


def validate_flights_on_stand_after_change(db: Session, stand: Stand):
    assigned_flights = db.query(Flight).filter(Flight.stand_id == stand.id).all()
    for flight in assigned_flights:
        if flight.aircraft_type == AircraftType.WIDE_BODY and not stand.supports_wide_body:
            release_stand_for_flight(flight)
        elif flight.status in [FlightStatus.BOARDING, FlightStatus.CHECK_IN] and not stand.is_bridge:
            release_stand_for_flight(flight)


def update_stand(db: Session, stand_id: int, stand_update: StandUpdate):
    db_stand = get_stand(db, stand_id)
    if not db_stand:
        return None
    
    update_data = stand_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_stand, key, value)
    
    if "is_bridge" in update_data or "supports_wide_body" in update_data:
        validate_flights_on_stand_after_change(db, db_stand)
    
    db.commit()
    db.refresh(db_stand)
    return db_stand


def assign_stand_to_flight(db: Session, flight_id: int):
    flight = get_flight(db, flight_id)
    if not flight:
        return None, "Flight not found"
    
    if flight.status in [FlightStatus.DEPARTED, FlightStatus.COMPLETED]:
        return None, "Flight already departed or completed"
    
    if flight.stand_id:
        return flight.stand, "Flight already has a stand"
    
    candidate_stands = db.query(Stand).filter(Stand.is_occupied == False)
    
    if flight.aircraft_type == AircraftType.WIDE_BODY:
        candidate_stands = candidate_stands.filter(Stand.supports_wide_body == True)
    
    if flight.status in [FlightStatus.BOARDING, FlightStatus.CHECK_IN]:
        candidate_stands = candidate_stands.filter(Stand.is_bridge == True)
    
    stands_list = candidate_stands.order_by(Stand.is_bridge.desc(), Stand.id.asc()).all()
    
    if not stands_list:
        return None, "No available stands matching requirements"
    
    selected_stand = stands_list[0]
    selected_stand.is_occupied = True
    flight.stand_id = selected_stand.id
    
    db.commit()
    db.refresh(selected_stand)
    return selected_stand, None


def get_gate(db: Session, gate_id: int):
    return db.query(Gate).filter(Gate.id == gate_id).first()


def get_gates(db: Session, skip: int = 0, limit: int = 100):
    return db.query(Gate).offset(skip).limit(limit).all()


def create_gate(db: Session, gate: GateCreate):
    db_gate = Gate(**gate.model_dump())
    db.add(db_gate)
    db.commit()
    db.refresh(db_gate)
    return db_gate


def create_gate_adjacency(db: Session, adjacency: GateAdjacencyCreate):
    exists = db.query(GateAdjacency).filter(
        or_(
            and_(GateAdjacency.gate_id_1 == adjacency.gate_id_1, GateAdjacency.gate_id_2 == adjacency.gate_id_2),
            and_(GateAdjacency.gate_id_1 == adjacency.gate_id_2, GateAdjacency.gate_id_2 == adjacency.gate_id_1)
        )
    ).first()
    if exists:
        return exists
    db_adj = GateAdjacency(**adjacency.model_dump())
    db.add(db_adj)
    db.commit()
    db.refresh(db_adj)
    return db_adj


def get_adjacent_gate_ids(db: Session, gate_id: int) -> List[int]:
    adjacencies = db.query(GateAdjacency).filter(
        or_(GateAdjacency.gate_id_1 == gate_id, GateAdjacency.gate_id_2 == gate_id)
    ).all()
    adjacent = []
    for adj in adjacencies:
        if adj.gate_id_1 == gate_id:
            adjacent.append(adj.gate_id_2)
        else:
            adjacent.append(adj.gate_id_1)
    return adjacent


def can_flight_board_at_gate(db: Session, flight: Flight) -> tuple[bool, str]:
    if not flight.gate_id:
        return True, ""
    
    boarding_flights = db.query(Flight).filter(
        Flight.status == FlightStatus.BOARDING,
        Flight.id != flight.id
    ).all()
    
    used_gate_ids = [f.gate_id for f in boarding_flights if f.gate_id]
    adjacent_to_used = set()
    for gid in used_gate_ids:
        adjacent_to_used.update(get_adjacent_gate_ids(db, gid))
    
    if flight.gate_id in used_gate_ids:
        return False, "Gate is already in use by another boarding flight"
    if flight.gate_id in adjacent_to_used:
        return False, "Adjacent gate is already in use by another boarding flight"
    
    return True, ""


def assign_gate_to_flight(db: Session, flight_id: int):
    flight = get_flight(db, flight_id)
    if not flight:
        return None, "Flight not found"
    
    if flight.status in [FlightStatus.DEPARTED, FlightStatus.COMPLETED]:
        return None, "Flight already departed or completed"
    
    if flight.gate_id:
        return flight.gate, "Flight already has a gate"
    
    available_gates = db.query(Gate).filter(Gate.is_available == True).all()
    if not available_gates:
        return None, "No available gates"
    
    boarding_flights = db.query(Flight).filter(Flight.status == FlightStatus.BOARDING).all()
    used_gate_ids = [f.gate_id for f in boarding_flights if f.gate_id]
    adjacent_to_used = set()
    for gid in used_gate_ids:
        adjacent_to_used.update(get_adjacent_gate_ids(db, gid))
    
    for gate in available_gates:
        if gate.id in used_gate_ids:
            continue
        if gate.id in adjacent_to_used:
            continue
        selected_gate = gate
        break
    else:
        return None, "No gates available (adjacency constraint)"
    
    selected_gate.is_available = False
    flight.gate_id = selected_gate.id
    
    db.commit()
    db.refresh(selected_gate)
    return selected_gate, None


def get_vehicle(db: Session, vehicle_id: int):
    return db.query(Vehicle).filter(Vehicle.id == vehicle_id).first()


def get_vehicles(db: Session, skip: int = 0, limit: int = 100):
    return db.query(Vehicle).offset(skip).limit(limit).all()


def create_vehicle(db: Session, vehicle: VehicleCreate):
    db_vehicle = Vehicle(**vehicle.model_dump())
    db.add(db_vehicle)
    db.commit()
    db.refresh(db_vehicle)
    return db_vehicle


def update_vehicle(db: Session, vehicle_id: int, vehicle_update: VehicleUpdate):
    db_vehicle = get_vehicle(db, vehicle_id)
    if not db_vehicle:
        return None
    
    update_data = vehicle_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_vehicle, key, value)
    
    db.commit()
    db.refresh(db_vehicle)
    return db_vehicle


def get_vehicle_dispatches(db: Session, vehicle_id: int, skip: int = 0, limit: int = 100):
    return db.query(VehicleDispatch).filter(VehicleDispatch.vehicle_id == vehicle_id).offset(skip).limit(limit).all()


def get_vehicle_maintenances(db: Session, vehicle_id: int, skip: int = 0, limit: int = 100):
    return db.query(VehicleMaintenance).filter(VehicleMaintenance.vehicle_id == vehicle_id).offset(skip).limit(limit).all()


def create_vehicle_maintenance(db: Session, vehicle_id: int, maintenance: VehicleMaintenanceCreate):
    db_vehicle = get_vehicle(db, vehicle_id)
    if not db_vehicle:
        return None
    data = maintenance.model_dump()
    data["vehicle_id"] = vehicle_id
    db_maint = VehicleMaintenance(**data)
    db.add(db_maint)
    db.commit()
    db.refresh(db_maint)
    return db_maint


def update_vehicle_maintenance(db: Session, maintenance_id: int, maintenance_update: VehicleMaintenanceUpdate):
    db_maint = db.query(VehicleMaintenance).filter(VehicleMaintenance.id == maintenance_id).first()
    if not db_maint:
        return None
    
    update_data = maintenance_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_maint, key, value)
    
    db.commit()
    db.refresh(db_maint)
    return db_maint


def dispatch_vehicle_for_flight(db: Session, dispatch_req: VehicleDispatchCreate):
    flight = get_flight(db, dispatch_req.flight_id)
    if not flight:
        return None, "Flight not found"
    
    if flight.status in [FlightStatus.DEPARTED, FlightStatus.COMPLETED]:
        return None, "Flight already departed or completed"
    
    pending_waits = db.query(VehicleDispatch).filter(
        VehicleDispatch.vehicle_type == dispatch_req.vehicle_type,
        VehicleDispatch.is_waiting == True
    ).all()
    
    available_vehicles = db.query(Vehicle).filter(
        Vehicle.vehicle_type == dispatch_req.vehicle_type,
        Vehicle.status == VehicleStatus.AVAILABLE
    ).order_by(Vehicle.id.asc()).all()
    
    all_waiting = list(pending_waits)
    all_waiting.append(VehicleDispatch(
        flight_id=dispatch_req.flight_id,
        priority=dispatch_req.priority,
        is_waiting=True,
        vehicle_type=dispatch_req.vehicle_type
    ))
    
    flight_arrival_map = {}
    for w in all_waiting:
        f = db.query(Flight).filter(Flight.id == w.flight_id).first()
        flight_arrival_map[w.flight_id] = f.scheduled_arrival or f.scheduled_departure
    
    sorted_waiting = sorted(
        all_waiting,
        key=lambda x: (-x.priority, flight_arrival_map.get(x.flight_id, datetime.max))
    )
    
    assigned = []
    vehicle_idx = 0
    for waiting in sorted_waiting:
        if vehicle_idx < len(available_vehicles):
            vehicle = available_vehicles[vehicle_idx]
            vehicle.status = VehicleStatus.IN_USE
            waiting.vehicle_id = vehicle.id
            waiting.is_waiting = False
            waiting.started_at = datetime.utcnow()
            assigned.append(waiting)
            vehicle_idx += 1
        else:
            waiting.is_waiting = True
    
    for w in sorted_waiting:
        if w.id is None:
            db.add(w)
    
    db.commit()
    
    latest = db.query(VehicleDispatch).filter(
        VehicleDispatch.flight_id == dispatch_req.flight_id,
        VehicleDispatch.vehicle_type == dispatch_req.vehicle_type
    ).order_by(VehicleDispatch.id.desc()).first()
    
    return latest, None
