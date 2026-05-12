from sqlalchemy.orm import Session
from sqlalchemy import and_, or_
from datetime import datetime, timedelta
from typing import List, Optional
from models import (
    Terminal, Berth, Ship, Flight, Ticket, Refund, Weather,
    FlightStatus, BerthStatus, ShipStatus, TicketStatus
)
from schemas import (
    TerminalCreate, TerminalUpdate, BerthCreate, BerthUpdate,
    ShipCreate, ShipUpdate, FlightCreate, FlightUpdate,
    TicketCreate, WeatherCreate
)


MIN_BERTH_INTERVAL_MINUTES = 30
MAX_WIND_SPEED = 6
MIN_VISIBILITY = 1000
DELAY_NOTIFICATION_THRESHOLD = 30


def get_terminals(db: Session, skip: int = 0, limit: int = 100):
    return db.query(Terminal).offset(skip).limit(limit).all()


def get_terminal(db: Session, terminal_id: int):
    return db.query(Terminal).filter(Terminal.id == terminal_id).first()


def create_terminal(db: Session, terminal: TerminalCreate):
    db_terminal = Terminal(**terminal.model_dump())
    db.add(db_terminal)
    db.commit()
    db.refresh(db_terminal)
    return db_terminal


def update_terminal(db: Session, terminal_id: int, terminal: TerminalUpdate):
    db_terminal = get_terminal(db, terminal_id)
    if db_terminal:
        update_data = terminal.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(db_terminal, key, value)
        db.commit()
        db.refresh(db_terminal)
    return db_terminal


def delete_terminal(db: Session, terminal_id: int):
    db_terminal = get_terminal(db, terminal_id)
    if db_terminal:
        db.delete(db_terminal)
        db.commit()
    return db_terminal


def get_berths(db: Session, skip: int = 0, limit: int = 100, terminal_id: Optional[int] = None):
    query = db.query(Berth)
    if terminal_id:
        query = query.filter(Berth.terminal_id == terminal_id)
    return query.offset(skip).limit(limit).all()


def get_berth(db: Session, berth_id: int):
    return db.query(Berth).filter(Berth.id == berth_id).first()


def create_berth(db: Session, berth: BerthCreate):
    db_berth = Berth(**berth.model_dump())
    db.add(db_berth)
    db.commit()
    db.refresh(db_berth)
    return db_berth


def update_berth(db: Session, berth_id: int, berth: BerthUpdate):
    db_berth = get_berth(db, berth_id)
    if db_berth:
        update_data = berth.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(db_berth, key, value)
        db.commit()
        db.refresh(db_berth)
    return db_berth


def delete_berth(db: Session, berth_id: int):
    db_berth = get_berth(db, berth_id)
    if db_berth:
        db.delete(db_berth)
        db.commit()
    return db_berth


def check_berth_conflict(db: Session, berth_id: int, scheduled_departure: datetime, exclude_flight_id: Optional[int] = None):
    berth = get_berth(db, berth_id)
    if not berth:
        return False
    
    window_start = scheduled_departure - timedelta(minutes=MIN_BERTH_INTERVAL_MINUTES)
    window_end = scheduled_departure + timedelta(minutes=MIN_BERTH_INTERVAL_MINUTES)
    
    query = db.query(Flight).filter(
        Flight.berth_id == berth_id,
        Flight.status.in_([FlightStatus.PLANNED, FlightStatus.DELAYED, FlightStatus.DEPARTED]),
        or_(
            and_(Flight.scheduled_departure >= window_start, Flight.scheduled_departure <= window_end),
            and_(Flight.scheduled_arrival >= window_start, Flight.scheduled_arrival <= window_end)
        )
    )
    
    if exclude_flight_id:
        query = query.filter(Flight.id != exclude_flight_id)
    
    conflicting_flights = query.all()
    return len(conflicting_flights) > 0


def get_ships(db: Session, skip: int = 0, limit: int = 100, status: Optional[ShipStatus] = None):
    query = db.query(Ship)
    if status:
        query = query.filter(Ship.status == status)
    return query.offset(skip).limit(limit).all()


def get_ship(db: Session, ship_id: int):
    return db.query(Ship).filter(Ship.id == ship_id).first()


def create_ship(db: Session, ship: ShipCreate):
    db_ship = Ship(**ship.model_dump())
    db.add(db_ship)
    db.commit()
    db.refresh(db_ship)
    return db_ship


def update_ship(db: Session, ship_id: int, ship: ShipUpdate):
    db_ship = get_ship(db, ship_id)
    if db_ship:
        update_data = ship.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(db_ship, key, value)
        db.commit()
        db.refresh(db_ship)
    return db_ship


