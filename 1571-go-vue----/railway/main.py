import os
from datetime import datetime, timedelta
from typing import List
from fastapi import FastAPI, Depends, HTTPException, status
from sqlmodel import Session, select

from .database import init_db, get_session
from .models import (
    Wagon, Station, FreightOrder, Train, TrainWagon,
    TrainArrivalDeparture, OrderStatus, CarType, CargoType,
    is_cargo_compatible
)
from .schemas import (
    OrderCreate, AssignRequest, LoadRequest,
    TrainCreate, StationCreate, WagonCreate,
    OrderResponse, WagonResponse, StationSummary, TrainResponse
)
from .services import (
    check_status_transition, check_not_cancelled,
    check_cargo_compatibility, check_weight_capacity,
    validate_train_composition, ensure_train_not_departed,
    calculate_train_weight
)

app = FastAPI(title="铁路货运调度系统")


@app.on_event("startup")
def on_startup():
    init_db()


def get_order_or_404(session: Session, order_no: str) -> FreightOrder:
    statement = select(FreightOrder).where(FreightOrder.order_no == order_no)
    order = session.exec(statement).first()
    if not order:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"订单 {order_no} 不存在"
        )
    return order


def get_train_or_404(session: Session, train_no: str) -> Train:
    statement = select(Train).where(Train.train_no == train_no)
    train = session.exec(statement).first()
    if not train:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"列车 {train_no} 不存在"
        )
    return train


def get_station_or_404(session: Session, code: str) -> Station:
    statement = select(Station).where(Station.code == code)
    station = session.exec(statement).first()
    if not station:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"车站 {code} 不存在"
        )
    return station


