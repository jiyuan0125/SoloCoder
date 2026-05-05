from datetime import datetime
from typing import TYPE_CHECKING
from uuid import UUID

from fastapi import APIRouter, FastAPI, HTTPException, Query, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from server import service
from server.config import settings
from server.database import db
from server.exceptions import BusinessException
from shared.enums import ErrorCode, TransferStatus, TransferType
from shared.protocols import (
    InterCompanyApproveRequest,
    InventoryDetailResponse,
    InventoryOverviewResponse,
    LossDetailResponse,
    LossListResponse,
    LossSummaryResponse,
    PrintDocumentResponse,
    ProductResponse,
    TransactionDetailResponse,
    TransactionListResponse,
    TransferApproveRequest,
    TransferArriveRequest,
    TransferCancelRequest,
    TransferConfirmRequest,
    TransferCreateRequest,
    TransferDetailResponse,
    TransferListResponse,
    TransferShipRequest,
    TransferStockRequest,
    WarehouseResponse,
)

if TYPE_CHECKING:
    from shared.models import (
        Inventory,
        InventoryTransaction,
        LossRecord,
        TransferOrder,
    )

router = APIRouter(prefix="/api/v1")


def format_datetime(dt: datetime | None) -> str | None:
    if dt is None:
        return None
    return dt.strftime("%Y-%m-%d %H:%M:%S")


def build_transfer_detail_response(transfer: "TransferOrder") -> TransferDetailResponse:
    source_wh = db.get_warehouse(transfer.source_warehouse_id)
    target_wh = db.get_warehouse(transfer.target_warehouse_id)

    return TransferDetailResponse(
        transfer_id=transfer.transfer_id,
        transfer_no=transfer.transfer_no,
        source_warehouse_id=transfer.source_warehouse_id,
        target_warehouse_id=transfer.target_warehouse_id,
        source_warehouse_name=source_wh.name if source_wh else None,
        target_warehouse_name=target_wh.name if target_wh else None,
        status=transfer.status,
        items=transfer.items,
        transfer_type=transfer.transfer_type,
        approval_records=transfer.approval_records,
        required_approval_level=transfer.required_approval_level,
        inter_company_approved=transfer.inter_company_approved,
        total_requested_amount=transfer.total_requested_amount,
        total_shipped_amount=transfer.total_shipped_amount,
        total_arrived_amount=transfer.total_arrived_amount,
        total_loss_amount=transfer.total_loss_amount,
        is_approved=transfer.is_approved,
        created_by=transfer.created_by,
        created_at=format_datetime(transfer.created_at),
        confirmed_at=format_datetime(transfer.confirmed_at),
        shipped_at=format_datetime(transfer.shipped_at),
        arrived_at=format_datetime(transfer.arrived_at),
        stocked_at=format_datetime(transfer.stocked_at),
        cancelled_at=format_datetime(transfer.cancelled_at),
        cancelled_by=transfer.cancelled_by,
        remark=transfer.remark,
    )


def build_inventory_detail_response(inventory: "Inventory") -> InventoryDetailResponse:
    warehouse = db.get_warehouse(inventory.warehouse_id)
    product = db.get_product(inventory.product_id)

    available_amount = None
    frozen_amount = None
    total_amount = None

    if product:
        available_amount = inventory.available_quantity * product.unit_price
        frozen_amount = inventory.frozen_quantity * product.unit_price
        total_amount = inventory.total_quantity * product.unit_price

    return InventoryDetailResponse(
        inventory_id=inventory.inventory_id,
        warehouse_id=inventory.warehouse_id,
        warehouse_name=warehouse.name if warehouse else None,
        product_id=inventory.product_id,
        product_sku=product.sku if product else None,
        product_name=product.name if product else None,
        product_unit=product.unit if product else "件",
        product_unit_price=product.unit_price if product else None,
        available_quantity=inventory.available_quantity,
        frozen_quantity=inventory.frozen_quantity,
        total_quantity=inventory.total_quantity,
        available_amount=available_amount,
        frozen_amount=frozen_amount,
        total_amount=total_amount,
        last_updated=format_datetime(inventory.last_updated),
    )


