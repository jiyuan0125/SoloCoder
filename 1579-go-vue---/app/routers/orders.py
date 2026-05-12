from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from app.database import get_db
from app.models import (
    Order, Container, Vehicle, Segment, OrderContainer, ContainerTrajectory,
    OrderStatus, SegmentStatus, VehicleStatus, ContainerStatus, Station
)
from app.schemas import (
    OrderCreate, OrderResponse, OperationResponse, 
    ContainerStatusResponse, TransportPlanResponse, TrajectoryResponse
)
from app.services import (
    create_order, segment_depart, segment_arrive, segment_cancel,
    assign_vehicle, release_vehicle, vehicle_maintenance,
    validate_container_vehicle_compatibility, create_trajectory_record
)

router = APIRouter(prefix="/orders", tags=["orders"])


@router.post("", response_model=OrderResponse)
def create_new_order(order: OrderCreate, db: Session = Depends(get_db)):
    order_dict = order.model_dump()
    db_order = create_order(db, order_dict)
    db.commit()
    db.refresh(db_order)
    
    containers = []
    for oc in db_order.order_containers:
        containers.append(oc.container)
    
    db_order.containers = containers
    return db_order


@router.get("", response_model=List[OrderResponse])
def list_orders(db: Session = Depends(get_db)):
    orders = db.query(Order).all()
    for order in orders:
        containers = []
        for oc in order.order_containers:
            containers.append(oc.container)
        order.containers = containers
    return orders


@router.get("/{order_number}", response_model=OrderResponse)
def get_order(order_number: str, db: Session = Depends(get_db)):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    containers = []
    for oc in order.order_containers:
        containers.append(oc.container)
    order.containers = containers
    return order


@router.get("/{order_number}/plan", response_model=TransportPlanResponse)
def export_transport_plan(order_number: str, db: Session = Depends(get_db)):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    segments_data = []
    for segment in order.segments:
        seg_data = {
            "segment_number": segment.segment_number,
            "sequence": segment.sequence,
            "mode": segment.mode.value,
            "status": segment.status.value,
            "origin": segment.origin_station.code if segment.origin_station else None,
            "destination": segment.destination_station.code if segment.destination_station else None,
            "is_transit": segment.is_transit_point == 1,
            "planned_departure": segment.planned_departure.isoformat() if segment.planned_departure else None,
            "planned_arrival": segment.planned_arrival.isoformat() if segment.planned_arrival else None,
            "actual_departure": segment.actual_departure.isoformat() if segment.actual_departure else None,
            "actual_arrival": segment.actual_arrival.isoformat() if segment.actual_arrival else None,
            "assigned_vehicle": segment.assigned_vehicle.vehicle_id if segment.assigned_vehicle else None
        }
        segments_data.append(seg_data)
    
    containers_data = []
    for oc in order.order_containers:
        container = oc.container
        containers_data.append({
            "container_number": container.container_number,
            "size": container.size.value,
            "status": container.status.value,
            "current_station": container.current_station_id,
            "current_vehicle": container.current_vehicle_id
        })
    
    return TransportPlanResponse(
        order_number=order.order_number,
        status=order.status,
        deadline=order.deadline,
        segments=segments_data,
        containers=containers_data
    )


@router.post("/{order_number}/containers/{container_number}/load", response_model=OperationResponse)
def load_container_to_vehicle(
    order_number: str,
    container_number: str,
    vehicle_id: str = Query(..., description="运输工具ID"),
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    container = db.query(Container).filter(
        Container.container_number == container_number
    ).first()
    if not container:
        raise HTTPException(status_code=404, detail="Container not found")
    
    order_container = db.query(OrderContainer).filter(
        OrderContainer.order_id == order.id,
        OrderContainer.container_id == container.id
    ).first()
    if not order_container:
        raise HTTPException(status_code=400, detail="Container not part of this order")
    
    vehicle = db.query(Vehicle).filter(Vehicle.vehicle_id == vehicle_id).first()
    if not vehicle:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    
    if container.current_vehicle_id:
        raise HTTPException(status_code=400, detail="Container is already loaded")
    
    if not validate_container_vehicle_compatibility(container.size, vehicle.max_container_size):
        raise HTTPException(
            status_code=400,
            detail="40ft container cannot be loaded into 20ft vehicle"
        )
    
    if vehicle.status not in [VehicleStatus.AVAILABLE, VehicleStatus.ASSIGNED]:
        raise HTTPException(status_code=400, detail="Vehicle is not available")
    
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
        description=f"Container loaded onto vehicle {vehicle_id}"
    )
    
    db.commit()
    return OperationResponse(
        success=True,
        message=f"Container {container_number} loaded onto vehicle {vehicle_id}"
    )


