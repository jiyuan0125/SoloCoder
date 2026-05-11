from typing import List, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import asyncio

from core import (
    Rider,
    Order,
    Metrics,
    RiderStatus,
    OrderStatus,
    create_order,
    create_rider,
    update_rider_position,
    set_rider_status,
    accept_order,
    complete_order,
    manually_reassign_order,
    dispatch_orders,
    check_timeouts,
    check_offline_riders,
    get_metrics,
    get_rider,
    get_all_riders,
    get_order,
    get_all_orders,
)


class CreateRiderRequest(BaseModel):
    rider_id: str


class SetRiderStatusRequest(BaseModel):
    status: RiderStatus


class AcceptOrderRequest(BaseModel):
    order_id: str


class ReassignOrderRequest(BaseModel):
    rider_id: str


async def background_tasks():
    while True:
        dispatch_orders()
        check_timeouts()
        check_offline_riders()
        await asyncio.sleep(5)


@asynccontextmanager
async def lifespan(app: FastAPI):
    task = asyncio.create_task(background_tasks())
    yield
    task.cancel()
    try:
        await task
    except asyncio.CancelledError:
        pass


app = FastAPI(title="配送调度系统", lifespan=lifespan)


@app.post("/riders", response_model=Rider)
def api_create_rider(request: CreateRiderRequest):
    rider = get_rider(request.rider_id)
    if rider:
        raise HTTPException(status_code=400, detail="骑手已存在")
    return create_rider(request.rider_id)


@app.get("/riders", response_model=List[Rider])
def api_list_riders(status: Optional[RiderStatus] = None):
    if status:
        return [r for r in get_all_riders() if r.status == status]
    return get_all_riders()


@app.get("/riders/{rider_id}", response_model=Rider)
def api_get_rider(rider_id: str):
    rider = get_rider(rider_id)
    if not rider:
        raise HTTPException(status_code=404, detail="骑手不存在")
    return rider


@app.post("/riders/{rider_id}/heartbeat", response_model=Rider)
def api_heartbeat(rider_id: str):
    rider = update_rider_position(rider_id)
    if not rider:
        raise HTTPException(status_code=404, detail="骑手不存在")
    return rider


@app.post("/riders/{rider_id}/status", response_model=Rider)
def api_set_rider_status(rider_id: str, request: SetRiderStatusRequest):
    rider = get_rider(rider_id)
    if not rider:
        raise HTTPException(status_code=404, detail="骑手不存在")
    updated = set_rider_status(rider_id, request.status)
    if not updated:
        raise HTTPException(status_code=400, detail="状态更新失败")
    return updated


@app.post("/riders/{rider_id}/accept-order", response_model=Order)
def api_accept_order(rider_id: str, request: AcceptOrderRequest):
    rider = get_rider(rider_id)
    if not rider:
        raise HTTPException(status_code=404, detail="骑手不存在")
    order = accept_order(rider_id, request.order_id)
    if not order:
        raise HTTPException(status_code=400, detail="接单失败")
    return order


@app.post("/orders", response_model=Order)
def api_create_order():
    return create_order()


@app.get("/orders", response_model=List[Order])
def api_list_orders(status: Optional[OrderStatus] = None):
    if status:
        return [o for o in get_all_orders() if o.status == status]
    return get_all_orders()


@app.get("/orders/{order_id}", response_model=Order)
def api_get_order(order_id: str):
    order = get_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="订单不存在")
    return order


@app.post("/orders/{order_id}/complete", response_model=Order)
def api_complete_order(order_id: str):
    order = complete_order(order_id)
    if not order:
        raise HTTPException(status_code=400, detail="完成订单失败")
    return order


@app.post("/orders/{order_id}/reassign", response_model=Order)
def api_reassign_order(order_id: str, request: ReassignOrderRequest):
    order = get_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="订单不存在")
    if not order.needs_manual_reassign:
        raise HTTPException(status_code=400, detail="该订单不需要手动重新分配")
    updated = manually_reassign_order(order_id, request.rider_id)
    if not updated:
        raise HTTPException(status_code=400, detail="重新分配失败")
    return updated


@app.get("/metrics", response_model=Metrics)
def api_get_metrics():
    return get_metrics()
