import os
import uuid
from typing import List, Optional
from datetime import datetime, date

from fastapi import FastAPI, HTTPException, Query, Body
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from src.core import (
    Store,
    Pond, WaterType,
    Species,
    BreedingCycle, BreedingStage,
    SalesOrder,
    DeliveryVehicle, DeliveryTask, DeliveryStatus,
    TodoItem, TodoPriority, TodoStatus,
    InventoryItem,
    PondTemperatureLog
)


app = FastAPI(
    title="水产种苗场管理系统",
    description="繁育和销售管理后端系统",
    version="0.1.0"
)

store = Store()


def generate_id() -> str:
    return str(uuid.uuid4())


class PondCreate(BaseModel):
    name: str
    water_type: WaterType
    capacity: float
    species_id: Optional[str] = None


class PondUpdate(BaseModel):
    name: Optional[str] = None
    water_type: Optional[WaterType] = None
    capacity: Optional[float] = None
    species_id: Optional[str] = None


class TemperatureRecord(BaseModel):
    temperature: float


class SpeciesCreate(BaseModel):
    name: str
    min_temperature: float
    max_temperature: float
    description: Optional[str] = None


class BreedingCycleCreate(BaseModel):
    pond_id: str
    species_id: str
    spawning_date: date
    incubation_date: Optional[date] = None
    emergence_date: Optional[date] = None
    expected_quantity: Optional[int] = None
    actual_quantity: Optional[int] = None
    notes: Optional[str] = None


class BreedingCycleUpdate(BaseModel):
    incubation_date: Optional[date] = None
    emergence_date: Optional[date] = None
    expected_quantity: Optional[int] = None
    actual_quantity: Optional[int] = None
    notes: Optional[str] = None


class SalesOrderCreate(BaseModel):
    customer_name: str
    customer_phone: Optional[str] = None
    species_id: str
    requested_quantity: int
    unit_price: float
    notes: Optional[str] = None


class DeliveryVehicleCreate(BaseModel):
    license_plate: str
    name: str
    capacity: Optional[int] = None


class DeliveryTaskCreate(BaseModel):
    order_id: str
    vehicle_id: str
    driver_name: str
    driver_phone: Optional[str] = None
    delivery_address: str
    scheduled_start: datetime
    scheduled_end: Optional[datetime] = None
    notes: Optional[str] = None


class DeliveryTaskUpdate(BaseModel):
    actual_start: Optional[datetime] = None
    actual_end: Optional[datetime] = None
    status: Optional[DeliveryStatus] = None
    notes: Optional[str] = None


class TodoUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    priority: Optional[TodoPriority] = None
    status: Optional[TodoStatus] = None


@app.get("/health")
def health_check():
    return {"status": "ok", "message": "服务运行正常"}


@app.get("/species", response_model=List[Species])
def list_species():
    return store.get_all_species()


@app.post("/species", response_model=Species)
def create_species(data: SpeciesCreate):
    species = Species(
        id=generate_id(),
        **data.dict()
    )
    return store.create_species(species)


@app.get("/species/{species_id}", response_model=Species)
def get_species(species_id: str):
    species = store.get_species(species_id)
    if not species:
        raise HTTPException(status_code=404, detail="品种不存在")
    return species


@app.get("/ponds", response_model=List[Pond])
def list_ponds():
    return store.get_all_ponds()


@app.post("/ponds", response_model=Pond)
def create_pond(data: PondCreate):
    pond = Pond(
        id=generate_id(),
        **data.dict()
    )
    return store.create_pond(pond)


@app.get("/ponds/{pond_id}", response_model=Pond)
def get_pond(pond_id: str):
    pond = store.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="繁育池不存在")
    return pond


@app.patch("/ponds/{pond_id}", response_model=Pond)
def update_pond(pond_id: str, data: PondUpdate):
    update_data = {k: v for k, v in data.dict().items() if v is not None}
    pond = store.update_pond(pond_id, update_data)
    if not pond:
        raise HTTPException(status_code=404, detail="繁育池不存在")
    return pond


@app.post("/ponds/{pond_id}/temperature", response_model=PondTemperatureLog)
def record_temperature(pond_id: str, data: TemperatureRecord):
    pond = store.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="繁育池不存在")
    log = store.record_temperature(pond_id, data.temperature)
    if not log:
        raise HTTPException(status_code=400, detail="温度记录失败")
    return log


@app.get("/temperature-logs", response_model=List[PondTemperatureLog])
def list_temperature_logs(pond_id: Optional[str] = Query(None)):
    return store.get_temperature_logs(pond_id)


@app.get("/breeding", response_model=List[BreedingCycle])
def list_breeding_cycles():
    return store.get_all_breeding_cycles()


@app.post("/breeding", response_model=BreedingCycle)
def create_breeding_cycle(data: BreedingCycleCreate):
    pond = store.get_pond(data.pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="繁育池不存在")
    
    species = store.get_species(data.species_id)
    if not species:
        raise HTTPException(status_code=404, detail="品种不存在")
    
    if data.incubation_date and data.incubation_date < data.spawning_date:
        raise HTTPException(status_code=400, detail="孵化日期不能早于产卵日期")
    
    if data.emergence_date and data.emergence_date < data.spawning_date:
        raise HTTPException(status_code=400, detail="出苗日期不能早于产卵日期")
    
    cycle = BreedingCycle(
        id=generate_id(),
        **data.dict()
    )
    return store.create_breeding_cycle(cycle)


