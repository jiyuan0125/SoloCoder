from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Enum, Float, Text
from sqlalchemy.orm import relationship
from datetime import datetime
import enum
from app.database import Base


class TransportMode(str, enum.Enum):
    ROAD = "road"
    RAIL = "rail"
    WATER = "water"
    AIR = "air"


class ContainerSize(str, enum.Enum):
    SIZE_20FT = "20ft"
    SIZE_40FT = "40ft"


class ContainerStatus(str, enum.Enum):
    AT_ORIGIN = "at_origin"
    WAITING_TRANSIT = "waiting_transit"
    IN_TRANSIT = "in_transit"
    ARRIVED = "arrived"


class SegmentStatus(str, enum.Enum):
    PENDING = "pending"
    ASSIGNED = "assigned"
    IN_PROGRESS = "in_progress"
    ARRIVED = "arrived"
    CANCELLED = "cancelled"


class VehicleStatus(str, enum.Enum):
    AVAILABLE = "available"
    ASSIGNED = "assigned"
    IN_USE = "in_use"
    MAINTENANCE = "maintenance"


class OrderStatus(str, enum.Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class Station(Base):
    __tablename__ = "stations"
    
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(20), unique=True, index=True)
    name = Column(String(100))
    city = Column(String(100))
    country = Column(String(100))
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Route(Base):
    __tablename__ = "routes"
    
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(50), unique=True, index=True)
    name = Column(String(200))
    origin_station_id = Column(Integer, ForeignKey("stations.id"))
    destination_station_id = Column(Integer, ForeignKey("stations.id"))
    
    origin_station = relationship("Station", foreign_keys=[origin_station_id])
    destination_station = relationship("Station", foreign_keys=[destination_station_id])
    segments = relationship("RouteSegment", back_populates="route", order_by="RouteSegment.sequence")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class RouteSegment(Base):
    __tablename__ = "route_segments"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"))
    sequence = Column(Integer)
    
    mode = Column(Enum(TransportMode))
    origin_station_id = Column(Integer, ForeignKey("stations.id"))
    destination_station_id = Column(Integer, ForeignKey("stations.id"))
    
    estimated_hours = Column(Float, default=0)
    is_transit_point = Column(Integer, default=0)
    
    route = relationship("Route", back_populates="segments")
    origin_station = relationship("Station", foreign_keys=[origin_station_id])
    destination_station = relationship("Station", foreign_keys=[destination_station_id])
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Container(Base):
    __tablename__ = "containers"
    
    id = Column(Integer, primary_key=True, index=True)
    container_number = Column(String(50), unique=True, index=True)
    size = Column(Enum(ContainerSize))
    status = Column(Enum(ContainerStatus), default=ContainerStatus.AT_ORIGIN)
    current_station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    current_vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Vehicle(Base):
    __tablename__ = "vehicles"
    
    id = Column(Integer, primary_key=True, index=True)
    vehicle_id = Column(String(50), unique=True, index=True)
    name = Column(String(100))
    mode = Column(Enum(TransportMode))
    max_container_size = Column(Enum(ContainerSize))
    status = Column(Enum(VehicleStatus), default=VehicleStatus.AVAILABLE)
    current_station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Order(Base):
    __tablename__ = "orders"
    
    id = Column(Integer, primary_key=True, index=True)
    order_number = Column(String(50), unique=True, index=True)
    status = Column(Enum(OrderStatus), default=OrderStatus.PENDING)
    
    origin_station_id = Column(Integer, ForeignKey("stations.id"))
    destination_station_id = Column(Integer, ForeignKey("stations.id"))
    deadline = Column(DateTime)
    
    route_id = Column(Integer, ForeignKey("routes.id"))
    
    origin_station = relationship("Station", foreign_keys=[origin_station_id])
    destination_station = relationship("Station", foreign_keys=[destination_station_id])
    route = relationship("Route")
    segments = relationship("Segment", back_populates="order", order_by="Segment.sequence")
    order_containers = relationship("OrderContainer", back_populates="order")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class OrderContainer(Base):
    __tablename__ = "order_containers"
    
    id = Column(Integer, primary_key=True, index=True)
    order_id = Column(Integer, ForeignKey("orders.id"))
    container_id = Column(Integer, ForeignKey("containers.id"))
    
    order = relationship("Order", back_populates="order_containers")
    container = relationship("Container")
    
    created_at = Column(DateTime, default=datetime.utcnow)


class Segment(Base):
    __tablename__ = "segments"
    
    id = Column(Integer, primary_key=True, index=True)
    segment_number = Column(String(50), unique=True, index=True)
    order_id = Column(Integer, ForeignKey("orders.id"))
    route_segment_id = Column(Integer, ForeignKey("route_segments.id"))
    sequence = Column(Integer)
    
    mode = Column(Enum(TransportMode))
    status = Column(Enum(SegmentStatus), default=SegmentStatus.PENDING)
    
    origin_station_id = Column(Integer, ForeignKey("stations.id"))
    destination_station_id = Column(Integer, ForeignKey("stations.id"))
    is_transit_point = Column(Integer, default=0)
    
    planned_departure = Column(DateTime, nullable=True)
    planned_arrival = Column(DateTime, nullable=True)
    actual_departure = Column(DateTime, nullable=True)
    actual_arrival = Column(DateTime, nullable=True)
    
    assigned_vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    
    order = relationship("Order", back_populates="segments")
    route_segment = relationship("RouteSegment")
    origin_station = relationship("Station", foreign_keys=[origin_station_id])
    destination_station = relationship("Station", foreign_keys=[destination_station_id])
    assigned_vehicle = relationship("Vehicle")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class ContainerTrajectory(Base):
    __tablename__ = "container_trajectories"
    
    id = Column(Integer, primary_key=True, index=True)
    container_id = Column(Integer, ForeignKey("containers.id"))
    order_id = Column(Integer, ForeignKey("orders.id"), nullable=True)
    segment_id = Column(Integer, ForeignKey("segments.id"), nullable=True)
    
    event_type = Column(String(50))
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    
    event_time = Column(DateTime, default=datetime.utcnow)
    description = Column(Text, nullable=True)
    
    container = relationship("Container")
    station = relationship("Station")
    vehicle = relationship("Vehicle")
