from datetime import datetime
from typing import List, Optional

from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import PlainTextResponse

from core import DataStore, DeliveryTask, Location, Scheduler, Vehicle
from core.models import TaskStatus


def create_app() -> FastAPI:
    app = FastAPI(title="冷链物流服务")
    store = DataStore()
    scheduler = Scheduler(store)

    @app.on_event("startup")
    async def seed_data() -> None:
        if not store.list_locations():
            store.add_location(
                Location(
                    id="loc_1",
                    name="北京产地",
                    latitude=39.9042,
                    longitude=116.4074,
                    is_origin=True,
                )
            )
            store.add_location(
                Location(
                    id="loc_2",
                    name="上海产地",
                    latitude=31.2304,
                    longitude=121.4737,
                    is_origin=True,
                )
            )
            store.add_location(
                Location(
                    id="loc_3",
                    name="广州配送点",
                    latitude=23.1291,
                    longitude=113.2644,
                    is_origin=False,
                )
            )
            store.add_location(
                Location(
                    id="loc_4",
                    name="深圳配送点",
                    latitude=22.5431,
                    longitude=114.0579,
                    is_origin=False,
                )
            )

        if not store.list_vehicles():
            store.add_vehicle(
                Vehicle(
                    id="v_1",
                    vehicle_number="冷链001",
                    current_temperature=-18.0,
                    max_load=5000.0,
                    current_location_id="loc_1",
                )
            )
            store.add_vehicle(
                Vehicle(
                    id="v_2",
                    vehicle_number="冷链002",
                    current_temperature=4.0,
                    max_load=3000.0,
                    current_location_id="loc_2",
                )
            )
            store.add_vehicle(
                Vehicle(
                    id="v_3",
                    vehicle_number="冷链003",
                    current_temperature=-5.0,
                    max_load=4000.0,
                    current_location_id="loc_1",
                )
            )

    @app.get("/health")
    async def health() -> dict:
        return {"status": "ok"}

    @app.get("/locations")
    async def list_locations() -> List[Location]:
        return store.list_locations()

    @app.post("/locations")
    async def create_location(location: Location) -> Location:
        return store.add_location(location)

    @app.get("/vehicles")
    async def list_vehicles() -> List[Vehicle]:
        return store.list_vehicles()

    @app.post("/vehicles")
    async def create_vehicle(vehicle: Vehicle) -> Vehicle:
        return store.add_vehicle(vehicle)

    @app.get("/vehicles/{vehicle_id}")
    async def get_vehicle(vehicle_id: str) -> Vehicle:
        vehicle = store.get_vehicle(vehicle_id)
        if not vehicle:
            raise HTTPException(status_code=404, detail="Vehicle not found")
        return vehicle

    @app.get("/tasks")
    async def list_tasks(
        status: Optional[TaskStatus] = Query(None),
    ) -> List[DeliveryTask]:
        return store.list_tasks(status)

    @app.post("/tasks")
    async def create_task(task: DeliveryTask) -> DeliveryTask:
        return store.add_task(task)

    @app.get("/tasks/{task_id}")
    async def get_task(task_id: str) -> DeliveryTask:
        task = store.get_task(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")
        return task

    @app.post("/tasks/{task_id}/assign")
    async def assign_task(task_id: str) -> dict:
        task = store.get_task(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")

        if task.status != TaskStatus.PENDING:
            raise HTTPException(status_code=400, detail="Task already assigned")

        vehicle = scheduler.assign_task(task)
        if not vehicle:
            raise HTTPException(status_code=400, detail="No eligible vehicle available")

        return {
            "task_id": task.id,
            "vehicle_id": vehicle.id,
            "vehicle_number": vehicle.vehicle_number,
            "status": TaskStatus.ASSIGNED,
        }

    @app.post("/tasks/{task_id}/start")
    async def start_task(task_id: str) -> DeliveryTask:
        try:
            return scheduler.start_task(task_id)
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/tasks/{task_id}/complete")
    async def complete_task(task_id: str) -> DeliveryTask:
        task = store.get_task(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")

        if task.status not in (
            TaskStatus.IN_TRANSIT,
            TaskStatus.TEMPERATURE_ABNORMAL,
            TaskStatus.COMMUNICATION_LOST,
        ):
            raise HTTPException(status_code=400, detail="Task not in transit")

        return store.complete_task(task_id)

    @app.post("/tasks/{task_id}/report/temperature")
    async def report_temperature(
        task_id: str, vehicle_id: str, temperature: float
    ) -> dict:
        task = store.get_task(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")

        if task.assigned_vehicle_id != vehicle_id:
            raise HTTPException(status_code=400, detail="Vehicle not assigned to this task")

        report = store.add_temperature_report(task_id, vehicle_id, temperature)
        return {
            "task_id": report.task_id,
            "vehicle_id": report.vehicle_id,
            "temperature": report.temperature,
            "is_normal": report.is_normal,
            "reported_at": report.reported_at,
        }

    @app.post("/tasks/{task_id}/report/location")
    async def report_location(
        task_id: str, vehicle_id: str, latitude: float, longitude: float
    ) -> dict:
        task = store.get_task(task_id)
        if not task:
            raise HTTPException(status_code=404, detail="Task not found")

        if task.assigned_vehicle_id != vehicle_id:
            raise HTTPException(status_code=400, detail="Vehicle not assigned to this task")

        report = store.add_location_report(task_id, vehicle_id, latitude, longitude)
        return {
            "task_id": report.task_id,
            "vehicle_id": report.vehicle_id,
            "latitude": report.latitude,
            "longitude": report.longitude,
            "reported_at": report.reported_at,
        }

    @app.post("/system/check-communication")
    async def check_communication() -> List[dict]:
        affected = store.check_communication_status()
        return [
            {
                "task_id": t.id,
                "order_id": t.order_id,
                "vehicle_id": t.assigned_vehicle_id,
                "status": t.status,
            }
            for t in affected
        ]

    @app.get("/export/delivery-records", response_class=PlainTextResponse)
    async def export_delivery_records(
        start_date: str = Query(..., description="开始日期 (YYYY-MM-DD)"),
        end_date: str = Query(..., description="结束日期 (YYYY-MM-DD)"),
    ) -> str:
        try:
            start_dt = datetime.strptime(start_date, "%Y-%m-%d")
            end_dt = datetime.strptime(end_date, "%Y-%m-%d")
            end_dt = end_dt.replace(hour=23, minute=59, second=59)
        except ValueError:
            raise HTTPException(
                status_code=400,
                detail="日期格式错误，请使用 YYYY-MM-DD",
            )

        records = store.get_delivery_records(start_dt, end_dt)

        if not records:
            return "指定日期范围内暂无配送记录。\n"

        headers = [
            "任务ID",
            "订单ID",
            "车辆编号",
            "起点",
            "终点",
            "重量(kg)",
            "温度范围(℃)",
            "创建时间",
            "开始时间",
            "完成时间",
            "最终状态",
        ]

        col_widths = [12, 12, 12, 14, 14, 10, 14, 22, 22, 22, 16]

        separator = "+" + "+".join("-" * w for w in col_widths) + "+\n"

        def format_row(values: list) -> str:
            return "|" + "|".join(
                f"{str(v)[:w-2]:^{w-2}}" for v, w in zip(values, col_widths)
            ) + "|\n"

        output = separator
        output += format_row(headers)
        output += separator

        for r in records:
            temp_range = f"{r.required_min_temp:.1f}~{r.required_max_temp:.1f}"
            created = r.created_at.strftime("%Y-%m-%d %H:%M:%S")
            started = r.started_at.strftime("%Y-%m-%d %H:%M:%S") if r.started_at else "-"
            completed = r.completed_at.strftime("%Y-%m-%d %H:%M:%S") if r.completed_at else "-"

            output += format_row([
                r.task_id,
                r.order_id,
                r.vehicle_number,
                r.origin_name,
                r.destination_name,
                f"{r.weight:.1f}",
                temp_range,
                created,
                started,
                completed,
                r.final_status.value,
            ])
        output += separator

        output += f"\n共 {len(records)} 条配送记录\n"
        output += f"日期范围: {start_date} 至 {end_date}\n"

        return output

    return app
