import os
import asyncio
from contextlib import asynccontextmanager
from typing import List, Optional
from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel

from core import SystemState, Scheduler, MetricsCalculator, RiderStatus, OrderStatus


state = SystemState()
scheduler = Scheduler(state)
metrics = MetricsCalculator(state)


class RiderCreateRequest(BaseModel):
    rider_id: str


class OrderCreateRequest(BaseModel):
    order_id: str
    merchant_name: str


class OrderAssignRequest(BaseModel):
    rider_id: str


@asynccontextmanager
async def lifespan(app: FastAPI):
    scheduler_task = asyncio.create_task(run_scheduler())
    try:
        yield
    finally:
        scheduler_task.cancel()


async def run_scheduler():
    while True:
        scheduler.tick()
        await asyncio.sleep(1)


def create_app() -> FastAPI:
    app = FastAPI(title="外卖配送调度系统", lifespan=lifespan)

    @app.get("/health")
    async def health():
        return {"status": "ok"}

    @app.get("/metrics")
    async def get_metrics():
        return metrics.get_all_metrics()

    @app.post("/riders")
    async def add_rider(request: RiderCreateRequest):
        try:
            rider = state.add_rider(request.rider_id)
            return rider.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/riders")
    async def list_riders():
        return [rider.model_dump() for rider in state.get_all_riders()]

    @app.get("/riders/{rider_id}")
    async def get_rider(rider_id: str):
        rider = state.get_rider(rider_id)
        if not rider:
            raise HTTPException(status_code=404, detail="Rider not found")
        return rider.model_dump()

    @app.post("/riders/{rider_id}/status")
    async def update_rider_status(rider_id: str, status: RiderStatus):
        try:
            rider = state.update_rider_status(rider_id, status)
            return rider.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/riders/{rider_id}/position")
    async def update_rider_position(rider_id: str):
        try:
            rider = state.update_rider_position(rider_id)
            return rider.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/riders/{rider_id}/back-online")
    async def rider_back_online(rider_id: str):
        try:
            state.handle_rider_back_online(rider_id)
            rider = state.get_rider(rider_id)
            return rider.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/orders")
    async def add_order(request: OrderCreateRequest, background_tasks: BackgroundTasks):
        try:
            order = state.add_order(request.order_id, request.merchant_name)
            background_tasks.add_task(scheduler.tick)
            return order.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/orders")
    async def list_orders(status: Optional[OrderStatus] = None):
        if status:
            orders = state.get_orders_by_status(status)
        else:
            orders = state.get_all_orders()
        return [order.model_dump() for order in orders]

    @app.get("/orders/{order_id}")
    async def get_order(order_id: str):
        order = state.get_order(order_id)
        if not order:
            raise HTTPException(status_code=404, detail="Order not found")
        return order.model_dump()

    @app.post("/orders/{order_id}/complete")
    async def complete_order(order_id: str, background_tasks: BackgroundTasks):
        try:
            state.complete_order(order_id)
            background_tasks.add_task(scheduler.tick)
            order = state.get_order(order_id)
            return order.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/orders/{order_id}/assign")
    async def manual_assign_order(order_id: str, request: OrderAssignRequest, background_tasks: BackgroundTasks):
        try:
            state.assign_order_to_rider(order_id, request.rider_id)
            background_tasks.add_task(scheduler.tick)
            order = state.get_order(order_id)
            return order.model_dump()
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/scheduler/tick")
    async def manual_tick():
        return scheduler.tick()

    return app