@app.post("/orders/{order_no}/record", response_model=OrderResponse, status_code=status.HTTP_201_CREATED)
def record_order(order_no: str, data: OrderCreate, session: Session = Depends(get_session)):
    statement = select(FreightOrder).where(FreightOrder.order_no == order_no)
    existing = session.exec(statement).first()
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"订单号 {order_no} 已存在"
        )
    
    order = FreightOrder(order_no=order_no, **data.model_dump())
    session.add(order)
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/accept", response_model=OrderResponse)
def accept_order(order_no: str, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    check_status_transition(order, OrderStatus.CREATED, "受理")
    
    order.status = OrderStatus.ACCEPTED
    order.accepted_at = datetime.now()
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/assign", response_model=OrderResponse)
def assign_order(order_no: str, data: AssignRequest, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    check_status_transition(order, OrderStatus.ACCEPTED, "配车")
    
    statement = select(Wagon).where(Wagon.id == data.wagon_id)
    wagon = session.exec(statement).first()
    if not wagon:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"车皮 {data.wagon_id} 不存在"
        )
    
    check_cargo_compatibility(order.cargo_type, wagon)
    check_weight_capacity(order.weight_tons, wagon)
    
    if wagon.status != "空车":
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"车皮 {wagon.wagon_number} 不是空车"
        )
    
    if wagon.current_station != order.origin_station:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"车皮当前在 {wagon.current_station}，不在发站 {order.origin_station}"
        )
    
    order.assigned_wagon_id = wagon.id
    order.status = OrderStatus.ASSIGNED
    order.assigned_at = datetime.now()
    wagon.status = "已配车"
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/load", response_model=OrderResponse)
def load_order(order_no: str, data: LoadRequest, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    check_status_transition(order, OrderStatus.ASSIGNED, "装车")
    
    if not order.assigned_wagon_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="订单未配车"
        )
    
    train_statement = select(Train).where(Train.id == data.train_id)
    train = session.exec(train_statement).first()
    if not train:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"列车 {data.train_id} 不存在"
        )
    
    ensure_train_not_departed(session, train)
    
    if train.origin_station != order.origin_station:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"列车发站 {train.origin_station} 与订单发站 {order.origin_station} 不一致"
        )
    
    if train.dest_station != order.dest_station:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"列车到站 {train.dest_station} 与订单到站 {order.dest_station} 不一致"
        )
    
    wagon_statement = select(Wagon).where(Wagon.id == order.assigned_wagon_id)
    wagon = session.exec(wagon_statement).first()
    if not wagon:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="车皮不存在"
        )
    
    validate_train_composition(session, train, wagon, order)
    
    max_pos_statement = select(TrainWagon.position).where(TrainWagon.train_id == train.id).order_by(TrainWagon.position.desc())
    max_pos = session.exec(max_pos_statement).first()
    next_position = (max_pos or 0) + 1
    
    train_wagon = TrainWagon(
        train_id=train.id,
        wagon_id=wagon.id,
        position=next_position,
        order_id=order.id
    )
    session.add(train_wagon)
    
    wagon.current_train_id = train.id
    wagon.status = "重车"
    
    order.train_id = train.id
    order.status = OrderStatus.LOADED
    order.loaded_at = datetime.now()
    
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/depart", response_model=OrderResponse)
def depart_order(order_no: str, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    check_status_transition(order, OrderStatus.LOADED, "发运")
    
    if not order.train_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="订单未编入列车"
        )
    
    train_statement = select(Train).where(Train.id == order.train_id)
    train = session.exec(train_statement).first()
    
    if train.departure_time is None:
        train.departure_time = datetime.now()
        train_arrival = TrainArrivalDeparture(
            train_no=train.train_no,
            station_code=train.origin_station,
            event_type="出发",
            event_time=train.departure_time
        )
        session.add(train_arrival)
    
    order.status = OrderStatus.DEPARTED
    order.departed_at = datetime.now()
    
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/arrive", response_model=OrderResponse)
def arrive_order(order_no: str, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    check_status_transition(order, OrderStatus.DEPARTED, "到达")
    
    if not order.train_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="订单未编入列车"
        )
    
    train_statement = select(Train).where(Train.id == order.train_id)
    train = session.exec(train_statement).first()
    
    if train.arrival_time is None:
        train.arrival_time = datetime.now()
        train_arrival = TrainArrivalDeparture(
            train_no=train.train_no,
            station_code=train.dest_station,
            event_type="到达",
            event_time=train.arrival_time
        )
        session.add(train_arrival)
    
    order.status = OrderStatus.ARRIVED
    order.arrived_at = datetime.now()
    
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/deliver", response_model=OrderResponse)
def deliver_order(order_no: str, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    check_status_transition(order, OrderStatus.ARRIVED, "交付")
    
    order.status = OrderStatus.DELIVERED
    order.delivered_at = datetime.now()
    
    if order.assigned_wagon_id:
        wagon_statement = select(Wagon).where(Wagon.id == order.assigned_wagon_id)
        wagon = session.exec(wagon_statement).first()
        if wagon:
            wagon.status = "空车"
            wagon.current_train_id = None
            wagon.current_station = order.dest_station
    
    session.commit()
    session.refresh(order)
    return order


@app.post("/orders/{order_no}/cancel", response_model=OrderResponse)
def cancel_order(order_no: str, session: Session = Depends(get_session)):
    order = get_order_or_404(session, order_no)
    
    if order.status == OrderStatus.DELIVERED:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="已交付的订单无法取消"
        )
    
    if order.status == OrderStatus.CANCELLED:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="订单已取消"
        )
    
    if order.assigned_wagon_id:
        wagon_statement = select(Wagon).where(Wagon.id == order.assigned_wagon_id)
        wagon = session.exec(wagon_statement).first()
        if wagon:
            wagon.status = "空车"
            wagon.current_train_id = None
    
    if order.train_id and order.status in [OrderStatus.LOADED, OrderStatus.DEPARTED, OrderStatus.ARRIVED]:
        tw_statement = select(TrainWagon).where(
            TrainWagon.train_id == order.train_id,
            TrainWagon.order_id == order.id
        )
        train_wagon = session.exec(tw_statement).first()
        if train_wagon:
            session.delete(train_wagon)
    
    old_status = order.status
    order.status = OrderStatus.CANCELLED
    order.cancelled_at = datetime.now()
    
    session.commit()
    session.refresh(order)
    return order