def delete_ship(db: Session, ship_id: int):
    db_ship = get_ship(db, ship_id)
    if db_ship:
        db.delete(db_ship)
        db.commit()
    return db_ship


def is_ship_available(db: Session, ship_id: int, scheduled_time: datetime):
    ship = get_ship(db, ship_id)
    if not ship:
        return False
    
    if ship.status == ShipStatus.IN_MAINTENANCE:
        return False
    
    if ship.inspection_start_date and ship.inspection_end_date:
        if ship.inspection_start_date <= scheduled_time <= ship.inspection_end_date:
            return False
    
    return True


def get_available_ships(db: Session, scheduled_time: datetime):
    ships = db.query(Ship).filter(Ship.status != ShipStatus.IN_MAINTENANCE).all()
    available = []
    for ship in ships:
        if ship.inspection_start_date and ship.inspection_end_date:
            if ship.inspection_start_date <= scheduled_time <= ship.inspection_end_date:
                continue
        available.append(ship)
    return available


def dispatch_ship(db: Session, estimated_passengers: int, scheduled_time: datetime):
    available_ships = get_available_ships(db, scheduled_time)
    if not available_ships:
        return None
    
    scored_ships = []
    for ship in available_ships:
        capacity_ratio = estimated_passengers / ship.passenger_capacity if ship.passenger_capacity > 0 else 0
        if capacity_ratio <= 1.0:
            scored_ships.append((ship, capacity_ratio))
    
    scored_ships.sort(key=lambda x: (abs(x[1] - 1.0), -x[0].passenger_capacity))
    
    if scored_ships:
        return scored_ships[0][0]
    
    return None


def get_flights(db: Session, skip: int = 0, limit: int = 100, 
                status: Optional[FlightStatus] = None,
                departure_terminal_id: Optional[int] = None):
    query = db.query(Flight)
    if status:
        query = query.filter(Flight.status == status)
    if departure_terminal_id:
        query = query.filter(Flight.departure_terminal_id == departure_terminal_id)
    return query.offset(skip).limit(limit).all()


def get_flight(db: Session, flight_id: int):
    return db.query(Flight).filter(Flight.id == flight_id).first()


def get_flight_by_number(db: Session, flight_number: str):
    return db.query(Flight).filter(Flight.flight_number == flight_number).first()


def create_flight(db: Session, flight: FlightCreate):
    flight_data = flight.model_dump()
    
    if flight_data.get('ship_id'):
        if not is_ship_available(db, flight_data['ship_id'], flight_data['scheduled_departure']):
            raise ValueError("该船舶在计划时间不可用")
    else:
        dispatched_ship = dispatch_ship(db, flight_data['estimated_passengers'], flight_data['scheduled_departure'])
        if dispatched_ship:
            flight_data['ship_id'] = dispatched_ship.id
    
    if flight_data.get('berth_id'):
        if check_berth_conflict(db, flight_data['berth_id'], flight_data['scheduled_departure']):
            raise ValueError("该泊位在计划时间有冲突")
    
    db_flight = Flight(**flight_data)
    db.add(db_flight)
    db.commit()
    db.refresh(db_flight)
    return db_flight


def check_weather_conditions(db: Session, terminal_id: int):
    latest_weather = db.query(Weather).filter(
        Weather.terminal_id == terminal_id
    ).order_by(Weather.timestamp.desc()).first()
    
    if not latest_weather:
        return True
    
    if latest_weather.wind_speed > MAX_WIND_SPEED:
        return False
    
    if latest_weather.visibility < MIN_VISIBILITY:
        return False
    
    return True


def check_load_status(ship: Ship, passengers: int, vehicles: int):
    passenger_ratio = passengers / ship.passenger_capacity if ship.passenger_capacity > 0 else 0
    vehicle_ratio = vehicles / ship.vehicle_capacity if ship.vehicle_capacity > 0 else 0
    
    if passenger_ratio > 1.0 or vehicle_ratio > 1.0:
        return "overloaded"
    elif passenger_ratio >= 0.9 or vehicle_ratio >= 0.9:
        return "near_full"
    elif passenger_ratio >= 1.0 or vehicle_ratio >= 1.0:
        return "full"
    return "normal"


