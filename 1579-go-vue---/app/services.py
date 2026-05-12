from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from fastapi import HTTPException
from app.models import (
    Station, Route, RouteSegment, Container, Vehicle, Order,
    OrderContainer, Segment, ContainerTrajectory, TransportMode,
    ContainerSize, ContainerStatus, SegmentStatus, VehicleStatus, OrderStatus
)
import uuid


def generate_id(prefix: str) -> str:
    return f"{prefix}-{uuid.uuid4().hex[:8].upper()}"


def validate_container_vehicle_compatibility(
    container_size: ContainerSize,
    vehicle_max_size: ContainerSize
) -> bool:
    if container_size == ContainerSize.SIZE_40FT:
        return vehicle_max_size == ContainerSize.SIZE_40FT
    return True


def create_trajectory_record(
    db: Session,
    container_id: int,
    event_type: str,
    order_id: Optional[int] = None,
    segment_id: Optional[int] = None,
    station_id: Optional[int] = None,
    vehicle_id: Optional[int] = None,
    description: str = ""
) -> ContainerTrajectory:
    trajectory = ContainerTrajectory(
        container_id=container_id,
        order_id=order_id,
        segment_id=segment_id,
        event_type=event_type,
        station_id=station_id,
        vehicle_id=vehicle_id,
        description=description
    )
    db.add(trajectory)
    db.flush()
    return trajectory


def find_route(db: Session, origin_station_id: int, destination_station_id: int) -> Optional[Route]:
    return db.query(Route).filter(
        Route.origin_station_id == origin_station_id,
        Route.destination_station_id == destination_station_id
    ).first()


def find_available_vehicle(
    db: Session,
    mode: TransportMode,
    container_size: ContainerSize,
    station_id: Optional[int] = None
) -> Optional[Vehicle]:
    query = db.query(Vehicle).filter(
        Vehicle.status == VehicleStatus.AVAILABLE,
        Vehicle.mode == mode
    )
    
    if station_id:
        query = query.filter(Vehicle.current_station_id == station_id)
    
    vehicles = query.all()
    
    for vehicle in vehicles:
        if validate_container_vehicle_compatibility(container_size, vehicle.max_container_size):
            return vehicle
    
    return None


def split_order_into_segments(
    db: Session,
    order: Order,
    route: Route,
    deadline: datetime
) -> List[Segment]:
    route_segments = sorted(route.segments, key=lambda s: s.sequence)
    total_estimated_hours = sum(rs.estimated_hours for rs in route_segments)
    
    if total_estimated_hours == 0:
        total_estimated_hours = len(route_segments) * 24
    
    current_time = datetime.utcnow()
    segments: List[Segment] = []
    accumulated_hours = 0
    
    for route_segment in route_segments:
        segment_number = generate_id("SEG")
        
        planned_departure = current_time + timedelta(hours=accumulated_hours)
        planned_arrival = planned_departure + timedelta(hours=route_segment.estimated_hours)
        
        segment = Segment(
            segment_number=segment_number,
            order_id=order.id,
            route_segment_id=route_segment.id,
            sequence=route_segment.sequence,
            mode=route_segment.mode,
            origin_station_id=route_segment.origin_station_id,
            destination_station_id=route_segment.destination_station_id,
            is_transit_point=route_segment.is_transit_point,
            planned_departure=planned_departure,
            planned_arrival=planned_arrival,
            status=SegmentStatus.PENDING
        )
        segments.append(segment)
        accumulated_hours += route_segment.estimated_hours
    
    return segments


def update_order_status(db: Session, order: Order):
    segments = db.query(Segment).filter(Segment.order_id == order.id).all()
    
    if not segments:
        return
    
    statuses = [s.status for s in segments]
    
    if SegmentStatus.CANCELLED in statuses and all(
        s in [SegmentStatus.CANCELLED, SegmentStatus.ARRIVED] 
        for s in statuses
    ):
        order.status = OrderStatus.CANCELLED
    elif all(s == SegmentStatus.ARRIVED for s in statuses):
        order.status = OrderStatus.COMPLETED
    elif any(s in [SegmentStatus.IN_PROGRESS, SegmentStatus.ASSIGNED] for s in statuses):
        order.status = OrderStatus.IN_PROGRESS
    
    db.flush()


def cascade_cancel_segments(db: Session, order_id: int, cancelled_segment_sequence: int):
    segments = db.query(Segment).filter(
        Segment.order_id == order_id,
        Segment.sequence > cancelled_segment_sequence,
        Segment.status.notin_([SegmentStatus.ARRIVED, SegmentStatus.CANCELLED])
    ).all()
    
    for segment in segments:
        if segment.assigned_vehicle_id:
            vehicle = db.query(Vehicle).filter(Vehicle.id == segment.assigned_vehicle_id).first()
            if vehicle:
                vehicle.status = VehicleStatus.AVAILABLE
                vehicle.current_station_id = segment.origin_station_id
        
        segment.status = SegmentStatus.CANCELLED
    
    db.flush()


