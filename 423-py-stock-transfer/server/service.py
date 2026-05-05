from datetime import date, datetime
from decimal import Decimal
from typing import TYPE_CHECKING
from uuid import UUID, uuid4

from server.config import settings
from server.database import db
from server.exceptions import (
    ApprovalPermissionDeniedException,
    InsufficientInventoryException,
    InterCompanyApprovalPendingException,
    InvalidParamException,
    InvalidStatusTransitionException,
    ProductNotFoundException,
    TransferAlreadyApprovedException,
    TransferNotFoundException,
    TransferRejectedException,
    WarehouseNotFoundException,
)
from shared.enums import (
    ApprovalLevel,
    ApprovalStatus,
    TransactionType,
    TransferStatus,
    TransferType,
)
from shared.models import (
    ApprovalRecord,
    Inventory,
    InventoryTransaction,
    LossRecord,
    TransferItem,
    TransferOrder,
)

if TYPE_CHECKING:
    from shared.protocols import (
        TransferArriveRequest,
        TransferCancelRequest,
        TransferConfirmRequest,
        TransferCreateRequest,
        TransferShipRequest,
        TransferStockRequest,
    )


def generate_transfer_no() -> str:
    return db.get_next_transfer_no(settings.transfer_no_prefix)


def calculate_approval_level(total_amount: Decimal) -> ApprovalLevel:
    threshold1 = Decimal(settings.approval_threshold_1)
    threshold2 = Decimal(settings.approval_threshold_2)

    if total_amount >= threshold2:
        return ApprovalLevel.LOGISTICS_DIRECTOR
    elif total_amount >= threshold1:
        return ApprovalLevel.WAREHOUSE_MANAGER
    return ApprovalLevel.NONE


def check_warehouse_company(
    source_warehouse_id: UUID,
    target_warehouse_id: UUID,
) -> tuple[bool, UUID, UUID]:
    source_wh = db.get_warehouse(source_warehouse_id)
    target_wh = db.get_warehouse(target_warehouse_id)

    if source_wh is None:
        raise WarehouseNotFoundException(str(source_warehouse_id))
    if target_wh is None:
        raise WarehouseNotFoundException(str(target_warehouse_id))

    is_same_company = source_wh.company_id == target_wh.company_id
    return is_same_company, source_wh.company_id, target_wh.company_id


def create_transfer(request: "TransferCreateRequest") -> TransferOrder:
    source_wh = db.get_warehouse(request.source_warehouse_id)
    target_wh = db.get_warehouse(request.target_warehouse_id)

    if source_wh is None:
        raise WarehouseNotFoundException(str(request.source_warehouse_id))
    if target_wh is None:
        raise WarehouseNotFoundException(str(request.target_warehouse_id))

    items: list[TransferItem] = []
    total_amount = Decimal("0")

    for req_item in request.items:
        product = db.get_product(req_item.product_id)
        if product is None:
            raise ProductNotFoundException(str(req_item.product_id))

        inventory = db.get_inventory(request.source_warehouse_id, req_item.product_id)
        available_qty = inventory.available_quantity if inventory else 0

        if available_qty < req_item.requested_quantity:
            raise InsufficientInventoryException(
                warehouse_id=str(request.source_warehouse_id),
                product_id=str(req_item.product_id),
                available=available_qty,
                required=req_item.requested_quantity,
            )

        item = TransferItem(
            item_id=uuid4(),
            product_id=req_item.product_id,
            requested_quantity=req_item.requested_quantity,
            unit_price=product.unit_price,
        )
        items.append(item)
        total_amount += item.requested_amount

    is_same_company = source_wh.company_id == target_wh.company_id
    transfer_type = (
        TransferType.INTRA_COMPANY if is_same_company else TransferType.INTER_COMPANY
    )

    approval_level = calculate_approval_level(total_amount)

    transfer = TransferOrder(
        transfer_id=uuid4(),
        transfer_no=generate_transfer_no(),
        source_warehouse_id=request.source_warehouse_id,
        target_warehouse_id=request.target_warehouse_id,
        status=TransferStatus.PENDING_CONFIRM,
        items=items,
        transfer_type=transfer_type,
        required_approval_level=approval_level,
        inter_company_approved=is_same_company,
        created_by=request.created_by,
        remark=request.remark,
    )

    return db.create_transfer(transfer)