def can_depart(db: Session, flight: Flight):
    if not flight.ship_id:
        return False, "未分配船舶"
    
    ship = get_ship(db, flight.ship_id)
    if not ship:
        return False, "船舶不存在"
    
    if ship.status == ShipStatus.IN_MAINTENANCE:
        return False, "船舶处于年度检验中"
    
    if not flight.berth_id:
        return False, "未分配泊位"
    
    berth = get_berth(db, flight.berth_id)
    if not berth or berth.status != BerthStatus.AVAILABLE:
        return False, "泊位不可用"
    
    load_status = check_load_status(ship, flight.actual_passengers, flight.actual_vehicles)
    if load_status == "overloaded":
        return False, "超载，不允许开航"
    
    if not check_weather_conditions(db, flight.departure_terminal_id):
        return False, "天气条件不允许开航（风力超6级或能见度低于1000米）"
    
    return True, f"可以开航（{load_status}）"


def update_flight_departure(db: Session, flight_id: int):
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    
    can_go, message = can_depart(db, flight)
    if not can_go:
        raise ValueError(message)
    
    flight.status = FlightStatus.DEPARTED
    flight.actual_departure = datetime.utcnow()
    
    if flight.berth_id:
        berth = get_berth(db, flight.berth_id)
        if berth:
            berth.status = BerthStatus.OCCUPIED
    
    db.commit()
    db.refresh(flight)
    return flight


def update_flight_arrival(db: Session, flight_id: int):
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    
    flight.status = FlightStatus.ARRIVED
    flight.actual_arrival = datetime.utcnow()
    
    if flight.berth_id:
        berth = get_berth(db, flight.berth_id)
        if berth:
            berth.status = BerthStatus.AVAILABLE
    
    db.commit()
    db.refresh(flight)
    return flight


def cancel_flight(db: Session, flight_id: int, reason: str = "航班取消"):
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    
    flight.status = FlightStatus.CANCELLED
    
    if flight.berth_id:
        berth = get_berth(db, flight.berth_id)
        if berth and berth.status == BerthStatus.OCCUPIED:
            berth.status = BerthStatus.AVAILABLE
    
    tickets = db.query(Ticket).filter(
        Ticket.flight_id == flight_id,
        Ticket.status == TicketStatus.SOLD,
        Ticket.is_on_site == False
    ).all()
    
    for ticket in tickets:
        ticket.status = TicketStatus.REFUNDED
        refund = Refund(
            ticket_id=ticket.id,
            refund_amount=ticket.amount,
            reason=reason,
            fee_waived=True
        )
        db.add(refund)
    
    db.commit()
    db.refresh(flight)
    return flight


def update_flight_delay(db: Session, flight_id: int, delay_minutes: int):
    flight = get_flight(db, flight_id)
    if not flight:
        return None
    
    flight.delay_minutes = delay_minutes
    if delay_minutes > 0:
        flight.status = FlightStatus.DELAYED
    
    db.commit()
    db.refresh(flight)
    return flight


def check_delay_notification(db: Session, flight_id: int):
    flight = get_flight(db, flight_id)
    if not flight:
        return False
    
    if flight.delay_minutes >= DELAY_NOTIFICATION_THRESHOLD and not flight.delay_notification_sent:
        flight.delay_notification_sent = True
        db.commit()
        return True
    
    return False


def get_tickets(db: Session, skip: int = 0, limit: int = 100, flight_id: Optional[int] = None):
    query = db.query(Ticket)
    if flight_id:
        query = query.filter(Ticket.flight_id == flight_id)
    return query.offset(skip).limit(limit).all()


def create_ticket(db: Session, ticket: TicketCreate):
    flight = get_flight(db, ticket.flight_id)
    if not flight:
        raise ValueError("航班不存在")
    
    if flight.status in [FlightStatus.DEPARTED, FlightStatus.ARRIVED, FlightStatus.CANCELLED]:
        raise ValueError("该航班不可售票")
    
    db_ticket = Ticket(**ticket.model_dump())
    db.add(db_ticket)
    db.commit()
    db.refresh(db_ticket)
    return db_ticket


def get_weather(db: Session, skip: int = 0, limit: int = 100, terminal_id: Optional[int] = None):
    query = db.query(Weather)
    if terminal_id:
        query = query.filter(Weather.terminal_id == terminal_id)
    return query.order_by(Weather.timestamp.desc()).offset(skip).limit(limit).all()


def create_weather(db: Session, weather: WeatherCreate):
    db_weather = Weather(**weather.model_dump())
    db.add(db_weather)
    db.commit()
    db.refresh(db_weather)
    return db_weather


def get_refunds(db: Session, skip: int = 0, limit: int = 100):
    return db.query(Refund).offset(skip).limit(limit).all()