@router.post("/{order_number}/containers/{container_number}/unload", response_model=OperationResponse)
def unload_container_from_vehicle(
    order_number: str,
    container_number: str,
    station_code: str = Query(..., description="站点代码"),
    is_transit: bool = Query(True, description="是否为中转站"),
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    container = db.query(Container).filter(
        Container.container_number == container_number
    ).first()
    if not container:
        raise HTTPException(status_code=404, detail="Container not found")
    
    order_container = db.query(OrderContainer).filter(
        OrderContainer.order_id == order.id,
        OrderContainer.container_id == container.id
    ).first()
    if not order_container:
        raise HTTPException(status_code=400, detail="Container not part of this order")
    
    if not container.current_vehicle_id:
        raise HTTPException(status_code=400, detail="Container is not loaded")
    
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")
    
    vehicle = db.query(Vehicle).filter(Vehicle.id == container.current_vehicle_id).first()
    
    container.current_vehicle_id = None
    container.current_station_id = station.id
    
    if is_transit:
        container.status = ContainerStatus.WAITING_TRANSIT
    else:
        container.status = ContainerStatus.ARRIVED
    
    if vehicle:
        vehicle.status = VehicleStatus.AVAILABLE
        vehicle.current_station_id = station.id
    
    create_trajectory_record(
        db,
        container_id=container.id,
        order_id=order.id,
        station_id=station.id,
        vehicle_id=vehicle.id if vehicle else None,
        event_type="unloaded",
        description=f"Container unloaded at {station.name}, status: {container.status.value}"
    )
    
    db.commit()
    return OperationResponse(
        success=True,
        message=f"Container {container_number} unloaded at {station_code}"
    )


@router.get("/{order_number}/containers/{container_number}/status", response_model=ContainerStatusResponse)
def get_container_status(
    order_number: str,
    container_number: str,
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    container = db.query(Container).filter(
        Container.container_number == container_number
    ).first()
    if not container:
        raise HTTPException(status_code=404, detail="Container not found")
    
    order_container = db.query(OrderContainer).filter(
        OrderContainer.order_id == order.id,
        OrderContainer.container_id == container.id
    ).first()
    if not order_container:
        raise HTTPException(status_code=400, detail="Container not part of this order")
    
    current_station = db.query(Station).filter(Station.id == container.current_station_id).first() if container.current_station_id else None
    current_vehicle = db.query(Vehicle).filter(Vehicle.id == container.current_vehicle_id).first() if container.current_vehicle_id else None
    
    return ContainerStatusResponse(
        container_number=container.container_number,
        status=container.status,
        current_station=current_station,
        current_vehicle=current_vehicle,
        last_update=container.updated_at
    )


@router.get("/{order_number}/containers/{container_number}/trajectory", response_model=TrajectoryResponse)
def get_container_trajectory(
    order_number: str,
    container_number: str,
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    container = db.query(Container).filter(
        Container.container_number == container_number
    ).first()
    if not container:
        raise HTTPException(status_code=404, detail="Container not found")
    
    trajectories = db.query(ContainerTrajectory).filter(
        ContainerTrajectory.container_id == container.id,
        ContainerTrajectory.order_id == order.id
    ).order_by(ContainerTrajectory.event_time).all()
    
    events = []
    for t in trajectories:
        event = {
            "event_type": t.event_type,
            "event_time": t.event_time.isoformat(),
            "station": t.station.code if t.station else None,
            "vehicle": t.vehicle.vehicle_id if t.vehicle else None,
            "description": t.description
        }
        events.append(event)
    
    return TrajectoryResponse(
        container_number=container_number,
        events=events
    )


@router.post("/{order_number}/segments/{segment_number}/depart", response_model=OperationResponse)
def depart_segment(
    order_number: str,
    segment_number: str,
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    segment = db.query(Segment).filter(
        Segment.segment_number == segment_number,
        Segment.order_id == order.id
    ).first()
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    
    segment_depart(db, segment)
    db.commit()
    
    return OperationResponse(
        success=True,
        message=f"Segment {segment_number} departed"
    )


@router.post("/{order_number}/segments/{segment_number}/arrive", response_model=OperationResponse)
def arrive_segment(
    order_number: str,
    segment_number: str,
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    segment = db.query(Segment).filter(
        Segment.segment_number == segment_number,
        Segment.order_id == order.id
    ).first()
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    
    segment_arrive(db, segment)
    db.commit()
    
    return OperationResponse(
        success=True,
        message=f"Segment {segment_number} arrived"
    )


@router.post("/{order_number}/segments/{segment_number}/cancel", response_model=OperationResponse)
def cancel_segment(
    order_number: str,
    segment_number: str,
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    segment = db.query(Segment).filter(
        Segment.segment_number == segment_number,
        Segment.order_id == order.id
    ).first()
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    
    segment_cancel(db, segment)
    db.commit()
    
    return OperationResponse(
        success=True,
        message=f"Segment {segment_number} cancelled (and subsequent pending segments)"
    )


@router.post("/{order_number}/vehicles/{vehicle_id}/assign", response_model=OperationResponse)
def assign_vehicle_to_segment(
    order_number: str,
    vehicle_id: str,
    segment_number: str = Query(..., description="运输段编号"),
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    vehicle = db.query(Vehicle).filter(Vehicle.vehicle_id == vehicle_id).first()
    if not vehicle:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    
    segment = db.query(Segment).filter(
        Segment.segment_number == segment_number,
        Segment.order_id == order.id
    ).first()
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    
    assign_vehicle(db, segment, vehicle)
    db.commit()
    
    return OperationResponse(
        success=True,
        message=f"Vehicle {vehicle_id} assigned to segment {segment_number}"
    )


@router.post("/{order_number}/vehicles/{vehicle_id}/release", response_model=OperationResponse)
def release_vehicle_from_segment(
    order_number: str,
    vehicle_id: str,
    segment_number: str = Query(..., description="运输段编号"),
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    vehicle = db.query(Vehicle).filter(Vehicle.vehicle_id == vehicle_id).first()
    if not vehicle:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    
    segment = db.query(Segment).filter(
        Segment.segment_number == segment_number,
        Segment.order_id == order.id
    ).first()
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    
    if segment.assigned_vehicle_id != vehicle.id:
        raise HTTPException(status_code=400, detail="Vehicle not assigned to this segment")
    
    release_vehicle(db, segment)
    db.commit()
    
    return OperationResponse(
        success=True,
        message=f"Vehicle {vehicle_id} released from segment {segment_number}"
    )


@router.post("/{order_number}/vehicles/{vehicle_id}/maintenance", response_model=OperationResponse)
def send_vehicle_to_maintenance(
    order_number: str,
    vehicle_id: str,
    db: Session = Depends(get_db)
):
    order = db.query(Order).filter(Order.order_number == order_number).first()
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    vehicle = db.query(Vehicle).filter(Vehicle.vehicle_id == vehicle_id).first()
    if not vehicle:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    
    assigned_segment = db.query(Segment).filter(
        Segment.assigned_vehicle_id == vehicle.id,
        Segment.order_id == order.id
    ).first()
    if assigned_segment:
        if assigned_segment.status == SegmentStatus.IN_PROGRESS:
            raise HTTPException(
                status_code=400,
                detail="Cannot send vehicle to maintenance while it's in use"
            )
        release_vehicle(db, assigned_segment)
    
    vehicle_maintenance(db, vehicle)
    db.commit()
    
    return OperationResponse(
        success=True,
        message=f"Vehicle {vehicle_id} sent to maintenance"
    )
