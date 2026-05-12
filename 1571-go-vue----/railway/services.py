from datetime import datetime, timedelta
from typing import Optional, List
from sqlmodel import Session, select, func
from fastapi import HTTPException, status

from .models import (
    Wagon, Station, FreightOrder, Train, TrainWagon,
    TrainArrivalDeparture, OrderStatus, CarType, CargoType,
    is_cargo_compatible
)
from .schemas import (
    OrderCreate, AssignRequest, LoadRequest,
    TrainCreate, TrainAddWagon, StationCreate, WagonCreate,
    StationSummary, TrainResponse
)


def ensure_train_not_departed(session: Session, train: Train):
    if train.departure_time is not None:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"列车 {train.train_no} 已发运，无法修改编组"
        )


def check_cargo_compatibility(cargo_type: CargoType, wagon: Wagon):
    if not is_cargo_compatible(cargo_type, wagon.car_type):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"货物类型 {cargo_type.value} 与车种 {wagon.car_type.value} 不匹配"
        )


def check_status_transition(order: FreightOrder, expected_status: OrderStatus, operation: str):
    if order.status != expected_status:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"订单当前状态为 {order.status.value}，无法执行 {operation} 操作"
        )
    if order.status == OrderStatus.CANCELLED:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="订单已取消，无法执行任何操作"
        )


def check_not_cancelled(order: FreightOrder):
    if order.status == OrderStatus.CANCELLED:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="订单已取消，无法执行该操作"
        )


def check_weight_capacity(weight_tons: float, wagon: Wagon):
    if weight_tons > wagon.capacity_tons:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"货物重量 {weight_tons} 吨超过车皮载重 {wagon.capacity_tons} 吨"
        )


def calculate_train_weight(session: Session, train: Train) -> float:
    statement = select(func.sum(FreightOrder.weight_tons)).where(
        FreightOrder.train_id == train.id,
        FreightOrder.status.in_([OrderStatus.LOADED, OrderStatus.DEPARTED, OrderStatus.ARRIVED, OrderStatus.DELIVERED])
    )
    result = session.exec(statement).first()
    return result or 0.0


def validate_train_composition(session: Session, train: Train, new_wagon: Optional[Wagon] = None, new_order: Optional[FreightOrder] = None):
    statement = select(TrainWagon).where(TrainWagon.train_id == train.id).order_by(TrainWagon.position)
    train_wagons = list(session.exec(statement).all())
    
    if len(train_wagons) >= train.max_wagons:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"列车编组已达最大 {train.max_wagons} 节"
        )
    
    total_weight = calculate_train_weight(session, train)
    if new_order:
        total_weight += new_order.weight_tons
    
    if total_weight > train.line_traction_tons:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"列车总重 {total_weight} 吨超过线路牵引定数 {train.line_traction_tons} 吨"
        )
    
    if new_wagon and new_wagon.car_type == CarType.TANK:
        if len(train_wagons) == 0:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="罐车不能编在机车后面第一节"
            )
