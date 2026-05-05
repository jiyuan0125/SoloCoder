from datetime import datetime, timedelta, timezone
from typing import Optional, List, AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI, Depends
from fastapi.responses import JSONResponse

from shared.models import (
    CrossDockOrderCreate,
    CrossDockOrderResponse,
    InboundScanRequest,
    OutboundScanRequest,
    BatchCrossDockCreate,
    BatchCrossDockResponse,
    EfficiencyStatistics,
    DailyReport,
    ZoneMonitorResponse,
    ExceptionPlan,
)
from shared.errors import ErrorCode, ErrorMessage
from server.services.cross_dock_service import CrossDockService
from server.repositories.in_memory_repository import InMemoryRepository
from server.exceptions import (
    OrderNotFoundException,
    InvalidStatusTransitionException,
    SameOperatorException,
    ItemMismatchException,
    TimeoutException,
    CrossDayException,
    BatchNotReadyException,
    DuplicateOrderException,
    OutboundOrderNotFoundException,
    ExceptionPlanNotFoundException,
)
from server.middleware.timeout_monitor import TimeoutMonitor


repository: Optional[InMemoryRepository] = None
service: Optional[CrossDockService] = None
timeout_monitor: Optional[TimeoutMonitor] = None


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    global repository, service, timeout_monitor
    repository = InMemoryRepository()
    service = CrossDockService(repository)
    timeout_monitor = TimeoutMonitor(service)
    timeout_monitor.start()
    yield
    if timeout_monitor:
        timeout_monitor.stop()


app = FastAPI(
    title="越库管理API",
    description="越库作业管理系统API",
    version="0.1.0",
    lifespan=lifespan,
)


def get_service() -> CrossDockService:
    if service is None:
        raise RuntimeError("Service not initialized")
    return service


@app.exception_handler(OrderNotFoundException)
async def order_not_found_exception_handler(
    request: object, exc: OrderNotFoundException
) -> JSONResponse:
    return JSONResponse(
        status_code=404,
        content={
            "error_code": ErrorCode.ORDER_NOT_FOUND,
            "message": ErrorMessage.ORDER_NOT_FOUND,
            "details": exc.details,
        },
    )


@app.exception_handler(InvalidStatusTransitionException)
async def invalid_status_transition_exception_handler(
    request: object, exc: InvalidStatusTransitionException
) -> JSONResponse:
    return JSONResponse(
        status_code=400,
        content={
            "error_code": ErrorCode.INVALID_STATUS_TRANSITION,
            "message": ErrorMessage.INVALID_STATUS_TRANSITION,
            "details": exc.details,
        },
    )


@app.exception_handler(SameOperatorException)
async def same_operator_exception_handler(
    request: object, exc: SameOperatorException
) -> JSONResponse:
    return JSONResponse(
        status_code=400,
        content={
            "error_code": ErrorCode.SAME_OPERATOR_FOR_INBOUND_OUTBOUND,
            "message": ErrorMessage.SAME_OPERATOR_FOR_INBOUND_OUTBOUND,
            "details": exc.details,
        },
    )


@app.exception_handler(ItemMismatchException)
async def item_mismatch_exception_handler(
    request: object, exc: ItemMismatchException
) -> JSONResponse:
    return JSONResponse(
        status_code=400,
        content={
            "error_code": ErrorCode.ITEM_MISMATCH,
            "message": ErrorMessage.ITEM_MISMATCH,
            "details": exc.details,
            "validation_result": (
                exc.validation_result.model_dump() if exc.validation_result else None
            ),
        },
    )


@app.exception_handler(TimeoutException)
async def timeout_exception_handler(
    request: object, exc: TimeoutException
) -> JSONResponse:
    return JSONResponse(
        status_code=400,
        content={
            "error_code": ErrorCode.TIMEOUT_OCCURRED,
            "message": ErrorMessage.TIMEOUT_OCCURRED,
            "details": exc.details,
        },
    )


@app.exception_handler(CrossDayException)
async def cross_day_exception_handler(
    request: object, exc: CrossDayException
) -> JSONResponse:
    return JSONResponse(
        status_code=400,
        content={
            "error_code": ErrorCode.CROSS_DAY_VIOLATION,
            "message": ErrorMessage.CROSS_DAY_VIOLATION,
            "details": exc.details,
        },
    )


@app.exception_handler(BatchNotReadyException)
async def batch_not_ready_exception_handler(
    request: object, exc: BatchNotReadyException
) -> JSONResponse:
    return JSONResponse(
        status_code=400,
        content={
            "error_code": ErrorCode.BATCH_NOT_READY,
            "message": ErrorMessage.BATCH_NOT_READY,
            "details": exc.details,
        },
    )


@app.exception_handler(DuplicateOrderException)
async def duplicate_order_exception_handler(
    request: object, exc: DuplicateOrderException
) -> JSONResponse:
    return JSONResponse(
        status_code=409,
        content={
            "error_code": ErrorCode.DUPLICATE_ORDER,
            "message": ErrorMessage.DUPLICATE_ORDER,
            "details": exc.details,
        },
    )