@app.post("/trains", response_model=TrainResponse, status_code=status.HTTP_201_CREATED)
def create_train(data: TrainCreate, session: Session = Depends(get_session)):
    statement = select(Train).where(Train.train_no == data.train_no)
    existing = session.exec(statement).first()
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"车次 {data.train_no} 已存在"
        )
    
    train = Train(**data.model_dump())
    session.add(train)
    session.commit()
    session.refresh(train)
    return TrainResponse(
        id=train.id,
        train_no=train.train_no,
        origin_station=train.origin_station,
        dest_station=train.dest_station,
        line_traction_tons=train.line_traction_tons,
        max_wagons=train.max_wagons,
        departure_time=train.departure_time,
        arrival_time=train.arrival_time,
        wagon_count=0,
        total_weight=0.0
    )


@app.get("/trains/{train_no}", response_model=TrainResponse)
def get_train(train_no: str, session: Session = Depends(get_session)):
    train = get_train_or_404(session, train_no)
    
    wagons_statement = select(TrainWagon).where(TrainWagon.train_id == train.id)
    wagon_count = len(session.exec(wagons_statement).all())
    
    total_weight = calculate_train_weight(session, train)
    
    return TrainResponse(
        id=train.id,
        train_no=train.train_no,
        origin_station=train.origin_station,
        dest_station=train.dest_station,
        line_traction_tons=train.line_traction_tons,
        max_wagons=train.max_wagons,
        departure_time=train.departure_time,
        arrival_time=train.arrival_time,
        wagon_count=wagon_count,
        total_weight=total_weight
    )


@app.post("/trains/{train_no}/arrival")
def record_train_arrival(train_no: str, session: Session = Depends(get_session)):
    train = get_train_or_404(session, train_no)
    
    arrival = TrainArrivalDeparture(
        train_no=train_no,
        station_code=train.dest_station,
        event_type="到达",
        event_time=datetime.now()
    )
    session.add(arrival)
    
    if train.arrival_time is None:
        train.arrival_time = arrival.event_time
    
    session.commit()
    return {"message": f"列车 {train_no} 到达 {train.dest_station}", "time": arrival.event_time}


@app.post("/trains/{train_no}/departure")
def record_train_departure(train_no: str, session: Session = Depends(get_session)):
    train = get_train_or_404(session, train_no)
    
    departure = TrainArrivalDeparture(
        train_no=train_no,
        station_code=train.origin_station,
        event_type="出发",
        event_time=datetime.now()
    )
    session.add(departure)
    
    if train.departure_time is None:
        train.departure_time = departure.event_time
    
    session.commit()
    return {"message": f"列车 {train_no} 从 {train.origin_station} 出发", "time": departure.event_time}


@app.post("/stations", status_code=status.HTTP_201_CREATED)
def create_station(data: StationCreate, session: Session = Depends(get_session)):
    statement = select(Station).where(Station.code == data.code)
    existing = session.exec(statement).first()
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"车站编码 {data.code} 已存在"
        )
    
    station = Station(**data.model_dump())
    session.add(station)
    session.commit()
    session.refresh(station)
    return station


@app.post("/wagons", response_model=WagonResponse, status_code=status.HTTP_201_CREATED)
def create_wagon(data: WagonCreate, session: Session = Depends(get_session)):
    statement = select(Wagon).where(Wagon.wagon_number == data.wagon_number)
    existing = session.exec(statement).first()
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"车皮编号 {data.wagon_number} 已存在"
        )
    
    wagon = Wagon(**data.model_dump())
    session.add(wagon)
    session.commit()
    session.refresh(wagon)
    return wagon