def validate_status_transition(
    current_status: TransferStatus,
    target_status: TransferStatus,
) -> bool:
    valid_transitions: dict[TransferStatus, list[TransferStatus]] = {
        TransferStatus.PENDING_CONFIRM: [
            TransferStatus.CONFIRMED,
            TransferStatus.CANCELLED,
        ],
        TransferStatus.CONFIRMED: [
            TransferStatus.IN_TRANSIT,
            TransferStatus.CANCELLED,
        ],
        TransferStatus.IN_TRANSIT: [TransferStatus.ARRIVED],
        TransferStatus.ARRIVED: [TransferStatus.STOCKED],
    }

    valid_targets = valid_transitions.get(current_status, [])
    return target_status in valid_targets


def confirm_transfer(transfer_id: UUID, request: "TransferConfirmRequest") -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if not validate_status_transition(transfer.status, TransferStatus.CONFIRMED):
        raise InvalidStatusTransitionException(
            current_status=transfer.status,
            target_status=TransferStatus.CONFIRMED,
        )

    if transfer.required_approval_level != ApprovalLevel.NONE and not transfer.is_approved:
        for record in transfer.approval_records:
            if (
                record.approval_level == transfer.required_approval_level
                and record.approval_status == ApprovalStatus.REJECTED
            ):
                raise TransferRejectedException(str(transfer_id))

    if transfer.transfer_type == TransferType.INTER_COMPANY:
        if not transfer.inter_company_approved:
            raise InterCompanyApprovalPendingException(str(transfer_id))

    for item in transfer.items:
        inventory = db.get_inventory(transfer.source_warehouse_id, item.product_id)
        if inventory is None or inventory.available_quantity < item.requested_quantity:
            available = inventory.available_quantity if inventory else 0
            raise InsufficientInventoryException(
                warehouse_id=str(transfer.source_warehouse_id),
                product_id=str(item.product_id),
                available=available,
                required=item.requested_quantity,
            )

    def update_transfer(t: TransferOrder) -> None:
        t.status = TransferStatus.CONFIRMED
        t.confirmed_at = datetime.now()

    transfer = db.update_transfer(transfer_id, update_transfer)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    for item in transfer.items:
        product = db.get_product(item.product_id)
        unit_price = product.unit_price if product else Decimal("0")
        amount = item.requested_quantity * unit_price

        transaction = InventoryTransaction(
            transaction_id=uuid4(),
            warehouse_id=transfer.source_warehouse_id,
            product_id=item.product_id,
            transaction_type=TransactionType.FREEZE,
            quantity=item.requested_quantity,
            unit_price=unit_price,
            amount=amount,
            reference_id=transfer.transfer_id,
        )
        db.create_transaction(transaction)

        def freeze_inv(inv: "Inventory") -> None:
            inv.available_quantity -= item.requested_quantity
            inv.frozen_quantity += item.requested_quantity

        db.update_inventory(transfer.source_warehouse_id, item.product_id, freeze_inv)

    return transfer


def approve_transfer(
    transfer_id: UUID,
    approval_level: ApprovalLevel,
    approval_status: ApprovalStatus,
    approver_id: UUID,
    approver_name: str,
    comment: str | None = None,
) -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if transfer.is_approved:
        raise TransferAlreadyApprovedException(str(transfer_id))

    if transfer.required_approval_level == ApprovalLevel.NONE:
        raise ApprovalPermissionDeniedException(str(transfer_id), "none")

    required_level = transfer.required_approval_level
    if approval_level != required_level:
        raise ApprovalPermissionDeniedException(str(transfer_id), required_level)

    for record in transfer.approval_records:
        if (
            record.approval_level == approval_level
            and record.approval_status == ApprovalStatus.APPROVED
        ):
            raise TransferAlreadyApprovedException(str(transfer_id))
        if (
            record.approval_level == approval_level
            and record.approval_status == ApprovalStatus.REJECTED
        ):
            raise TransferRejectedException(str(transfer_id))

    approval_record = ApprovalRecord(
        approval_id=uuid4(),
        transfer_id=transfer_id,
        approval_level=approval_level,
        approval_status=approval_status,
        approver_id=approver_id,
        approver_name=approver_name,
        comment=comment,
        approved_at=datetime.now() if approval_status == ApprovalStatus.APPROVED else None,
    )

    def update_transfer(t: TransferOrder) -> None:
        t.approval_records.append(approval_record)

    updated = db.update_transfer(transfer_id, update_transfer)
    if updated is None:
        raise TransferNotFoundException(str(transfer_id))
    return updated