def create_order(db: Session, order_data: dict) -> Order:
    origin_station = db.query(Station).filter(
        Station.code == order_data["origin_station_code"]
    ).first()
    if not origin_station:
        raise HTTPException(status_code=404, detail="Origin station not found")
    
    destination_station = db.query(Station).filter(
        Station.code == order_data["destination_station_code"]
    ).first()
    if not destination_station:
        raise HTTPException(status_code=404, detail="Destination station not found")
    
    route = find_route(db, origin_station.id, destination_station.id)
    if not route:
        raise HTTPException(
            status_code=400, 
            detail=f"No route found from {origin_station.code} to {destination_station.code}"
        )
    
    containers: List[Container] = []
    for container_number in order_data["containers"]:
        container = db.query(Container).filter(
            Container.container_number == container_number
        ).first()
        if not container:
            raise HTTPException(status_code=404, detail=f"Container {container_number} not found")
        if container.status not in [ContainerStatus.AT_ORIGIN, ContainerStatus.WAITING_TRANSIT]:
            raise HTTPException(status_code=400, detail=f"Container {container_number} is not available")
        containers.append(container)
    
    order_number = generate_id("ORD")
    
    order = Order(
        order_number=order_number,
        origin_station_id=origin_station.id,
        destination_station_id=destination_station.id,
        deadline=order_data["deadline"],
        route_id=route.id,
        status=OrderStatus.PENDING
    )
    db.add(order)
    db.flush()
    
    for container in containers:
        order_container = OrderContainer(
            order_id=order.id,
            container_id=container.id
        )
        db.add(order_container)
        create_trajectory_record(
            db,
            container_id=container.id,
            event_type="order_assigned",
            order_id=order.id,
            description=f"Container assigned to order {order_number}"
        )
    
    segments = split_order_into_segments(db, order, route, order_data["deadline"])
    for segment in segments:
        db.add(segment)
    
    db.flush()
    db.refresh(order)
    
    return order


def load_container(
    db: Session,
    order: Order,
    container: Container,
    vehicle: Vehicle
):
    if not validate_container_vehicle_compatibility(container.size, vehicle.max_container_size):
        raise HTTPException(
            status_code=400,
            detail=f"40ft container cannot be loaded into 20ft vehicle"
        )
    
    container.status = ContainerStatus.IN_TRANSIT
    container.current_vehicle_id = vehicle.id
    container.current_station_id = None
    
    vehicle.status = VehicleStatus.IN_USE
    
    create_trajectory_record(
        db,
        container_id=container.id,
        order_id=order.id,
        vehicle_id=vehicle.id,
        event_type="loaded",
        description=f"Container loaded onto vehicle {vehicle.vehicle_id}"
    )
    
    db.flush()


def unload_container(
    db: Session,
    order: Order,
    container: Container,
    vehicle: Vehicle,
    station: Station,
    is_transit: bool
):
    container.current_vehicle_id = None
    container.current_station_id = station.id
    
    if is_transit:
        container.status = ContainerStatus.WAITING_TRANSIT
    else:
        container.status = ContainerStatus.ARRIVED
    
    vehicle.status = VehicleStatus.AVAILABLE
    vehicle.current_station_id = station.id
    
    create_trajectory_record(
        db,
        container_id=container.id,
        order_id=order.id,
        station_id=station.id,
        vehicle_id=vehicle.id,
        event_type="unloaded",
        description=f"Container unloaded at {station.name}, status: {container.status.value}"
    )
    
    db.flush()


def segment_depart(db: Session, segment: Segment):
    if segment.status != SegmentStatus.ASSIGNED:
        raise HTTPException(
            status_code=400,
            detail=f"Segment must be assigned before departure"
        )
    
    segment.status = SegmentStatus.IN_PROGRESS
    segment.actual_departure = datetime.utcnow()
    
    order_containers = db.query(OrderContainer).filter(
        OrderContainer.order_id == segment.order_id
    ).all()
    
    for oc in order_containers:
        container = db.query(Container).filter(Container.id == oc.container_id).first()
        if container and container.current_vehicle_id:
            create_trajectory_record(
                db,
                container_id=container.id,
                order_id=segment.order_id,
                segment_id=segment.id,
                vehicle_id=segment.assigned_vehicle_id,
                event_type="departure",
                description=f"Segment {segment.segment_number} departed"
            )
    
    update_order_status(db, db.query(Order).filter(Order.id == segment.order_id).first())
    db.flush()