@app.get("/wagons/available")
def list_available_wagons(station: str, car_type: str, session: Session = Depends(get_session)):
    if car_type not in ["敞车", "棚车", "罐车", "平车", "冷藏车"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="无效的车种类型"
        )
    
    statement = select(Wagon).where(
        Wagon.current_station == station,
        Wagon.status == "空车",
        Wagon.car_type == car_type
    )
    wagons = session.exec(statement).all()
    return wagons


@app.get("/stats/stations/{station_code}/summary", response_model=StationSummary)
def get_station_summary(station_code: str, session: Session = Depends(get_session)):
    get_station_or_404(session, station_code)
    
    empty_statement = select(Wagon).where(
        Wagon.current_station == station_code,
        Wagon.status == "空车"
    )
    empty_count = len(session.exec(empty_statement).all())
    
    loaded_statement = select(Wagon).where(
        Wagon.current_station == station_code,
        Wagon.status == "重车"
    )
    loaded_count = len(session.exec(loaded_statement).all())
    
    shipped_statement = select(FreightOrder).where(
        FreightOrder.origin_station == station_code,
        FreightOrder.departed_at.isnot(None)
    )
    shipped_orders = session.exec(shipped_statement).all()
    shipped_tons = sum(o.weight_tons for o in shipped_orders)
    
    turnaround_times = []
    slow_count = 0
    
    for order in shipped_orders:
        if order.arrived_at and order.departed_at:
            delta = order.arrived_at - order.departed_at
            hours = delta.total_seconds() / 3600
            turnaround_times.append(hours)
            if hours > 48:
                slow_count += 1
    
    avg_turnaround = sum(turnaround_times) / len(turnaround_times) if turnaround_times else 0.0
    
    return StationSummary(
        station_code=station_code,
        empty_wagons=empty_count,
        loaded_wagons=loaded_count,
        shipped_tons=shipped_tons,
        avg_turnaround_hours=round(avg_turnaround, 2),
        slow_turnaround_count=slow_count
    )


@app.get("/orders/{order_no}", response_model=OrderResponse)
def get_order(order_no: str, session: Session = Depends(get_session)):
    return get_order_or_404(session, order_no)


@app.get("/orders", response_model=List[OrderResponse])
def list_orders(status: str = None, session: Session = Depends(get_session)):
    statement = select(FreightOrder)
    if status:
        if status in OrderStatus.__members__:
            statement = statement.where(FreightOrder.status == OrderStatus[status])
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"无效的订单状态: {status}"
            )
    orders = session.exec(statement).all()
    return orders


@app.get("/wagons", response_model=List[WagonResponse])
def list_wagons(station: str = None, session: Session = Depends(get_session)):
    statement = select(Wagon)
    if station:
        statement = statement.where(Wagon.current_station == station)
    wagons = session.exec(statement).all()
    return wagons


@app.get("/stations")
def list_stations(session: Session = Depends(get_session)):
    statement = select(Station)
    stations = session.exec(statement).all()
    return stations


@app.get("/trains", response_model=List[TrainResponse])
def list_trains(session: Session = Depends(get_session)):
    statement = select(Train)
    trains = session.exec(statement).all()
    result = []
    for train in trains:
        wagons_statement = select(TrainWagon).where(TrainWagon.train_id == train.id)
        wagon_count = len(session.exec(wagons_statement).all())
        total_weight = calculate_train_weight(session, train)
        result.append(TrainResponse(
            id=train.id,
            train_no=train.train_no,
            origin_station=train.origin_station,
            dest_station=train.dest_station,
            line_traction_tons=train.line_traction_tons,
            max_wagons=train.max_wagons,
            departure_time=train.departure_time,
            arrival_time=train.arrival_time,
            wagon_count=wagon_count,
            total_weight=total_weight
        ))
    return result
