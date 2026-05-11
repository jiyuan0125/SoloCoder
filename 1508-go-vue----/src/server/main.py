import os
import uuid
from datetime import datetime
from typing import List, Optional
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import PlainTextResponse

from src.core.models import (
    Origin, 
    Vehicle, 
    VehicleStatus, 
    Order, 
    DeliveryTask, 
    DeliveryTaskStatus,
    TemperatureReport,
    Location
)
from src.core.store import store
from src.core.scheduler import (
    find_best_vehicle,
    assign_task_to_vehicle,
    complete_task,
    update_task_statuses
)
from src.core.exporter import export_delivery_records


app = FastAPI(title="Cold Chain Logistics API", version="1.0.0")


@app.on_event("startup")
async def startup_event():
    pass


@app.get("/")
async def root():
    return {"message": "Cold Chain Logistics API", "version": "1.0.0"}


@app.get("/origins", response_model=List[Origin])
async def list_origins():
    return store.list_origins()


@app.post("/origins", response_model=Origin)
async def create_origin(origin: Origin):
    if store.get_origin(origin.id):
        raise HTTPException(status_code=400, detail="Origin with this ID already exists")
    return store.add_origin(origin)


@app.get("/origins/{origin_id}", response_model=Origin)
async def get_origin(origin_id: str):
    origin = store.get_origin(origin_id)
    if not origin:
        raise HTTPException(status_code=404, detail="Origin not found")
    return origin


@app.get("/vehicles", response_model=List[Vehicle])
async def list_vehicles():
    return store.list_vehicles()


@app.post("/vehicles", response_model=Vehicle)
async def create_vehicle(vehicle: Vehicle):
    if store.get_vehicle(vehicle.id):
        raise HTTPException(status_code=400, detail="Vehicle with this ID already exists")
    return store.add_vehicle(vehicle)


@app.get("/vehicles/{vehicle_id}", response_model=Vehicle)
async def get_vehicle(vehicle_id: str):
    vehicle = store.get_vehicle(vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    return vehicle


@app.put("/vehicles/{vehicle_id}/report")
async def report_vehicle_status(
    vehicle_id: str,
    temperature: float,
    latitude: float,
    longitude: float
):
    vehicle = store.get_vehicle(vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    
    vehicle.current_temperature = temperature
    vehicle.location = Location(latitude=latitude, longitude=longitude)
    vehicle.last_report_time = datetime.now()
    
    for task_id in vehicle.current_task_ids:
        task = store.get_task(task_id)
        if task and task.status != DeliveryTaskStatus.COMPLETED:
            report = TemperatureReport(
                timestamp=datetime.now(),
                temperature=temperature,
                location=Location(latitude=latitude, longitude=longitude)
            )
            task.temperature_reports.append(report)
            if task.status == DeliveryTaskStatus.ASSIGNED:
                task.status = DeliveryTaskStatus.IN_TRANSIT
            store.update_task(task)
    
    store.update_vehicle(vehicle)
    update_task_statuses()
    
    return {"message": "Status reported successfully"}


@app.get("/orders", response_model=List[Order])
async def list_orders():
    return store.list_orders()


@app.post("/orders", response_model=Order)
async def create_order(order: Order):
    if store.get_order(order.id):
        raise HTTPException(status_code=400, detail="Order with this ID already exists")
    
    if not store.get_origin(order.origin_id):
        raise HTTPException(status_code=400, detail="Origin not found")
    
    return store.add_order(order)


@app.get("/orders/{order_id}", response_model=Order)
async def get_order(order_id: str):
    order = store.get_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    return order


@app.post("/orders/{order_id}/create-task", response_model=DeliveryTask)
async def create_task_from_order(order_id: str):
    order = store.get_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    task_id = f"task-{uuid.uuid4().hex[:8]}"
    task = DeliveryTask(
        id=task_id,
        order_id=order.id,
        origin_id=order.origin_id,
        destination=order.destination,
        destination_address=order.destination_address,
        weight=order.weight,
        min_temperature=order.min_temperature,
        max_temperature=order.max_temperature,
        created_at=datetime.now()
    )
    
    return store.add_task(task)


@app.get("/tasks", response_model=List[DeliveryTask])
async def list_tasks(status: Optional[DeliveryTaskStatus] = None):
    tasks = store.list_tasks()
    if status:
        tasks = [t for t in tasks if t.status == status]
    return tasks


@app.get("/tasks/{task_id}", response_model=DeliveryTask)
async def get_task(task_id: str):
    task = store.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task


@app.post("/tasks/{task_id}/assign")
async def assign_task(task_id: str):
    task = store.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    
    if task.status != DeliveryTaskStatus.PENDING:
        raise HTTPException(status_code=400, detail="Task is not in pending status")
    
    vehicle = find_best_vehicle(task)
    if not vehicle:
        raise HTTPException(status_code=404, detail="No available vehicle found")
    
    success = assign_task_to_vehicle(task, vehicle)
    if not success:
        raise HTTPException(status_code=400, detail="Failed to assign task to vehicle")
    
    return {"message": "Task assigned successfully", "vehicle_id": vehicle.id}


@app.post("/tasks/{task_id}/complete")
async def complete_delivery_task(task_id: str):
    task = store.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    
    if task.status == DeliveryTaskStatus.COMPLETED:
        raise HTTPException(status_code=400, detail="Task already completed")
    
    if not task.vehicle_id:
        raise HTTPException(status_code=400, detail="Task is not assigned to any vehicle")
    
    success = complete_task(task)
    if not success:
        raise HTTPException(status_code=400, detail="Failed to complete task")
    
    return {"message": "Task completed successfully"}


@app.get("/tasks/{task_id}/temperature-reports", response_model=List[TemperatureReport])
async def get_task_temperature_reports(task_id: str):
    task = store.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task.temperature_reports


@app.get("/export/delivery-records", response_class=PlainTextResponse)
async def export_records(
    start_date: str = Query(..., description="Start date in YYYY-MM-DD format"),
    end_date: str = Query(..., description="End date in YYYY-MM-DD format")
):
    try:
        start_dt = datetime.strptime(start_date, "%Y-%m-%d")
        end_dt = datetime.strptime(end_date, "%Y-%m-%d")
        end_dt = end_dt.replace(hour=23, minute=59, second=59)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid date format. Please use YYYY-MM-DD")
    
    if start_dt > end_dt:
        raise HTTPException(status_code=400, detail="Start date must be before end date")
    
    update_task_statuses()
    
    table = export_delivery_records(start_dt, end_dt)
    return table


def run_server():
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    run_server()