def build_loss_detail_response(loss: "LossRecord") -> LossDetailResponse:
    warehouse = db.get_warehouse(loss.warehouse_id)
    product = db.get_product(loss.product_id)
    transfer = db.get_transfer(loss.transfer_id)

    return LossDetailResponse(
        loss_id=loss.loss_id,
        transfer_id=loss.transfer_id,
        transfer_no=transfer.transfer_no if transfer else None,
        item_id=loss.item_id,
        product_id=loss.product_id,
        product_sku=product.sku if product else None,
        product_name=product.name if product else None,
        warehouse_id=loss.warehouse_id,
        warehouse_name=warehouse.name if warehouse else None,
        loss_quantity=loss.loss_quantity,
        unit_price=loss.unit_price,
        loss_amount=loss.loss_amount,
        loss_date=loss.loss_date.isoformat(),
        created_at=format_datetime(loss.created_at),
        remark=loss.remark,
    )


def build_transaction_detail_response(tx: "InventoryTransaction") -> TransactionDetailResponse:
    warehouse = db.get_warehouse(tx.warehouse_id)
    product = db.get_product(tx.product_id)

    return TransactionDetailResponse(
        transaction_id=tx.transaction_id,
        warehouse_id=tx.warehouse_id,
        warehouse_name=warehouse.name if warehouse else None,
        product_id=tx.product_id,
        product_sku=product.sku if product else None,
        product_name=product.name if product else None,
        transaction_type=tx.transaction_type.value,
        quantity=tx.quantity,
        unit_price=tx.unit_price,
        amount=tx.amount,
        reference_type=tx.reference_type,
        reference_id=tx.reference_id,
        transaction_time=format_datetime(tx.transaction_time),
        remark=tx.remark,
    )


@router.post("/transfers", response_model=TransferDetailResponse)
async def create_transfer(request: TransferCreateRequest) -> TransferDetailResponse:
    transfer = service.create_transfer(request)
    return build_transfer_detail_response(transfer)


@router.get("/transfers/{transfer_id}", response_model=TransferDetailResponse)
async def get_transfer(transfer_id: UUID) -> TransferDetailResponse:
    transfer = service.get_transfer_detail(transfer_id)
    if transfer is None:
        raise HTTPException(status_code=404, detail="Transfer not found")
    return build_transfer_detail_response(transfer)


@router.get("/transfers/no/{transfer_no}", response_model=TransferDetailResponse)
async def get_transfer_by_no(transfer_no: str) -> TransferDetailResponse:
    transfer = service.get_transfer_by_no(transfer_no)
    if transfer is None:
        raise HTTPException(status_code=404, detail="Transfer not found")
    return build_transfer_detail_response(transfer)


@router.get("/transfers", response_model=TransferListResponse)
async def list_transfers(
    status: TransferStatus | None = Query(default=None),
    source_warehouse_id: UUID | None = Query(default=None),
    target_warehouse_id: UUID | None = Query(default=None),
    transfer_type: TransferType | None = Query(default=None),
    created_by: UUID | None = Query(default=None),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=20, ge=1, le=100),
) -> TransferListResponse:
    transfers = service.list_transfers(
        source_warehouse_id=source_warehouse_id,
        target_warehouse_id=target_warehouse_id,
        status=status,
        transfer_type=transfer_type,
        created_by=created_by,
    )

    total = len(transfers)
    start = (page - 1) * page_size
    end = start + page_size
    paged_transfers = transfers[start:end]

    return TransferListResponse(
        total=total,
        page=page,
        page_size=page_size,
        transfers=[build_transfer_detail_response(t) for t in paged_transfers],
    )


@router.post("/transfers/{transfer_id}/confirm", response_model=TransferDetailResponse)
async def confirm_transfer(
    transfer_id: UUID,
    request: TransferConfirmRequest,
) -> TransferDetailResponse:
    transfer = service.confirm_transfer(transfer_id, request)
    return build_transfer_detail_response(transfer)