def approve_inter_company(
    transfer_id: UUID,
    operator_id: UUID,
    operator_name: str,
    approved: bool,
    comment: str | None = None,
) -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if transfer.transfer_type != TransferType.INTER_COMPANY:
        raise InvalidParamException(
            "This is not an inter-company transfer", "transfer_type"
        )

    if transfer.inter_company_approved:
        raise TransferAlreadyApprovedException(str(transfer_id))

    def update_transfer(t: TransferOrder) -> None:
        t.inter_company_approved = approved
        if t.remark:
            if comment:
                t.remark = f"{t.remark}\n跨公司审批 - {operator_name}: {comment}"
            else:
                t.remark = f"{t.remark}\n跨公司审批 - {operator_name}: {'已通过' if approved else '已拒绝'}"
        else:
            if comment:
                t.remark = f"跨公司审批 - {operator_name}: {comment}"
            else:
                t.remark = f"跨公司审批 - {operator_name}: {'已通过' if approved else '已拒绝'}"

    updated = db.update_transfer(transfer_id, update_transfer)
    if updated is None:
        raise TransferNotFoundException(str(transfer_id))
    return updated


def ship_transfer(transfer_id: UUID, request: "TransferShipRequest") -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if not validate_status_transition(transfer.status, TransferStatus.IN_TRANSIT):
        raise InvalidStatusTransitionException(
            current_status=transfer.status,
            target_status=TransferStatus.IN_TRANSIT,
        )

    def update_transfer(t: TransferOrder) -> None:
        t.status = TransferStatus.IN_TRANSIT
        t.shipped_at = datetime.now()
        for item in t.items:
            item.shipped_quantity = item.requested_quantity

    updated = db.update_transfer(transfer_id, update_transfer)
    if updated is None:
        raise TransferNotFoundException(str(transfer_id))
    return updated


def arrive_transfer(transfer_id: UUID, request: "TransferArriveRequest") -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if not validate_status_transition(transfer.status, TransferStatus.ARRIVED):
        raise InvalidStatusTransitionException(
            current_status=transfer.status,
            target_status=TransferStatus.ARRIVED,
        )

    arrival_items_map = {item.item_id: item.arrived_quantity for item in request.arrival_items}

    for item in transfer.items:
        if item.item_id not in arrival_items_map:
            raise InvalidParamException(
                f"Missing arrival quantity for item: {item.item_id}",
                "arrival_items",
            )

        arrived_qty = arrival_items_map[item.item_id]
        if item.shipped_quantity is not None and arrived_qty > item.shipped_quantity:
            raise InvalidParamException(
                f"Arrived quantity ({arrived_qty}) cannot exceed "
                f"shipped quantity ({item.shipped_quantity})",
                "arrival_items",
            )

    def update_transfer(t: TransferOrder) -> None:
        t.status = TransferStatus.ARRIVED
        t.arrived_at = datetime.now()
        for item in t.items:
            item.arrived_quantity = arrival_items_map[item.item_id]
            if item.shipped_quantity is not None and item.arrived_quantity is not None:
                item.loss_quantity = item.shipped_quantity - item.arrived_quantity

    updated = db.update_transfer(transfer_id, update_transfer)
    if updated is None:
        raise TransferNotFoundException(str(transfer_id))
    return updated