def segment_arrive(db: Session, segment: Segment):
    if segment.status != SegmentStatus.IN_PROGRESS:
        raise HTTPException(
            status_code=400,
            detail=f"Segment is not in transit"
        )
    
    segment.status = SegmentStatus.ARRIVED
    segment.actual_arrival = datetime.utcnow()
    
    order = db.query(Order).filter(Order.id == segment.order_id).first()
    order_containers = db.query(OrderContainer).filter(
        OrderContainer.order_id == segment.order_id
    ).all()
    
    for oc in order_containers:
        container = db.query(Container).filter(Container.id == oc.container_id).first()
        if container and container.current_vehicle_id == segment.assigned_vehicle_id:
            is_transit = segment.is_transit_point == 1
            unload_container(
                db,
                order=order,
                container=container,
                vehicle=segment.assigned_vehicle,
                station=segment.destination_station,
                is_transit=is_transit
            )
    
    update_order_status(db, order)
    db.flush()


def segment_cancel(db: Session, segment: Segment):
    if segment.status in [SegmentStatus.ARRIVED, SegmentStatus.CANCELLED]:
        raise HTTPException(
            status_code=400,
            detail=f"Cannot cancel segment that has already arrived or been cancelled"
        )
    
    if segment.assigned_vehicle_id:
        vehicle = db.query(Vehicle).filter(Vehicle.id == segment.assigned_vehicle_id).first()
        if vehicle:
            vehicle.status = VehicleStatus.AVAILABLE
            vehicle.current_station_id = segment.origin_station_id
    
    segment.status = SegmentStatus.CANCELLED
    
    order = db.query(Order).filter(Order.id == segment.order_id).first()
    order_containers = db.query(OrderContainer).filter(
        OrderContainer.order_id == segment.order_id
    ).all()
    
    for oc in order_containers:
        create_trajectory_record(
            db,
            container_id=oc.container_id,
            order_id=segment.order_id,
            segment_id=segment.id,
            event_type="segment_cancelled",
            description=f"Segment {segment.segment_number} cancelled"
        )
    
    cascade_cancel_segments(db, segment.order_id, segment.sequence)
    update_order_status(db, order)
    db.flush()


def assign_vehicle(db: Session, segment: Segment, vehicle: Vehicle):
    order_containers = db.query(OrderContainer).filter(
        OrderContainer.order_id == segment.order_id
    ).all()
    
    for oc in order_containers:
        container = db.query(Container).filter(Container.id == oc.container_id).first()
        if container:
            if not validate_container_vehicle_compatibility(container.size, vehicle.max_container_size):
                raise HTTPException(
                    status_code=400,
                    detail=f"Vehicle {vehicle.vehicle_id} cannot accommodate container {container.container_number}"
                )
    
    if vehicle.status != VehicleStatus.AVAILABLE:
        raise HTTPException(
            status_code=400,
            detail=f"Vehicle {vehicle.vehicle_id} is not available"
        )
    
    segment.assigned_vehicle_id = vehicle.id
    segment.status = SegmentStatus.ASSIGNED
    vehicle.status = VehicleStatus.ASSIGNED
    
    for oc in order_containers:
        container = db.query(Container).filter(Container.id == oc.container_id).first()
        if container:
            load_container(db, segment.order, container, vehicle)
    
    order = db.query(Order).filter(Order.id == segment.order_id).first()
    update_order_status(db, order)
    db.flush()


def release_vehicle(db: Session, segment: Segment):
    if not segment.assigned_vehicle_id:
        raise HTTPException(
            status_code=400,
            detail="No vehicle assigned to this segment"
        )
    
    if segment.status == SegmentStatus.IN_PROGRESS:
        raise HTTPException(
            status_code=400,
            detail="Cannot release vehicle while segment is in progress"
        )
    
    vehicle = db.query(Vehicle).filter(Vehicle.id == segment.assigned_vehicle_id).first()
    if vehicle:
        vehicle.status = VehicleStatus.AVAILABLE
        vehicle.current_station_id = segment.origin_station_id
        
        order_containers = db.query(OrderContainer).filter(
            OrderContainer.order_id == segment.order_id
        ).all()
        
        for oc in order_containers:
            container = db.query(Container).filter(Container.id == oc.container_id).first()
            if container and container.current_vehicle_id == vehicle.id:
                container.current_vehicle_id = None
                container.current_station_id = segment.origin_station_id
                if segment.sequence == 1:
                    container.status = ContainerStatus.AT_ORIGIN
                else:
                    container.status = ContainerStatus.WAITING_TRANSIT
    
    segment.assigned_vehicle_id = None
    segment.status = SegmentStatus.PENDING
    
    order = db.query(Order).filter(Order.id == segment.order_id).first()
    update_order_status(db, order)
    db.flush()


def vehicle_maintenance(db: Session, vehicle: Vehicle):
    if vehicle.status == VehicleStatus.IN_USE:
        raise HTTPException(
            status_code=400,
            detail="Cannot send vehicle to maintenance while it's in use"
        )
    
    vehicle.status = VehicleStatus.MAINTENANCE
    db.flush()