@router.post("/transfers/{transfer_id}/approve", response_model=TransferDetailResponse)
async def approve_transfer(
    transfer_id: UUID,
    request: TransferApproveRequest,
) -> TransferDetailResponse:
    transfer = service.approve_transfer(
        transfer_id=transfer_id,
        approval_level=request.approval_level,
        approval_status=request.approval_status,
        approver_id=request.approver_id,
        approver_name=request.approver_name,
        comment=request.comment,
    )
    return build_transfer_detail_response(transfer)


@router.post("/transfers/{transfer_id}/approve-inter-company", response_model=TransferDetailResponse)
async def approve_inter_company(
    transfer_id: UUID,
    request: InterCompanyApproveRequest,
) -> TransferDetailResponse:
    transfer = service.approve_inter_company(
        transfer_id=transfer_id,
        operator_id=request.operator_id,
        operator_name=request.operator_name,
        approved=request.approved,
        comment=request.comment,
    )
    return build_transfer_detail_response(transfer)


@router.post("/transfers/{transfer_id}/ship", response_model=TransferDetailResponse)
async def ship_transfer(
    transfer_id: UUID,
    request: TransferShipRequest,
) -> TransferDetailResponse:
    transfer = service.ship_transfer(transfer_id, request)
    return build_transfer_detail_response(transfer)


@router.post("/transfers/{transfer_id}/arrive", response_model=TransferDetailResponse)
async def arrive_transfer(
    transfer_id: UUID,
    request: TransferArriveRequest,
) -> TransferDetailResponse:
    transfer = service.arrive_transfer(transfer_id, request)
    return build_transfer_detail_response(transfer)


@router.post("/transfers/{transfer_id}/stock", response_model=TransferDetailResponse)
async def stock_transfer(
    transfer_id: UUID,
    request: TransferStockRequest,
) -> TransferDetailResponse:
    transfer = service.stock_transfer(transfer_id, request)
    return build_transfer_detail_response(transfer)


@router.post("/transfers/{transfer_id}/cancel", response_model=TransferDetailResponse)
async def cancel_transfer(
    transfer_id: UUID,
    request: TransferCancelRequest,
) -> TransferDetailResponse:
    transfer = service.cancel_transfer(transfer_id, request)
    return build_transfer_detail_response(transfer)


@router.get("/transfers/{transfer_id}/print/outbound", response_model=PrintDocumentResponse)
async def print_outbound(transfer_id: UUID) -> PrintDocumentResponse:
    transfer = service.get_transfer_detail(transfer_id)
    if transfer is None:
        raise HTTPException(status_code=404, detail="Transfer not found")

    content = service.generate_outbound_print_content(transfer)
    return PrintDocumentResponse(
        document_type="outbound",
        transfer_no=transfer.transfer_no,
        content=content,
        generated_at=format_datetime(datetime.now()),
    )


@router.get("/transfers/{transfer_id}/print/inbound", response_model=PrintDocumentResponse)
async def print_inbound(transfer_id: UUID) -> PrintDocumentResponse:
    transfer = service.get_transfer_detail(transfer_id)
    if transfer is None:
        raise HTTPException(status_code=404, detail="Transfer not found")

    if transfer.status not in [TransferStatus.ARRIVED, TransferStatus.STOCKED]:
        raise HTTPException(
            status_code=400,
            detail="Inbound document can only be printed after arrival",
        )

    content = service.generate_inbound_print_content(transfer)
    return PrintDocumentResponse(
        document_type="inbound",
        transfer_no=transfer.transfer_no,
        content=content,
        generated_at=format_datetime(datetime.now()),
    )