@app.exception_handler(OutboundOrderNotFoundException)
async def outbound_order_not_found_exception_handler(
    request: object, exc: OutboundOrderNotFoundException
) -> JSONResponse:
    return JSONResponse(
        status_code=404,
        content={
            "error_code": ErrorCode.OUTBOUND_ORDER_NOT_FOUND,
            "message": ErrorMessage.OUTBOUND_ORDER_NOT_FOUND,
            "details": exc.details,
        },
    )


@app.exception_handler(ExceptionPlanNotFoundException)
async def exception_plan_not_found_exception_handler(
    request: object, exc: ExceptionPlanNotFoundException
) -> JSONResponse:
    return JSONResponse(
        status_code=404,
        content={
            "error_code": ErrorCode.EXCEPTION_PLAN_NOT_FOUND,
            "message": ErrorMessage.EXCEPTION_PLAN_NOT_FOUND,
            "details": exc.details,
        },
    )


@app.post("/api/cross-dock/orders", response_model=CrossDockOrderResponse)
def create_order(
    order_create: CrossDockOrderCreate,
    service: CrossDockService = Depends(get_service),
) -> CrossDockOrderResponse:
    order = service.create_order(order_create)
    return service.to_response(order)


@app.get("/api/cross-dock/orders/{order_id}", response_model=CrossDockOrderResponse)
def get_order(
    order_id: str,
    service: CrossDockService = Depends(get_service),
) -> CrossDockOrderResponse:
    order = service.get_order(order_id)
    if order is None:
        raise OrderNotFoundException(order_id=order_id)
    return service.to_response(order)


@app.get("/api/cross-dock/orders", response_model=List[CrossDockOrderResponse])
def list_orders(
    status: Optional[str] = None,
    warehouse_id: Optional[str] = None,
    service: CrossDockService = Depends(get_service),
) -> List[CrossDockOrderResponse]:
    orders = service.get_all_orders()
    if status:
        orders = [o for o in orders if o.status.value == status]
    if warehouse_id:
        orders = [o for o in orders if o.warehouse_id == warehouse_id]
    return [service.to_response(o) for o in orders]


@app.post("/api/cross-dock/inbound-scan", response_model=CrossDockOrderResponse)
def inbound_scan(
    request: InboundScanRequest,
    service: CrossDockService = Depends(get_service),
) -> CrossDockOrderResponse:
    return service.inbound_scan(request)


@app.post("/api/cross-dock/outbound-scan", response_model=CrossDockOrderResponse)
def outbound_scan(
    request: OutboundScanRequest,
    service: CrossDockService = Depends(get_service),
) -> CrossDockOrderResponse:
    return service.outbound_scan(request)


@app.post("/api/cross-dock/batch", response_model=BatchCrossDockResponse)
def create_batch_cross_dock(
    batch_create: BatchCrossDockCreate,
    service: CrossDockService = Depends(get_service),
) -> BatchCrossDockResponse:
    return service.create_batch_cross_dock(batch_create)


@app.get("/api/cross-dock/batch/{outbound_order_number}", response_model=BatchCrossDockResponse)
def get_batch_status(
    outbound_order_number: str,
    service: CrossDockService = Depends(get_service),
) -> BatchCrossDockResponse:
    return service.get_batch_status(outbound_order_number)


@app.get("/api/cross-dock/statistics/efficiency", response_model=EfficiencyStatistics)
def get_efficiency_statistics(
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    service: CrossDockService = Depends(get_service),
) -> EfficiencyStatistics:
    if end_time is None:
        end_time = datetime.now(timezone.utc)
    if start_time is None:
        start_time = end_time - timedelta(days=7)
    return service.get_efficiency_statistics(start_time, end_time)


@app.get("/api/cross-dock/reports/daily", response_model=DailyReport)
def get_daily_report(
    report_date: Optional[datetime] = None,
    service: CrossDockService = Depends(get_service),
) -> DailyReport:
    if report_date is None:
        report_date = datetime.now(timezone.utc)
    return service.generate_daily_report(report_date)


@app.get("/api/cross-dock/zone-monitor/{warehouse_id}", response_model=ZoneMonitorResponse)
def get_zone_monitor(
    warehouse_id: str,
    service: CrossDockService = Depends(get_service),
) -> ZoneMonitorResponse:
    return service.get_zone_monitor(warehouse_id)


@app.get("/api/cross-dock/exception-plans/{exception_type}", response_model=ExceptionPlan)
def get_exception_plan(
    exception_type: str,
    service: CrossDockService = Depends(get_service),
) -> ExceptionPlan:
    return service.get_exception_plan(exception_type)


@app.get("/api/health")
def health_check() -> dict[str, str]:
    return {"status": "healthy", "timestamp": datetime.now(timezone.utc).isoformat()}