def stock_transfer(transfer_id: UUID, request: "TransferStockRequest") -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if not validate_status_transition(transfer.status, TransferStatus.STOCKED):
        raise InvalidStatusTransitionException(
            current_status=transfer.status,
            target_status=TransferStatus.STOCKED,
        )

    def update_transfer(t: TransferOrder) -> None:
        t.status = TransferStatus.STOCKED
        t.stocked_at = datetime.now()

    transfer = db.update_transfer(transfer_id, update_transfer)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    today = date.today()

    for item in transfer.items:
        product = db.get_product(item.product_id)
        unit_price = product.unit_price if product else Decimal("0")

        shipped_qty = item.shipped_quantity
        if shipped_qty is not None and shipped_qty > 0:
            amount = shipped_qty * unit_price
            transaction = InventoryTransaction(
                transaction_id=uuid4(),
                warehouse_id=transfer.source_warehouse_id,
                product_id=item.product_id,
                transaction_type=TransactionType.TRANSFER_OUT,
                quantity=shipped_qty,
                unit_price=unit_price,
                amount=amount,
                reference_id=transfer.transfer_id,
            )
            db.create_transaction(transaction)

            def unfreeze_source(inv: Inventory) -> None:
                inv.frozen_quantity -= shipped_qty

            db.update_inventory(
                transfer.source_warehouse_id,
                item.product_id,
                unfreeze_source,
            )

            db.update_loss_summary_transfer(
                warehouse_id=transfer.source_warehouse_id,
                loss_date=today,
                transfer_quantity=shipped_qty,
                transfer_amount=shipped_qty * unit_price,
            )

        arrived_qty = item.arrived_quantity
        if arrived_qty is not None and arrived_qty > 0:
            amount = arrived_qty * unit_price
            transaction = InventoryTransaction(
                transaction_id=uuid4(),
                warehouse_id=transfer.target_warehouse_id,
                product_id=item.product_id,
                transaction_type=TransactionType.TRANSFER_IN,
                quantity=arrived_qty,
                unit_price=unit_price,
                amount=amount,
                reference_id=transfer.transfer_id,
            )
            db.create_transaction(transaction)

            def add_target(inv: Inventory) -> None:
                inv.available_quantity += arrived_qty

            db.update_inventory(
                transfer.target_warehouse_id,
                item.product_id,
                add_target,
            )

        if item.loss_quantity > 0:
            loss_amount = item.loss_quantity * unit_price
            loss_record = LossRecord(
                loss_id=uuid4(),
                transfer_id=transfer.transfer_id,
                item_id=item.item_id,
                product_id=item.product_id,
                warehouse_id=transfer.source_warehouse_id,
                loss_quantity=item.loss_quantity,
                unit_price=unit_price,
                loss_amount=loss_amount,
                loss_date=today,
            )
            db.create_loss_record(loss_record)

            transaction = InventoryTransaction(
                transaction_id=uuid4(),
                warehouse_id=transfer.source_warehouse_id,
                product_id=item.product_id,
                transaction_type=TransactionType.LOSS,
                quantity=item.loss_quantity,
                unit_price=unit_price,
                amount=loss_amount,
                reference_id=transfer.transfer_id,
            )
            db.create_transaction(transaction)

    return transfer


def cancel_transfer(transfer_id: UUID, request: "TransferCancelRequest") -> TransferOrder:
    transfer = db.get_transfer(transfer_id)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if not validate_status_transition(transfer.status, TransferStatus.CANCELLED):
        raise InvalidStatusTransitionException(
            current_status=transfer.status,
            target_status=TransferStatus.CANCELLED,
        )

    original_status = transfer.status

    def update_transfer(t: TransferOrder) -> None:
        t.status = TransferStatus.CANCELLED
        t.cancelled_at = datetime.now()
        t.cancelled_by = request.operator_id
        if t.remark:
            t.remark = f"{t.remark}\n取消原因: {request.reason}"
        else:
            t.remark = f"取消原因: {request.reason}"

    transfer = db.update_transfer(transfer_id, update_transfer)
    if transfer is None:
        raise TransferNotFoundException(str(transfer_id))

    if original_status == TransferStatus.CONFIRMED:
        for item in transfer.items:
            product = db.get_product(item.product_id)
            unit_price = product.unit_price if product else Decimal("0")
            amount = item.requested_quantity * unit_price

            transaction = InventoryTransaction(
                transaction_id=uuid4(),
                warehouse_id=transfer.source_warehouse_id,
                product_id=item.product_id,
                transaction_type=TransactionType.UNFREEZE,
                quantity=item.requested_quantity,
                unit_price=unit_price,
                amount=amount,
                reference_id=transfer.transfer_id,
            )
            db.create_transaction(transaction)

            def unfreeze_inv(inv: Inventory) -> None:
                inv.frozen_quantity -= item.requested_quantity
                inv.available_quantity += item.requested_quantity

            db.update_inventory(
                transfer.source_warehouse_id,
                item.product_id,
                unfreeze_inv,
            )

    return transfer


def get_transfer_detail(transfer_id: UUID) -> TransferOrder | None:
    return db.get_transfer(transfer_id)