@router.get("/inventory", response_model=InventoryOverviewResponse)
async def list_inventory(
    warehouse_id: UUID | None = Query(default=None),
    product_id: UUID | None = Query(default=None),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=100, ge=1, le=100),
) -> InventoryOverviewResponse:
    inventories = db.list_inventories(
        warehouse_id=warehouse_id,
        product_id=product_id,
    )

    total = len(inventories)
    start = (page - 1) * page_size
    end = start + page_size
    paged_inventories = inventories[start:end]

    return InventoryOverviewResponse(
        total=total,
        page=page,
        page_size=page_size,
        inventories=[build_inventory_detail_response(inv) for inv in paged_inventories],
    )


@router.get("/warehouses", response_model=list[WarehouseResponse])
async def list_warehouses() -> list[WarehouseResponse]:
    warehouses = db.list_warehouses()
    return [
        WarehouseResponse(
            warehouse_id=w.warehouse_id,
            name=w.name,
            location=w.location,
            company_id=w.company_id,
            created_at=format_datetime(w.created_at),
            updated_at=format_datetime(w.updated_at),
        )
        for w in warehouses
    ]


@router.get("/products", response_model=list[ProductResponse])
async def list_products() -> list[ProductResponse]:
    products = db.list_products()
    return [
        ProductResponse(
            product_id=p.product_id,
            sku=p.sku,
            name=p.name,
            unit_price=p.unit_price,
            unit=p.unit,
            created_at=format_datetime(p.created_at),
        )
        for p in products
    ]


@router.get("/losses", response_model=LossListResponse)
async def list_losses(
    warehouse_id: UUID | None = Query(default=None),
    transfer_id: UUID | None = Query(default=None),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=20, ge=1, le=100),
) -> LossListResponse:
    losses = db.list_loss_records(
        warehouse_id=warehouse_id,
        transfer_id=transfer_id,
    )

    total = len(losses)
    start = (page - 1) * page_size
    end = start + page_size
    paged_losses = losses[start:end]

    return LossListResponse(
        total=total,
        page=page,
        page_size=page_size,
        losses=[build_loss_detail_response(l) for l in paged_losses],
    )


@router.get("/losses/summary", response_model=LossSummaryResponse)
async def get_loss_summary(
    warehouse_id: UUID | None = Query(default=None),
    year: int | None = Query(default=None),
    month: int | None = Query(default=None),
) -> LossSummaryResponse:
    summaries = db.get_loss_summaries(
        warehouse_id=warehouse_id,
        year=year,
        month=month,
    )
    return LossSummaryResponse(summaries=summaries)


@router.get("/transactions", response_model=TransactionListResponse)
async def list_transactions(
    warehouse_id: UUID | None = Query(default=None),
    product_id: UUID | None = Query(default=None),
    reference_id: UUID | None = Query(default=None),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=20, ge=1, le=100),
) -> TransactionListResponse:
    transactions = db.list_transactions(
        warehouse_id=warehouse_id,
        product_id=product_id,
        reference_id=reference_id,
    )

    total = len(transactions)
    start = (page - 1) * page_size
    end = start + page_size
    paged_transactions = transactions[start:end]

    return TransactionListResponse(
        total=total,
        page=page,
        page_size=page_size,
        transactions=[build_transaction_detail_response(tx) for tx in paged_transactions],
    )


def create_app() -> FastAPI:
    app = FastAPI(
        title=settings.app_name,
        version=settings.app_version,
        debug=settings.debug,
    )

    @app.exception_handler(BusinessException)
    async def business_exception_handler(
        request: Request,
        exc: BusinessException,
    ) -> JSONResponse:
        return JSONResponse(
            status_code=400,
            content={
                "code": exc.error_code.value,
                "message": exc.message,
                "data": exc.details,
            },
        )

    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(
        request: Request,
        exc: RequestValidationError,
    ) -> JSONResponse:
        return JSONResponse(
            status_code=422,
            content={
                "code": ErrorCode.INVALID_PARAM.value,
                "message": "Validation error",
                "data": {"errors": exc.errors()},
            },
        )

    app.include_router(router)

    @app.get("/health")
    async def health_check() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
