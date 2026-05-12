from datetime import datetime
from enum import Enum as PyEnum
from typing import Optional, List
from sqlmodel import SQLModel, Field, Relationship


class CarType(str, PyEnum):
    OPEN = "敞车"
    BOX = "棚车"
    TANK = "罐车"
    FLAT = "平车"
    REFRIGERATED = "冷藏车"


class CargoType(str, PyEnum):
    BULK = "散装货物"
    BOXED = "箱装货物"
    LIQUID = "液体货物"
    CONTAINER = "集装箱"
    PERISHABLE = "易腐货物"


class OrderStatus(str, PyEnum):
    CREATED = "已录单"
    ACCEPTED = "已受理"
    ASSIGNED = "已配车"
    LOADED = "已装车"
    DEPARTED = "已发运"
    ARRIVED = "已到达"
    DELIVERED = "已交付"
    CANCELLED = "已取消"


CARGO_CAR_MAPPING = {
    CargoType.BULK: [CarType.OPEN],
    CargoType.BOXED: [CarType.BOX],
    CargoType.LIQUID: [CarType.TANK],
    CargoType.CONTAINER: [CarType.FLAT],
    CargoType.PERISHABLE: [CarType.REFRIGERATED],
}

CARGO_NAME_MAPPING = {
    "散装货物": CargoType.BULK,
    "箱装货物": CargoType.BOXED,
    "液体货物": CargoType.LIQUID,
    "集装箱": CargoType.CONTAINER,
    "易腐货物": CargoType.PERISHABLE,
}

CAR_TYPE_NAME_MAPPING = {
    "敞车": CarType.OPEN,
    "棚车": CarType.BOX,
    "罐车": CarType.TANK,
    "平车": CarType.FLAT,
    "冷藏车": CarType.REFRIGERATED,
}


def is_cargo_compatible(cargo_type: CargoType, car_type: CarType) -> bool:
    return car_type in CARGO_CAR_MAPPING.get(cargo_type, [])


class Wagon(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    wagon_number: str = Field(index=True, unique=True)
    car_type: CarType
    capacity_tons: float
    current_station: str = Field(index=True)
    status: str = Field(default="空车")
    current_train_id: Optional[int] = Field(default=None, foreign_key="train.id")


class Station(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    code: str = Field(index=True, unique=True)
    name: str
    line_traction_tons: float


class FreightOrder(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    order_no: str = Field(index=True, unique=True)
    cargo_type: CargoType
    cargo_name: str
    weight_tons: float
    origin_station: str
    dest_station: str
    status: OrderStatus = Field(default=OrderStatus.CREATED)
    assigned_wagon_id: Optional[int] = Field(default=None, foreign_key="wagon.id")
    train_id: Optional[int] = Field(default=None, foreign_key="train.id")
    
    created_at: datetime = Field(default_factory=datetime.now)
    accepted_at: Optional[datetime] = None
    assigned_at: Optional[datetime] = None
    loaded_at: Optional[datetime] = None
    departed_at: Optional[datetime] = None
    arrived_at: Optional[datetime] = None
    delivered_at: Optional[datetime] = None
    cancelled_at: Optional[datetime] = None


class Train(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    train_no: str = Field(index=True, unique=True)
    origin_station: str
    dest_station: str
    line_traction_tons: float
    max_wagons: int = Field(default=50)
    
    wagons: List["TrainWagon"] = Relationship(back_populates="train")
    
    departure_time: Optional[datetime] = None
    arrival_time: Optional[datetime] = None


class TrainWagon(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    train_id: int = Field(foreign_key="train.id")
    wagon_id: int = Field(foreign_key="wagon.id")
    position: int
    order_id: Optional[int] = Field(default=None, foreign_key="freightorder.id")
    
    train: Train = Relationship(back_populates="wagons")


class TrainArrivalDeparture(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    train_no: str = Field(index=True)
    station_code: str = Field(index=True)
    event_type: str
    event_time: datetime = Field(default_factory=datetime.now)