def get_transfer_by_no(transfer_no: str) -> TransferOrder | None:
    return db.get_transfer_by_no(transfer_no)


def list_transfers(
    source_warehouse_id: UUID | None = None,
    target_warehouse_id: UUID | None = None,
    status: TransferStatus | None = None,
    transfer_type: TransferType | None = None,
    created_by: UUID | None = None,
) -> list[TransferOrder]:
    transfers = db.list_transfers(
        source_warehouse_id=source_warehouse_id,
        target_warehouse_id=target_warehouse_id,
        status=status.value if status else None,
    )

    filtered: list[TransferOrder] = []
    for t in transfers:
        if transfer_type and t.transfer_type != transfer_type:
            continue
        if created_by and t.created_by != created_by:
            continue
        filtered.append(t)

    return filtered


def generate_outbound_print_content(transfer: TransferOrder) -> str:
    source_wh = db.get_warehouse(transfer.source_warehouse_id)
    target_wh = db.get_warehouse(transfer.target_warehouse_id)

    source_name = source_wh.name if source_wh else "Unknown"
    target_name = target_wh.name if target_wh else "Unknown"

    lines: list[str] = []
    lines.append("=" * 60)
    lines.append("                     调 拨 出 库 单")
    lines.append("=" * 60)
    lines.append(f"调拨单号: {transfer.transfer_no}")
    lines.append(f"创建时间: {transfer.created_at.strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append("-" * 60)
    lines.append(f"源仓库: {source_name}")
    lines.append(f"目标仓库: {target_name}")
    lines.append(f"调拨类型: {'公司内调拨' if transfer.transfer_type == TransferType.INTRA_COMPANY else '跨公司调拨'}")
    lines.append("-" * 60)
    lines.append(f"{'商品SKU':<15} {'商品名称':<20} {'数量':<8} {'单价':<12} {'金额':<12}")
    lines.append("-" * 60)

    for item in transfer.items:
        product = db.get_product(item.product_id)
        sku = product.sku if product else "Unknown"
        name = product.name if product else "Unknown"
        qty = item.shipped_quantity if item.shipped_quantity else item.requested_quantity
        lines.append(
            f"{sku:<15} {name:<20} {qty:<8} "
            f"{item.unit_price:<12.2f} {item.requested_amount:<12.2f}"
        )

    lines.append("-" * 60)
    lines.append(f"{'合计':<35} {transfer.total_requested_amount:>.2f}")
    lines.append("-" * 60)
    lines.append("出库仓管员签字: _______________")
    lines.append("日期: _______________")
    lines.append("=" * 60)

    return "\n".join(lines)


def generate_inbound_print_content(transfer: TransferOrder) -> str:
    source_wh = db.get_warehouse(transfer.source_warehouse_id)
    target_wh = db.get_warehouse(transfer.target_warehouse_id)

    source_name = source_wh.name if source_wh else "Unknown"
    target_name = target_wh.name if target_wh else "Unknown"

    lines: list[str] = []
    lines.append("=" * 60)
    lines.append("                     调 拨 入 库 单")
    lines.append("=" * 60)
    lines.append(f"调拨单号: {transfer.transfer_no}")
    lines.append(f"创建时间: {transfer.created_at.strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append("-" * 60)
    lines.append(f"源仓库: {source_name}")
    lines.append(f"目标仓库: {target_name}")
    lines.append("-" * 60)
    lines.append(f"{'商品SKU':<15} {'商品名称':<20} {'发出数量':<10} {'实收数量':<10} {'损耗数量':<10}")
    lines.append("-" * 60)

    for item in transfer.items:
        product = db.get_product(item.product_id)
        sku = product.sku if product else "Unknown"
        name = product.name if product else "Unknown"
        shipped = item.shipped_quantity if item.shipped_quantity else 0
        arrived = item.arrived_quantity if item.arrived_quantity else 0
        loss = item.loss_quantity
        lines.append(
            f"{sku:<15} {name:<20} {shipped:<10} {arrived:<10} {loss:<10}"
        )

    lines.append("-" * 60)
    if transfer.total_loss_amount > 0:
        lines.append(f"损耗金额: {transfer.total_loss_amount:.2f}")
    lines.append("-" * 60)
    lines.append("入库仓管员签字: _______________")
    lines.append("日期: _______________")
    lines.append("=" * 60)

    return "\n".join(lines)
