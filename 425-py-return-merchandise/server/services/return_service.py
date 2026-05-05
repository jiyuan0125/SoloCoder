from datetime import datetime, timedelta
from typing import List, Optional, Tuple

from shared.models import (
    OutboundOrder,
    ReturnStatus,
    InspectionResult,
    DisposalReason,
    ReturnApplyRequest,
    ReturnApplyResponse,
    WarehouseReceiveRequest,
    InspectionRequest,
    StockInRequest,
    ScrapRecord,
    DefectiveItem,
    ReturnOrderItem,
)
from shared.constants import ErrorCode, ReturnConstants
from server.data_store import DataStore, get_data_store
from server.models import ReturnOrder


class ReturnService:
    def __init__(self, data_store: Optional[DataStore] = None) -> None:
        self._data_store = data_store or get_data_store()

    def _get_outbound_order_or_raise(self, order_id: str) -> OutboundOrder:
        order = self._data_store.get_outbound_order(order_id)
        if not order:
            raise ValueError(f"出库单 {order_id} 不存在")
        return order

    def _get_return_order_or_raise(self, return_id: str) -> ReturnOrder:
        return_order = self._data_store.get_return_order(return_id)
        if not return_order:
            raise ValueError(f"退货单 {return_id} 不存在")
        return return_order

    def _check_return_expiry(self, return_order: ReturnOrder) -> None:
        if return_order.is_expired():
            if return_order.status == ReturnStatus.APPLIED:
                return_order.status = ReturnStatus.EXPIRED
                self._data_store.update_return_order(return_order)
            raise ValueError("退货单已过期")

    def apply_return(self, request: ReturnApplyRequest) -> ReturnApplyResponse:
        outbound = self._get_outbound_order_or_raise(request.outbound_order_id)

        outbound_sku_map = {item.sku: item for item in outbound.items}

        for item in request.items:
            if item.sku not in outbound_sku_map:
                raise ValueError(f"商品 {item.sku} 不在出库单中")

            outbound_item = outbound_sku_map[item.sku]
            already_returned = self._data_store.get_returned_quantity(
                request.outbound_order_id, item.sku
            )
            max_returnable = outbound_item.quantity - already_returned
            if item.quantity > max_returnable:
                raise ValueError(
                    f"商品 {item.sku} 退货数量超出限制，最多可退 {max_returnable} 件"
                )

            item.original_unit_price = outbound_item.unit_price
            item.category = outbound_item.category
            item.product_name = outbound_item.product_name

        now = datetime.now()
        expiry_date = now + timedelta(days=ReturnConstants.RETURN_VALIDITY_DAYS)

        return_order = ReturnOrder(
            return_order_id=request.return_order_id,
            outbound_order_id=request.outbound_order_id,
            status=ReturnStatus.APPLIED,
            items=request.items,
            reason=request.reason,
            customer_note=request.customer_note,
            created_at=now,
            expiry_date=expiry_date,
        )

        self._data_store.create_return_order(return_order)

        return ReturnApplyResponse(
            return_order_id=return_order.return_order_id,
            outbound_order_id=return_order.outbound_order_id,
            status=return_order.status,
            created_at=return_order.created_at,
            expiry_date=return_order.expiry_date,
        )

    def warehouse_receive(self, request: WarehouseReceiveRequest) -> ReturnOrder:
        return_order = self._get_return_order_or_raise(request.return_order_id)

        if return_order.status != ReturnStatus.APPLIED:
            raise ValueError("退货单状态不是已申请，无法收货")

        self._check_return_expiry(return_order)

        now = request.received_at or datetime.now()
        return_order.status = ReturnStatus.WAREHOUSE_RECEIVED
        return_order.warehouse_received_at = now
        return_order.warehouse_received_by = request.received_by

        return self._data_store.update_return_order(return_order)

    def inspect(self, request: InspectionRequest) -> ReturnOrder:
        return_order = self._get_return_order_or_raise(request.return_order_id)

        if return_order.status != ReturnStatus.WAREHOUSE_RECEIVED:
            raise ValueError("退货单状态不是已收货，无法质检")

        for item in return_order.items:
            if item.sku not in request.item_results:
                raise ValueError(f"商品 {item.sku} 缺少质检结果")

            result = request.item_results[item.sku]
            item.inspection_result = result

            if result == InspectionResult.MINOR_DEFECT:
                item.disposal_price = item.original_unit_price * ReturnConstants.DEFECTIVE_PRICE_RATE
            elif result == InspectionResult.GOOD:
                item.disposal_price = item.original_unit_price

        now = request.inspection_date or datetime.now()
        return_order.status = ReturnStatus.INSPECTED
        return_order.inspected_at = now
        return_order.inspected_by = request.inspector
        return_order.stockin_deadline = now + timedelta(days=ReturnConstants.STOCKIN_DEADLINE_DAYS)

        return self._data_store.update_return_order(return_order)

    def stock_in(self, request: StockInRequest) -> ReturnOrder:
        return_order = self._get_return_order_or_raise(request.return_order_id)

        if return_order.status != ReturnStatus.INSPECTED:
            raise ValueError("退货单状态不是已质检，无法入库")

        now = request.stock_in_at or datetime.now()
        if return_order.stockin_deadline and now > return_order.stockin_deadline:
            return_order.status = ReturnStatus.RETURNED_TO_CUSTOMER
            self._data_store.update_return_order(return_order)
            raise ValueError("入库期限已过，退货已退回给客户")

        for item in return_order.items:
            if item.inspection_result == InspectionResult.GOOD:
                if item.disposal_price:
                    self._data_store.add_normal_inventory(
                        item.sku, item.quantity, item.disposal_price
                    )
                self._data_store.add_returned_quantity(
                    return_order.outbound_order_id, item.sku, item.quantity
                )

            elif item.inspection_result == InspectionResult.MINOR_DEFECT:
                defective_item = DefectiveItem(
                    defective_id=self._data_store.generate_defective_id(),
                    sku=item.sku,
                    product_name=item.product_name,
                    quantity=item.quantity,
                    unit_price=item.original_unit_price,
                    defective_price=item.disposal_price or 0.0,
                    category=item.category,
                    return_order_id=return_order.return_order_id,
                    created_at=now,
                )
                self._data_store.create_defective_item(defective_item)
                self._data_store.add_returned_quantity(
                    return_order.outbound_order_id, item.sku, item.quantity
                )

            elif item.inspection_result == InspectionResult.SEVERE_DAMAGE:
                month_key = now.strftime("%Y-%m")
                scrap_record = ScrapRecord(
                    scrap_id=self._data_store.generate_scrap_id(),
                    return_order_id=return_order.return_order_id,
                    sku=item.sku,
                    product_name=item.product_name,
                    quantity=item.quantity,
                    original_price=item.original_unit_price,
                    reason=return_order.reason,
                    approver=request.stock_in_by,
                    created_at=now,
                    month=month_key,
                )
                self._data_store.create_scrap_record(scrap_record)
                self._data_store.add_returned_quantity(
                    return_order.outbound_order_id, item.sku, item.quantity
                )

        return_order.status = ReturnStatus.STOCKED_IN
        return_order.stocked_in_at = now
        return_order.stocked_in_by = request.stock_in_by

        return self._data_store.update_return_order(return_order)

    def get_return_order(self, return_id: str) -> ReturnOrder:
        return self._get_return_order_or_raise(return_id)

    def list_return_orders(self) -> List[ReturnOrder]:
        return self._data_store.list_return_orders()