@app.get("/breeding/{cycle_id}", response_model=BreedingCycle)
def get_breeding_cycle(cycle_id: str):
    cycle = store.get_breeding_cycle(cycle_id)
    if not cycle:
        raise HTTPException(status_code=404, detail="繁育周期不存在")
    return cycle


@app.patch("/breeding/{cycle_id}", response_model=BreedingCycle)
def update_breeding_cycle(cycle_id: str, data: BreedingCycleUpdate):
    cycle = store.get_breeding_cycle(cycle_id)
    if not cycle:
        raise HTTPException(status_code=404, detail="繁育周期不存在")
    
    update_data = {k: v for k, v in data.dict().items() if v is not None}
    
    if 'emergence_date' in update_data and update_data['emergence_date'] < cycle.spawning_date:
        raise HTTPException(status_code=400, detail="出苗日期不能早于产卵日期")
    
    if 'incubation_date' in update_data and update_data['incubation_date'] < cycle.spawning_date:
        raise HTTPException(status_code=400, detail="孵化日期不能早于产卵日期")
    
    updated = store.update_breeding_cycle(cycle_id, update_data)
    return updated


@app.get("/inventory", response_model=List[InventoryItem])
def list_inventory():
    return store.get_all_inventory()


@app.post("/inventory/{species_id}/check", response_model=TodoItem)
def check_inventory(species_id: str):
    species = store.get_species(species_id)
    if not species:
        raise HTTPException(status_code=404, detail="品种不存在")
    
    todo = store.check_inventory_and_generate_todo(species_id)
    if todo:
        return todo
    raise HTTPException(status_code=200, detail="库存充足")


@app.get("/sales", response_model=List[SalesOrder])
def list_sales_orders():
    return store.get_all_sales_orders()


@app.post("/sales")
def create_sales_order(data: SalesOrderCreate):
    species = store.get_species(data.species_id)
    if not species:
        raise HTTPException(status_code=404, detail="品种不存在")
    
    order = SalesOrder(
        id=generate_id(),
        **data.dict()
    )
    result = store.create_sales_order(order)
    return result


@app.get("/sales/{order_id}", response_model=SalesOrder)
def get_sales_order(order_id: str):
    order = store.get_sales_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="销售订单不存在")
    return order


@app.get("/vehicles", response_model=List[DeliveryVehicle])
def list_vehicles():
    return store.get_all_delivery_vehicles()


@app.post("/vehicles", response_model=DeliveryVehicle)
def create_vehicle(data: DeliveryVehicleCreate):
    vehicle = DeliveryVehicle(
        id=generate_id(),
        **data.dict()
    )
    return store.create_delivery_vehicle(vehicle)


@app.get("/vehicles/{vehicle_id}", response_model=DeliveryVehicle)
def get_vehicle(vehicle_id: str):
    vehicle = store.get_delivery_vehicle(vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="配送车辆不存在")
    return vehicle


@app.get("/deliveries", response_model=List[DeliveryTask])
def list_delivery_tasks():
    return store.get_all_delivery_tasks()


@app.post("/deliveries")
def create_delivery_task(data: DeliveryTaskCreate):
    order = store.get_sales_order(data.order_id)
    if not order:
        raise HTTPException(status_code=404, detail="销售订单不存在")
    
    vehicle = store.get_delivery_vehicle(data.vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="配送车辆不存在")
    
    task = DeliveryTask(
        id=generate_id(),
        **data.dict()
    )
    result = store.create_delivery_task(task)
    
    if not result['success']:
        raise HTTPException(status_code=400, detail=result['message'])
    
    return result


@app.get("/deliveries/{task_id}", response_model=DeliveryTask)
def get_delivery_task(task_id: str):
    task = store.get_delivery_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="配送任务不存在")
    return task


@app.patch("/deliveries/{task_id}", response_model=DeliveryTask)
def update_delivery_task(task_id: str, data: DeliveryTaskUpdate):
    task = store.get_delivery_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="配送任务不存在")
    
    update_data = {k: v for k, v in data.dict().items() if v is not None}
    updated = store.update_delivery_task(task_id, update_data)
    return updated


@app.get("/todos", response_model=List[TodoItem])
def list_todos():
    return store.get_all_todos()


@app.get("/todos/{todo_id}", response_model=TodoItem)
def get_todo(todo_id: str):
    todo = store.get_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    return todo


@app.patch("/todos/{todo_id}", response_model=TodoItem)
def update_todo(todo_id: str, data: TodoUpdate):
    todo = store.get_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    
    update_data = {k: v for k, v in data.dict().items() if v is not None}
    if 'status' in update_data and update_data['status'] == TodoStatus.RESOLVED:
        update_data['resolved_at'] = datetime.utcnow()
    
    updated = store.update_todo(todo_id, update_data)
    return updated
