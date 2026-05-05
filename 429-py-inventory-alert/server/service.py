from datetime import datetime, timedelta, timezone
from typing import List, Optional, Tuple
from uuid import uuid4

from shared.errors import ErrorCode
from shared.models import (
    AlertLevel,
    AlertQueryParams,
    AlertRecord,
    AlertType,
    BulkSafetyStockUpdate,
    InventoryHealthReport,
    InventoryHealthReportItem,
    InventoryStatus,
    InventoryTransaction,
    InventoryTurnoverAnalysis,
    Month,
    Product,
    ProductCreate,
    ProductUpdate,
    ReplenishmentOrderDraft,
    SafetyStock,
    SeasonalSafetyStock,
    StockAdjustment,
    TransactionType,
    TurnoverAnalysisItem,
)
from server.storage import MemoryStorage


ALERT_COOLDOWN_HOURS = 24
TURNOVER_DAYS_INFINITY = 999999999.0


class InventoryService:
    def __init__(self, storage: MemoryStorage) -> None:
        self._storage = storage
        self._cooldown_delta = timedelta(hours=ALERT_COOLDOWN_HOURS)

    def _now_utc(self) -> datetime:
        return datetime.now(timezone.utc)

    def _generate_id(self) -> str:
        return str(uuid4())

    def _get_current_month(self) -> int:
        return self._now_utc().month

    def create_product(self, product_data: ProductCreate) -> Product:
        product = self._storage.create_product(product_data)
        self._check_and_trigger_alerts(product.sku)
        return product

    def get_product(self, sku: str) -> Optional[Product]:
        return self._storage.get_product(sku)

    def update_product(self, sku: str, update_data: ProductUpdate) -> Optional[Product]:
        return self._storage.update_product(sku, update_data)

    def delete_product(self, sku: str) -> bool:
        return self._storage.delete_product(sku)

    def list_products(self, category: Optional[str] = None) -> List[Product]:
        return self._storage.list_products(category)

    def update_safety_stock(self, sku: str, safety_stock: SafetyStock) -> Optional[Product]:
        product = self._storage.update_safety_stock(sku, safety_stock)
        if product is not None:
            self._check_and_trigger_alerts(sku)
        return product

    def bulk_update_safety_stock(self, bulk_update: BulkSafetyStockUpdate) -> int:
        count = self._storage.bulk_update_safety_stock(bulk_update)
        products = self._storage.list_products(bulk_update.category)
        for product in products:
            self._check_and_trigger_alerts(product.sku)
        return count

    def stock_inbound(self, sku: str, adjustment: StockAdjustment) -> InventoryTransaction:
        product = self._storage.get_product(sku)
        if product is None:
            raise ValueError(ErrorCode.PRODUCT_NOT_FOUND.value)

        transaction = self._storage.adjust_stock(sku, TransactionType.INBOUND, adjustment)
        if transaction is None:
            raise ValueError(ErrorCode.INTERNAL_ERROR.value)

        self._check_and_resolve_alerts(sku)
        self._check_and_trigger_alerts(sku)

        return transaction

    def stock_outbound(self, sku: str, adjustment: StockAdjustment) -> InventoryTransaction:
        product = self._storage.get_product(sku)
        if product is None:
            raise ValueError(ErrorCode.PRODUCT_NOT_FOUND.value)

        try:
            transaction = self._storage.adjust_stock(sku, TransactionType.OUTBOUND, adjustment)
        except ValueError as e:
            if str(e) == ErrorCode.INSUFFICIENT_STOCK.value:
                raise
            raise ValueError(ErrorCode.INTERNAL_ERROR.value)

        if transaction is None:
            raise ValueError(ErrorCode.INTERNAL_ERROR.value)

        self._check_and_resolve_alerts(sku)
        self._check_and_trigger_alerts(sku)

        return transaction

    def stock_transfer(self, sku: str, adjustment: StockAdjustment) -> InventoryTransaction:
        product = self._storage.get_product(sku)
        if product is None:
            raise ValueError(ErrorCode.PRODUCT_NOT_FOUND.value)

        try:
            transaction = self._storage.adjust_stock(sku, TransactionType.TRANSFER, adjustment)
        except ValueError as e:
            if str(e) == ErrorCode.INSUFFICIENT_STOCK.value:
                raise
            raise ValueError(ErrorCode.INTERNAL_ERROR.value)

        if transaction is None:
            raise ValueError(ErrorCode.INTERNAL_ERROR.value)

        self._check_and_resolve_alerts(sku)
        self._check_and_trigger_alerts(sku)

        return transaction

    def _check_and_trigger_alerts(self, sku: str) -> None:
        product = self._storage.get_product(sku)
        if product is None:
            return

        current_month = self._get_current_month()
        ss = product.safety_stock.get_current_safety_stock(current_month)
        current_stock = product.current_stock
        now = self._now_utc()

        if current_stock < ss.min_stock:
            if not self._is_in_cooldown(sku, AlertType.LOW_STOCK, now):
                self._trigger_low_stock_alert(product, ss, now)
        elif current_stock > ss.max_stock:
            if not self._is_in_cooldown(sku, AlertType.OVERSTOCK, now):
                self._trigger_overstock_alert(product, ss, now)

    def _is_in_cooldown(self, sku: str, alert_type: AlertType, now: datetime) -> bool:
        latest_alert = self._storage.get_latest_alert_by_type(sku, alert_type)
        if latest_alert is None:
            return False
        return (now - latest_alert.triggered_at) < self._cooldown_delta

    def _trigger_low_stock_alert(
        self, product: Product, ss: SeasonalSafetyStock, now: datetime
    ) -> None:
        suggestion_qty = ss.min_stock - product.current_stock

        alert_level = AlertLevel.WARNING
        if product.current_stock < (ss.min_stock // 2):
            alert_level = AlertLevel.CRITICAL

        alert = AlertRecord(
            id=self._generate_id(),
            sku=product.sku,
            alert_type=AlertType.LOW_STOCK,
            alert_level=alert_level,
            current_stock=product.current_stock,
            min_stock=ss.min_stock,
            max_stock=ss.max_stock,
            suggestion_quantity=suggestion_qty,
            triggered_at=now,
        )

        self._storage.create_alert(alert)

        draft = ReplenishmentOrderDraft(
            id=self._generate_id(),
            sku=product.sku,
            alert_id=alert.id,
            suggested_quantity=suggestion_qty,
            created_at=now,
        )
        self._storage.create_replenishment_draft(draft)

    def _trigger_overstock_alert(
        self, product: Product, ss: SeasonalSafetyStock, now: datetime
    ) -> None:
        suggestion_qty = product.current_stock - ss.max_stock

        alert_level = AlertLevel.WARNING
        if product.current_stock > (ss.max_stock * 2):
            alert_level = AlertLevel.CRITICAL

        alert = AlertRecord(
            id=self._generate_id(),
            sku=product.sku,
            alert_type=AlertType.OVERSTOCK,
            alert_level=alert_level,
            current_stock=product.current_stock,
            min_stock=ss.min_stock,
            max_stock=ss.max_stock,
            suggestion_quantity=suggestion_qty,
            triggered_at=now,
        )

        self._storage.create_alert(alert)

    def _check_and_resolve_alerts(self, sku: str) -> None:
        product = self._storage.get_product(sku)
        if product is None:
            return

        unresolved_alerts = self._storage.get_unresolved_alerts_by_sku(sku)
        if not unresolved_alerts:
            return

        current_month = self._get_current_month()
        ss = product.safety_stock.get_current_safety_stock(current_month)
        now = self._now_utc()

        for alert in unresolved_alerts:
            if alert.alert_type == AlertType.LOW_STOCK:
                if product.current_stock >= ss.min_stock:
                    self._storage.resolve_alert(alert.id, now)
            elif alert.alert_type == AlertType.OVERSTOCK:
                if product.current_stock <= ss.max_stock:
                    self._storage.resolve_alert(alert.id, now)

    def get_alert(self, alert_id: str) -> Optional[AlertRecord]:
        return self._storage.get_alert(alert_id)

    def query_alerts(self, params: AlertQueryParams) -> List[AlertRecord]:
        return self._storage.query_alerts(params)

    def resolve_alert_manually(self, alert_id: str) -> Optional[AlertRecord]:
        return self._storage.resolve_alert(alert_id, self._now_utc())

    def get_replenishment_draft(self, draft_id: str) -> Optional[ReplenishmentOrderDraft]:
        return self._storage.get_replenishment_draft(draft_id)

    def list_replenishment_drafts(
        self, sku: Optional[str] = None, confirmed: Optional[bool] = None
    ) -> List[ReplenishmentOrderDraft]:
        return self._storage.list_replenishment_drafts(sku, confirmed)

    def confirm_replenishment_draft(self, draft_id: str) -> Optional[ReplenishmentOrderDraft]:
        draft = self._storage.get_replenishment_draft(draft_id)
        if draft is None or draft.confirmed:
            return None

        return self._storage.confirm_replenishment_draft(draft_id, self._now_utc())

    def get_transaction(self, transaction_id: str) -> Optional[InventoryTransaction]:
        return self._storage.get_transaction(transaction_id)

    def list_transactions(
        self,
        sku: Optional[str] = None,
        transaction_type: Optional[TransactionType] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
    ) -> List[InventoryTransaction]:
        return self._storage.list_transactions(sku, transaction_type, start_time, end_time)

    def generate_health_report(self) -> InventoryHealthReport:
        now = self._now_utc()
        current_month = self._get_current_month()
        products = self._storage.list_products()

        items: List[InventoryHealthReportItem] = []
        normal_count = 0
        low_stock_count = 0
        overstock_count = 0

        for product in products:
            ss = product.safety_stock.get_current_safety_stock(current_month)
            status = product.get_inventory_status(current_month)

            item = InventoryHealthReportItem(
                sku=product.sku,
                name=product.name,
                category=product.category,
                current_stock=product.current_stock,
                min_stock=ss.min_stock,
                max_stock=ss.max_stock,
                status=status,
            )
            items.append(item)

            if status == InventoryStatus.NORMAL:
                normal_count += 1
            elif status == InventoryStatus.LOW_STOCK:
                low_stock_count += 1
            elif status == InventoryStatus.OVERSTOCK:
                overstock_count += 1

        report = InventoryHealthReport(
            id=self._generate_id(),
            generated_at=now,
            total_products=len(products),
            normal_count=normal_count,
            low_stock_count=low_stock_count,
            overstock_count=overstock_count,
            items=items,
        )

        return self._storage.create_health_report(report)

    def get_health_report(self, report_id: str) -> Optional[InventoryHealthReport]:
        return self._storage.get_health_report(report_id)

    def get_latest_health_report(self) -> Optional[InventoryHealthReport]:
        return self._storage.get_latest_health_report()

    def generate_turnover_analysis(
        self, start_date: datetime, end_date: datetime
    ) -> InventoryTurnoverAnalysis:
        products = self._storage.list_products()
        items: List[TurnoverAnalysisItem] = []

        for product in products:
            outbound_transactions = self._storage.list_transactions(
                sku=product.sku,
                transaction_type=TransactionType.OUTBOUND,
                start_time=start_date,
                end_time=end_date,
            )

            total_outbound = sum(tx.quantity for tx in outbound_transactions)

            if not outbound_transactions:
                average_stock = float(product.current_stock)
            else:
                stock_sum = 0
                count = 0
                for tx in outbound_transactions:
                    stock_sum += tx.previous_stock
                    count += 1
                stock_sum += product.current_stock
                count += 1
                average_stock = stock_sum / count

            if average_stock > 0:
                turnover_rate = total_outbound / average_stock
                turnover_days = 30.0 / turnover_rate if turnover_rate > 0 else TURNOVER_DAYS_INFINITY
            else:
                turnover_rate = 0.0
                turnover_days = TURNOVER_DAYS_INFINITY

            item = TurnoverAnalysisItem(
                sku=product.sku,
                name=product.name,
                category=product.category,
                total_outbound=total_outbound,
                average_stock=average_stock,
                turnover_rate=turnover_rate,
                turnover_days=turnover_days,
            )
            items.append(item)

        analysis = InventoryTurnoverAnalysis(
            id=self._generate_id(),
            start_date=start_date,
            end_date=end_date,
            items=items,
        )

        return self._storage.create_turnover_analysis(analysis)

    def get_turnover_analysis(self, analysis_id: str) -> Optional[InventoryTurnoverAnalysis]:
        return self._storage.get_turnover_analysis(analysis_id)

    def get_product_inventory_status(self, sku: str) -> Optional[Tuple[Product, InventoryStatus, SeasonalSafetyStock]]:
        product = self._storage.get_product(sku)
        if product is None:
            return None

        current_month = self._get_current_month()
        ss = product.safety_stock.get_current_safety_stock(current_month)
        status = product.get_inventory_status(current_month)

        return (product, status, ss)
